import argparse
import dataclasses
import glob
import logging
import os
import pathlib
import tempfile
from typing import Any, Callable, Dict, Iterator, List, Optional, Tuple, Type, Union, cast

import omegaconf
import pytorch_lightning as pl
import torch
from pytorch_lightning import utilities
from pytorch_lightning.utilities import deepspeed

import determined as det
from determined import core

CHECKPOINT_DOWNLOAD_PATH = "determined_checkpoint_download"
TEMP_CHECKPOINT_FILE = "determined.ckpt"


def flatten(xs: List[List]) -> List:
    return [item for items in xs for item in items]


def get_cluster_info_with_assert() -> det.ClusterInfo:
    """
    Raise an exception if not run on a Determined cluster.  Returns ClusterInfo.
    """
    info = det.get_cluster_info()
    assert info, "This code can only be run on-cluster."
    return info


def download_checkpoint(core_context: det.core.Context, module_load_only: bool) -> Optional[str]:
    info = det.get_cluster_info()
    if info:
        ckpt_id = info.latest_checkpoint
        if ckpt_id:
            core_context.checkpoint.download(ckpt_id, CHECKPOINT_DOWNLOAD_PATH)
    if os.path.isdir(CHECKPOINT_DOWNLOAD_PATH):
        if "latest" in os.listdir(CHECKPOINT_DOWNLOAD_PATH):
            if module_load_only:
                # DeepSpeed checkpoint; convert to a .ckpt file.
                deepspeed.convert_zero_checkpoint_to_fp32_state_dict(
                    CHECKPOINT_DOWNLOAD_PATH, TEMP_CHECKPOINT_FILE
                )
                return TEMP_CHECKPOINT_FILE
            else:
                return CHECKPOINT_DOWNLOAD_PATH
        else:
            ckpt_files = glob.glob(os.path.join(CHECKPOINT_DOWNLOAD_PATH, "*.ckpt"))
            assert len(ckpt_files) == 1, "Checkpoint must contain exactly one .ckpt file."
            return ckpt_files[0]
    return None


def get_checkpoint_metadata(core_context: det.core.Context) -> Optional[Dict]:
    info = det.get_cluster_info()
    if info:
        ckpt_id = info.latest_checkpoint
        if ckpt_id:
            return cast(Dict, core_context.checkpoint.get_metadata(ckpt_id))
    return None


def get_searcher_metric_name() -> str:
    return cast(str, get_cluster_info_with_assert().trial._config["searcher"]["metric"])


def get_searcher_max_length() -> int:
    max_length_entry = get_cluster_info_with_assert().trial._config["searcher"]["max_length"]
    if isinstance(max_length_entry, dict):
        assert tuple(max_length_entry.keys()) == (
            "epochs",
        ), "Must express max training length in epochs."
        return cast(int, max_length_entry["epochs"])
    else:
        return cast(int, max_length_entry)


@dataclasses.dataclass
class DeterminedIntegrationSharedState:
    """
    State shared between the components of the Determined integration on a single Trainer.
    """

    core_context: det.core.Context
    searcher_ops: Iterator[core.SearcherOperation]
    current_op: core.SearcherOperation
    global_step: int = 0


# Default environment settings in PTL don't work with multi-node DeepSpeed launch, so we
# need to explicitly configure this.
class DeterminedClusterEnvironment(pl.plugins.ClusterEnvironment):  # type: ignore
    def __init__(self, shared: DeterminedIntegrationSharedState):
        self.shared = shared

    @property
    def creates_processes_externally(self) -> bool:
        return True

    @property
    def main_address(self) -> str:
        return os.environ["DET_CHIEF_IP"]

    @property
    def main_port(self) -> int:
        if "USE_DEEPSPEED" in os.environ:
            # Determined uses the default port for DeepSpeed init_distributed:
            # - https://deepspeed.readthedocs.io/en/latest/initialize.html
            return 29500
        else:
            return int(os.environ["MASTER_PORT"])

    @staticmethod
    def detect() -> bool:
        raise Exception("Unimplemented")

    def world_size(self) -> int:
        return self.shared.core_context.distributed.size

    def set_world_size(self, size: int) -> None:
        assert size == self.shared.core_context.distributed.size

    def global_rank(self) -> int:
        return self.shared.core_context.distributed.rank

    def set_global_rank(self, rank: int) -> None:
        assert rank == self.shared.core_context.distributed.rank

    def local_rank(self) -> int:
        return self.shared.core_context.distributed.local_rank

    def node_rank(self) -> int:
        return self.shared.core_context.distributed.cross_rank


