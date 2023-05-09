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


def _init_device() -> str:
    device = "cuda" if torch.cuda.is_available() else "cpu"
    return device


def _calculate_progress(
    skip: int, local_batch_idx: int, batch_size: int, num_replicas: int, dataset_len: int
):
    records_processed = (skip + local_batch_idx) * num_replicas * batch_size
    return records_processed / dataset_len


def _try_compile_model(model: TorchModel) -> TorchModel:
    if isinstance(model, torch.jit.ScriptModule):
        return model
    if isinstance(model, torch.nn.Module):
        try:
            return torch.jit.script(model)
        except Exception:
            logging.info("Model is not scriptable, falling back to using eager mode model.")
        return model


def _prepare_model(model: TorchModel, device: str) -> TorchModel:
    model = _try_compile_model(model)
    model.eval()
    model.to(device)
    return model


def _report_progress_to_master(
    searcher_op: core.DummySearcherOperation, rank: int, completion_rate: float
) -> None:
    if rank == 0:
        searcher_op.report_progress(completion_rate)


def _load_state(checkpoint_directory):
    checkpoint_directory = pathlib.Path(checkpoint_directory)
    with checkpoint_directory.joinpath("metadata.json").open("r") as f:
        metadata = json.load(f)
        return metadata


class TorchOfflineBatchPredictor:
    def __init__(
        self,
        core_context,
        model: TorchModel,
        output_directory: str,
        batch_size: int = 64,
        dataloader_num_workers: int = 2,
        data_persist_interval=10,
    ):
        self._core_context = core_context
        self._model = model
        self._output_directory = output_directory
        self._batch_size = batch_size
        self._dataloader_num_workers = dataloader_num_workers
        self._metrics_reducer = None
        self._torch_profiler = None
        self._data_persist_interval = data_persist_interval

        info = det.get_cluster_info()
        self._total_worker = len(info.container_addrs)
        # [SWY]: what is the difference between container_rank vs rank from core context
        # can one container has multiple GPU?
        self._rank = info.container_rank
        latest_checkpoint = info.latest_checkpoint
        self._skip = 0

        if latest_checkpoint is not None:
            logging.info("Checkpoint is not none")
            with self.core_context.checkpoint.restore_path(latest_checkpoint) as path:
                metadata = self._load_state(path)
                self._skip = metadata["steps_completed"]
                logging.info(f"Previous run completed {self._skip} steps")

    def set_metrics_reducer(self, metrics_reducer):
        self._metrics_reducer = metrics_reducer

    def _create_dataloader(self, dataset: Dataset) -> DataLoader:
        if isinstance(dataset, torch.utils.data.IterableDataset):
            raise Exception("Only map style dataset with __getitem__ method is supported.")
        sampler = SequentialSampler(dataset)
        batch_sampler = BatchSampler(sampler, self._batch_size, drop_last=False)
        # Adapt batch_sampler for distributed inference and trial resumption if applicable
        batch_sampler = adapt_batch_sampler(
            batch_sampler,
            repeat=False,
            skip=self._skip,
            num_replicas=self._total_worker,
            rank=self._rank,
        )

        return torch.utils.data.DataLoader(
            dataset, batch_sampler=batch_sampler, num_workers=self._dataloader_num_workers
        )

    def select_torch_tensor_from_input(self, item: Any) -> torch.Tensor:
        # find torch.Tensor from dataset item
        if isinstance(item, torch.Tensor):
            return item

        try:
            item_iterator = iter(item)
        except TypeError:
            raise Exception(
                "Dataset item does not contain torch.Tensor for inference. "
                + "Please check __getitem__ method of your Torch dataset"
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

    def post_process_predictions(
        self, output: Any, filename: str, dir_path: str, metric_reducer=None
    ) -> None:
        file_path = pathlib.PosixPath(dir_path, filename)
        torch.save(output, file_path)

    def _predict_batch(self, batch: Any, output_buffer: List[Any], device: str) -> None:
        torch_tensor_input = self.select_torch_tensor_from_input(batch)
        torch_tensor_input = torch_tensor_input.to(device=device)
        with torch.no_grad():
            if self._torch_profiler is not None:
                with self._torch_profiler as profiler:
                    pred = self._model(torch_tensor_input)
                    profiler.step()
            else:
                pred = self._model(torch_tensor_input)
            pred = pred.to(device="cpu")
            output_buffer.append({"input": batch, "prediction": pred})

    def _get_tensorboard_path(self) -> pathlib.Path:
        """
        Get the path where files for consumption by TensorBoard should be written
        """
        return self._core_context.train.get_tensorboard_path()

    def set_torch_profiler(self, *args: List[str], **kwargs: Any) -> None:
        self._torch_profiler = torch.profiler.profile(
            on_trace_ready=torch.profiler.tensorboard_trace_handler(
                str(self._get_tensorboard_path())
            ),
            *args,
            **kwargs,
        )

    def _synchronize_and_checkpoint(self, batch: int):
        # After each batch, synchronize and update number of catches completed
        if self._rank == 0:
            self._core_context.distributed.gather(batch)
            checkpoint_metadata = {
                "steps_completed": batch + 1,
            }
            with self._core_context.checkpoint.store_path(checkpoint_metadata) as (path, uuid):
                with open(os.path.join(path, "batch_completed.json"), "w") as file_obj:
                    json.dump({"batch_completed": batch}, file_obj)
        else:
            self._core_context.distributed.gather(batch)

    def _report_progress_to_master(
        self, searcher_op: core.DummySearcherOperation, rank: int, completion_rate: float
    ) -> None:
        if rank == 0:
            searcher_op.report_progress(completion_rate)

    def _calculate_progress(
        self, skip: int, local_batch_idx: int, batch_size: int, num_replicas: int, dataset_len: int
    ):
        records_processed = (skip + local_batch_idx) * num_replicas * batch_size
        return records_processed / dataset_len

    def predict(self, dataset: Dataset):
        device = _init_device()
        self._model = _prepare_model(self._model, device)
        dataloader = self._create_dataloader(dataset)

        dummy_searcher_op = None
        # Initialize dummy searcher for progress report
        if self._rank == 0:
            dummy_searcher_op = core.DummySearcherOperation(1, True)

        dataset_length = len(dataset)

        # Initialize temp variables used in prediction
        output_buffer = []
        last_output_file_id = 0
        local_batch_idx = 0

        for local_batch_idx, X in enumerate(dataloader):
            logging.info(f"Currently processing batch {local_batch_idx + self._skip}")
            self._predict_batch(X, output_buffer, device)

            if (local_batch_idx + 1) % self._data_persist_interval == 0:
                output_file_id = int(
                    (local_batch_idx + 1 + self._skip) / self._data_persist_interval
                )
                last_output_file_id = output_file_id
                filename = f"prediction_output_{output_file_id}_worker_{self._rank}"
                self.post_process_predictions(
                    output_buffer, filename, self._output_directory, self._metrics_reducer
                )
                output_buffer = []

                self._synchronize_and_checkpoint(local_batch_idx)
                progress = self._calculate_progress(
                    self._skip,
                    local_batch_idx,
                    self._batch_size,
                    self._total_worker,
                    dataset_length,
                )
                self._report_progress_to_master(dummy_searcher_op, self._rank, progress)

                if self._core_context.preempt.should_preempt():
                    return

        # Process any remaining prediction output
        if len(output_buffer) > 0:
            filename = f"prediction_output_{last_output_file_id + 1}_worker_{self._rank}"
            self.post_process_predictions(
                output_buffer, filename, self._output_directory, self._metrics_reducer
            )

        # TODO: Update to use unified metrics phase 2 API when ready
        if self._metrics_reducer is not None:
            metrics = self._core_context.distributed.gather(self._metrics_reducer.per_slot_reduce())
            if self._rank == 0:
                output = self._metrics_reducer.cross_slot_reduce(metrics)
                self._core_context.train.report_validation_metrics(
                    steps_completed=local_batch_idx + 1 + self._skip, metrics=output
                )

        self._report_progress_to_master(dummy_searcher_op, self._rank, 1)


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
