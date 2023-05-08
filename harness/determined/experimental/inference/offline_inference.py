import determined as det
import logging
import json
import os
import pathlib
import torch
import torch.distributed as dist

from determined import core, pytorch
from determined.common import set_logger
from determined.pytorch import adapt_batch_sampler
from torch.utils.data import BatchSampler, Dataset, DataLoader, SequentialSampler
from typing import Any, Callable, List, Optional, Union

TorchModel = Union[torch.nn.Module, torch.jit.ScriptModule]
PredictionPostprocessFn = Callable[[Any, str, str, Optional[pytorch.Reducer]], None]
ModelInputTensorSelector = Callable[[Any], torch.Tensor]

set_logger(False)


def _prepare_model(model: TorchModel, device: str) -> TorchModel:
    model = _try_compile_model(model)
    model.eval()
    model.to(device)
    return model


def _try_compile_model(model: TorchModel) -> TorchModel:
    if isinstance(model, torch.jit.ScriptModule):
        return model
    if isinstance(model, torch.nn.Module):
        try:
            return torch.jit.script(model)
        except Exception:
            logging.info("Model is not scriptable, falling back to using eager mode model.")

    return model


def _create_dataloader(
    dataset: Dataset,
    batch_size: int,
    rank: int,
    num_replicas: int,
    skip: int = 0,
    dataloader_num_workers: int = 2,
) -> DataLoader:
    if isinstance(dataset, torch.utils.data.IterableDataset):
        raise Exception("Only map style dataset with __getitem__ method is supported.")
    sampler = SequentialSampler(dataset)
    batch_sampler = BatchSampler(sampler, batch_size, drop_last=False)
    # Adapt batch_sampler for distributed inference and trial resumption if applicable
    batch_sampler = adapt_batch_sampler(
        batch_sampler, repeat=False, skip=skip, num_replicas=num_replicas, rank=rank
    )

    return torch.utils.data.DataLoader(
        dataset,
        batch_sampler=batch_sampler,
        num_workers=dataloader_num_workers,
    )


def _predict_batch(
    batch: Any, model: TorchModel, field_selector: Callable, output_buffer: List[Any], device: str
):
    torch_tensor_input = field_selector(batch)
    if device != "cpu":
        torch_tensor_input = torch_tensor_input.to(device=device)
    with torch.no_grad():
        pred = model(torch_tensor_input)
        if device != "cpu":
            pred = pred.to(device="cpu")
        output_buffer.append({"input": batch, "prediction": pred})


def _default_torch_tensor_selector(item: Any) -> torch.Tensor:
    # find torch.Tensor from dataset item
    if isinstance(item, torch.Tensor):
        return item

    item_iterator = None
    try:
        item_iterator = iter(item)
    except TypeError:
        raise Exception(
            "Dataset item does not contain torch.Tensor for inference. "
            + "Please check __get sitem__ method of your Torch dataset"
        )

    torch_tensor = None

    for field in item_iterator:
        if isinstance(field, torch.Tensor):
            if torch_tensor is None:
                torch_tensor = field
            else:
                # Torch tensor field found already
                raise Exception(
                    "Dataset __getitem__ method returns multiple torch.Tensor,"
                    + "please pass in custom selector or modify dataset to only "
                    + "return one torch.Tensor"
                )

    return torch_tensor


def _default_persist_predictions(output: Any, filename: str, dir_path: pathlib.Path) -> None:
    file_path = pathlib.PosixPath(dir_path, filename)
    torch.save(output, file_path)


def _synchronize_and_checkpoint(rank: int, batch: int, core_context):
    # After each batch, synchronize and update number of catches completed
    if rank == 0:
        core_context.distributed.gather(batch)
        checkpoint_metadata = {
            "steps_completed": batch + 1,
        }
        with core_context.checkpoint.store_path(checkpoint_metadata) as (path, uuid):
            with open(os.path.join(path, "batch_completed.json"), "w") as file_obj:
                json.dump({"batch_completed": batch}, file_obj)
    else:
        core_context.distributed.gather(batch)


def _load_state(checkpoint_directory):
    checkpoint_directory = pathlib.Path(checkpoint_directory)
    with checkpoint_directory.joinpath("metadata.json").open("r") as f:
        metadata = json.load(f)
        return metadata