class DeterminedLogger(pl.loggers.logger.Logger):
    def __init__(self, shared: DeterminedIntegrationSharedState) -> None:
        self.shared = shared

    def log_hyperparams(
        self, params: Union[Dict[str, Any], argparse.Namespace], *args: Any, **kwargs: Any
    ) -> None:
        pass

    @utilities.rank_zero_only  # type: ignore
    def log_metrics(self, metrics: Dict, step: int) -> None:
        metrics.pop("epoch", None)
        self.shared.core_context.train._report_trial_metrics("logger", step if step else 0, metrics)

    @property
    def name(self) -> Optional[str]:
        pass

    @property
    def version(self) -> Optional[Union[int, str]]:
        pass


def _per_rank_upload(
    path: Union[str, pathlib.Path],
    shared: DeterminedIntegrationSharedState,
    shard: bool = False,
    selector: Optional[Callable[[str], bool]] = None,
) -> None:
    det_checkpoint_metadata = {
        "steps_completed": shared.global_step,
        "trial_id": get_cluster_info_with_assert().trial.trial_id,
    }
    logging.info("Finished checkpointing uploading files to Determined")
    if os.path.isfile(path):
        # Create a temporary directory with a symbolic link to the saved file,
        # so we can upload it without making a copy.
        # If path is a directory terminated with /, basename will return empty string --
        # we use normpath to ensure it returns the last directory.
        ckpt_name = os.path.basename(os.path.normpath(path))
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_ckpt_path = os.path.join(temp_dir, ckpt_name)
            os.symlink(os.path.abspath(path), os.path.abspath(temp_ckpt_path))
            shared.core_context.checkpoint.upload(
                temp_dir, det_checkpoint_metadata, shard=shard, selector=selector
            )
    else:
        shared.core_context.checkpoint.upload(
            path, det_checkpoint_metadata, shard=shard, selector=selector
        )


def upload_determined_checkpoint(
    strategy: pl.strategies.Strategy,
    path: Union[str, pathlib.Path],
    shared: DeterminedIntegrationSharedState,
) -> None:
    sharded = False
    if isinstance(strategy, pl.strategies.FSDPStrategy) or isinstance(
        strategy, pl.strategies.DeepSpeedStrategy
    ):
        sharded = True

        def get_selector(local_rank: int) -> Callable[[str], bool]:
            def always_include(x: str) -> bool:
                return True

            def always_exclude(x: str) -> bool:
                return False

            if local_rank == 0:
                return always_include
            return always_exclude

        _per_rank_upload(
            path, shared, sharded, get_selector(shared.core_context.distributed.local_rank)
        )
    else:
        if shared.core_context.distributed.rank == 0:
            _per_rank_upload(path, shared)


class DeterminedCallback(pl.callbacks.Callback):
    def __init__(self, shared: DeterminedIntegrationSharedState) -> None:
        self.shared = shared
        self.core_context = shared.core_context

    def setup(
        self, trainer: pl.Trainer, pl_module: pl.LightningModule, stage: Optional[str] = None
    ) -> None:
        # If fitting/testing multiple times, keep a monotonically increasing global step for
        # reporting Determined metrics and checkpoints.
        self.shared.global_step += 1
        self.initial_global_step = self.shared.global_step

    def on_train_batch_end(
        self,
        trainer: pl.Trainer,
        pl_module: pl.LightningModule,
        outputs: pl.utilities.types.STEP_OUTPUT,
        batch: Any,
        batch_idx: int,
    ) -> None:
        self.shared.global_step = self.initial_global_step + trainer.global_step

    def _get_and_report_searcher_metric(
        self, trainer: pl.Trainer, pl_module: pl.LightningModule
    ) -> None:
        validation_metrics = trainer.logged_metrics
        searcher_metric_matches = [
            k for k in validation_metrics.keys() if get_searcher_metric_name() in k
        ]
        if not searcher_metric_matches:
            searcher_metric = 0.0
            logging.warning(
                f"Searcher metric {get_searcher_metric_name()} was not " "logged.  Reporting as 0.",
            )
        else:
            searcher_metric_epoch = [k for k in searcher_metric_matches if "epoch" in k]
            if searcher_metric_epoch:
                searcher_metric = validation_metrics[searcher_metric_epoch[0]]  # type: ignore
            else:
                searcher_metric = validation_metrics[searcher_metric_matches[0]]  # type: ignore
            if isinstance(searcher_metric, torch.Tensor):
                searcher_metric = searcher_metric.detach().cpu().numpy()

        if not trainer.sanity_checking:
            if self.core_context.distributed.rank == 0:
                self.shared.current_op.report_progress(trainer.current_epoch + 1)
            if self.shared.core_context.distributed.rank == 0:
                self.shared.core_context.train.report_validation_metrics(
                    self.shared.global_step, {get_searcher_metric_name(): searcher_metric}
                )
            if (trainer.current_epoch + 1) >= self.shared.current_op.length:
                if self.core_context.distributed.rank == 0:
                    self.shared.current_op.report_completed(searcher_metric)
                try:
                    self.shared.current_op = next(self.shared.searcher_ops)
                except StopIteration:
                    logging.info("Reached end of searcher operations.")
                    trainer.should_stop = True

    def on_validation_epoch_end(self, trainer: pl.Trainer, pl_module: pl.LightningModule) -> None:
        self._get_and_report_searcher_metric(trainer, pl_module)
        if self.core_context.preempt.should_preempt():
            logging.info("Preemption signal received, exiting.")
            trainer.should_stop = True

    def on_train_epoch_end(self, trainer: pl.Trainer, pl_module: pl.LightningModule) -> None:
        if not trainer.enable_validation:
            logging.warning(
                "validation_step not defined for LightningModule, using training metrics to report searcher metric instead."
            )
            self._get_and_report_searcher_metric(trainer, pl_module)
        if self.core_context.preempt.should_preempt():
            logging.info("Preemption signal received, exiting.")
            trainer.should_stop = True


def get_hyperparameters() -> omegaconf.DictConfig:
    """
    Returns Determined trial hyperparameters as an OmegaConf.
    """
    info = det.get_cluster_info()
    assert info is not None, "This example only runs on-cluster"
    return omegaconf.OmegaConf.create(info.trial.hparams)


def determined_core_init() -> det.core.Context:
    """
    Checks for DeepSpeed and initializes a det.core.Context appropriately.
    """
    if "USE_DEEPSPEED" in os.environ:
        distributed_context = det.core.DistributedContext.from_deepspeed()
    else:
        distributed_context = det.core.DistributedContext.from_torch_distributed()
    return det.core.init(distributed=distributed_context)


class DeterminedTrainer(pl.Trainer):
    def __init__(self, shared: DeterminedIntegrationSharedState, **kwargs) -> None:  # type: ignore
        super().__init__(**kwargs)
        self.shared = shared

    def save_checkpoint(
        self,
        filepath: Union[str, pathlib.Path],
        weights_only: bool = False,
        storage_options: Optional[Any] = None,
    ) -> None:
        super().save_checkpoint(filepath, weights_only, storage_options)
        upload_determined_checkpoint(self.strategy, filepath, self.shared)