def _init_device() -> str:
    device = "cuda" if torch.cuda.is_available() else "cpu"
    return device


def _report_progress_to_master(
    searcher_op: core.DummySearcherOperation, rank: int, completion_rate: float
) -> None:
    if rank == 0:
        searcher_op.report_progress(completion_rate)


def _calculate_progress(
    skip: int, local_batch_idx: int, batch_size: int, num_replicas: int, dataset_len: int
):
    records_processed = (skip + local_batch_idx) * num_replicas * batch_size
    return records_processed / dataset_len


def torch_run_batch_inference(
    core_context,
    model: TorchModel,
    dataset: Dataset,
    output_directory: str,
    model_input_tensor_selector: Optional[ModelInputTensorSelector] = None,
    post_process_and_save_predictions: Optional[PredictionPostprocessFn] = None,
    batch_size: int = 64,
    dataloader_num_workers: int = 2,
    metrics_reducer: pytorch.Reducer = None,
):
    data_persist_interval = 10
    output_directory = pathlib.Path(output_directory)

    # Get job information
    info = det.get_cluster_info()
    total_worker = len(info.container_addrs)
    # [SWY]: what is the difference between container_rank vs rank from core context
    # can one container has multiple GPU?
    rank = info.container_rank
    latest_checkpoint = info.latest_checkpoint
    skip = 0

    device = _init_device()

    dummy_searcher_op = None
    if rank == 0:
        dummy_searcher_op = core.DummySearcherOperation(1, True)

    if latest_checkpoint is not None:
        logging.info("Checkpoint is not none")
        with core_context.checkpoint.restore_path(latest_checkpoint) as path:
            metadata = _load_state(path)
            skip = metadata["steps_completed"]
            logging.info(f"Previous run completed {skip} steps")

    model = _prepare_model(model, device)
    dataloader = _create_dataloader(
        dataset, batch_size, rank, total_worker, skip, dataloader_num_workers
    )

    dataset_length = len(dataset)

    if not model_input_tensor_selector:
        model_input_tensor_selector = _default_torch_tensor_selector
    if not post_process_and_save_predictions:
        post_process_and_save_predictions = _default_persist_predictions

    output_buffer = []

    last_output_file_id = 0

    local_batch_idx = 0

    for local_batch_idx, X in enumerate(dataloader):
        logging.info(f"Currently processing batch {local_batch_idx + skip}")
        _predict_batch(X, model, model_input_tensor_selector, output_buffer, device)

        if (local_batch_idx + 1) % data_persist_interval == 0:
            output_file_id = int((local_batch_idx + 1 + skip) / data_persist_interval)
            last_output_file_id = output_file_id
            filename = f"prediction_output_{output_file_id}_worker_{rank}"
            post_process_and_save_predictions(
                output_buffer, filename, output_directory, metrics_reducer
            )
            output_buffer = []

            _synchronize_and_checkpoint(rank, local_batch_idx, core_context)
            progress = _calculate_progress(
                skip, local_batch_idx, batch_size, total_worker, dataset_length
            )
            _report_progress_to_master(dummy_searcher_op, rank, progress)

            if core_context.preempt.should_preempt():
                return

    # Process any remaining prediction output
    filename = f"prediction_output_{last_output_file_id + 1}_worker_{rank}"
    post_process_and_save_predictions(output_buffer, filename, output_directory, metrics_reducer)

    # [TODO]: Update to use unified metrics phase 2 API when ready
    if metrics_reducer is not None:
        metrics = core_context.distributed.gather(metrics_reducer.per_slot_reduce())
        if rank == 0:
            output = metrics_reducer.cross_slot_reduce(metrics)
            core_context.train.report_validation_metrics(
                steps_completed=local_batch_idx + 1 + skip, metrics=output
            )

    _report_progress_to_master(dummy_searcher_op, rank, 1)


def initialize_distributed_backend() -> Optional[core.DistributedContext]:
    # Pytorch specific initialization
    if torch.cuda.is_available():
        dist.init_process_group(
            backend="nccl",
        )  # type: ignore
        return core.DistributedContext.from_torch_distributed()
    else:
        dist.init_process_group(backend="gloo")  # type: ignore
    return core.DistributedContext.from_torch_distributed()