def _add_integration_controlled_args(kwargs: Dict[str, Any], intargs: Dict[str, Any]) -> None:
    """
    Adds arguments to kwargs after asserting they're not present.
    """
    for k in intargs:
        assert (
            k not in kwargs
        ), f"`{k}` is supplied by build_determined_trainer, so can not be as an argument."
        kwargs[k] = intargs[k]


def _append_integration_controlled_args(kwargs: Dict[str, Any], intargs: Dict[str, Any]) -> None:
    """
    Appends the value in intargs to the associated list in kwargs, creating if necessary.
    """
    for k in intargs:
        val = kwargs.get(k, [])
        if not (isinstance(val, list)):
            val = [val]
        val.append(intargs[k])
        kwargs[k] = val


def _configure_deepspeed(kwargs: Dict[str, Any], shared: DeterminedIntegrationSharedState) -> None:
    if "strategy" in kwargs:
        strategy = kwargs["strategy"]
        if strategy is str and "deepspeed" in strategy:
            strategy = pl.strategies.StrategyRegistry.get(strategy)
        if isinstance(strategy, pl.strategies.DeepSpeedStrategy):
            strategy.cluster_environment = DeterminedClusterEnvironment(shared)
        kwargs["strategy"] = strategy


def _replace_tensorboard_path(
    kwargs: Dict[str, Any], shared: DeterminedIntegrationSharedState
) -> None:
    found_tb_logger = False
    if "logger" in kwargs:
        if not isinstance(kwargs["logger"], list):
            kwargs["logger"] = [kwargs["logger"]]
        for logger in kwargs["logger"]:
            if isinstance(logger, pl.loggers.TensorBoardLogger):
                logger._root_dir = shared.core_context.train.get_tensorboard_path()
    else:
        kwargs["logger"] = []
    if not found_tb_logger:
        tb = pl.loggers.TensorBoardLogger(shared.core_context.train.get_tensorboard_path())
        kwargs["logger"].append(tb)


def build_determined_trainer(
    core_context: det.core.Context,
    module_load_only: bool = False,
    **kwargs: Any,
) -> Tuple[DeterminedTrainer, Optional[str]]:
    """
    Returns a tuple of (Trainer, LightningModule) configured to run under Determined.
    The trainer and module state will be loaded from checkpoint if resumed from a pause.
    The module state will be loaded from checkpoint if this is a new trial with
    a checkpoint supplied (e.g. Continue Trial in the Web UI).

    Accepts the usual parameters to Trainer(...), with the following exceptions controlled
    by the Determined trial configuration:
    - num_nodes
    - devices
    - accelerator
    - resume_from_checkpoint
    - max_epochs
    """
    searcher_ops = core_context.searcher.operations()
    shared = DeterminedIntegrationSharedState(
        core_context=core_context,
        searcher_ops=searcher_ops,
        current_op=next(searcher_ops),
    )
    _configure_deepspeed(kwargs, shared)
    ckpt_metadata = get_checkpoint_metadata(core_context)
    if ckpt_metadata and ckpt_metadata["trial_id"] != get_cluster_info_with_assert().trial.trial_id:
        # New trial, so experiment hyperparameters may have changed.  Instead of fully loading
        # the training checkpoint, we just load the module.
        logging.info("New trial -- only loading module weights and not training state.")
        module_load_only = True
    ckpt_path = download_checkpoint(core_context, module_load_only)
    _replace_tensorboard_path(kwargs, shared)
    _append_integration_controlled_args(
        kwargs,
        {
            "callbacks": DeterminedCallback(shared),
            "logger": DeterminedLogger(shared),
        },
    )
    _add_integration_controlled_args(
        kwargs,
        {
            "num_nodes": core_context.distributed.cross_size,
            "devices": "auto",
            "accelerator": "gpu",
            "max_epochs": get_searcher_max_length(),
        },
    )
    return DeterminedTrainer(shared, **kwargs), ckpt_path
