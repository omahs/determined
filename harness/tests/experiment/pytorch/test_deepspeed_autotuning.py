import pathlib
import tempfile
import pytest

from argparse import Namespace
from collections import deque
from typing import Any, Deque, Dict, List, Optional, Sequence, Union
from unittest import mock

from determined import searcher
from determined.common.api import bindings
from determined.pytorch.deepspeed.dsat import (
    _defaults,
    _utils,
    autotune,
)
from determined.pytorch.deepspeed.dsat._dsat_search_method import (
    BaseDSATSearchMethod,
    RandomDSATSearchMethod,
)
from tests.custom_search_mocks import MockMasterSearchRunner

ERROR_METRIC_NAME = "error"

BASE_EXPERIMENT_FIXTURE_PATH = (
    pathlib.Path(__file__).resolve().parent.parent.joinpath("fixtures/deepspeed_autotune")
)

MODEL_INFO_PROFILE_METRIC_FIXTURE = {
    "num_params": 60192808,
    "trainable_num_params": 60192808,
    "activation_mem_per_gpu": 89828352,
    "rank": 0,
    "gpu_mem": 15843721216,
}


@mock.patch("determined.experimental.client.create_experiment")
def test_autotuning_module_transforms(
    create_experiment_mock: mock.MagicMock,
) -> None:
    model_dir = BASE_EXPERIMENT_FIXTURE_PATH.joinpath("example_experiment")
    config_path = model_dir.joinpath("autotune_config.yaml")
    args = Namespace(
        config_path=config_path,
        model_dir=model_dir,
        tuner_type="random",
        search_runner_config=None,
    )
    autotune.run_autotuning(args)
    create_experiment_mock.assert_called_once()
    submitted_single_searcher_config_dict = create_experiment_mock.call_args_list[0].kwargs[
        "config"
    ]
    # The Search Runner should be defaulted to a "single" searcher
    assert submitted_single_searcher_config_dict.get("searcher", {}).get("name", "") == "single"


def _run_searcher(search_method: BaseDSATSearchMethod, all_metrics):
    """
    Run a mocked version of the Determined master with a deterministic series of
    returned metrics for a given Deepspeed Autotune Custom Search Method
    """
    model_dir = BASE_EXPERIMENT_FIXTURE_PATH.joinpath("example_experiment")
    config_path = model_dir.joinpath("autotune_config.yaml")
    submitted_config_dict = _utils.get_dict_from_yaml_or_json_path(config_path)

    with tempfile.TemporaryDirectory() as searcher_dir:
        searcher_dir = pathlib.Path(searcher_dir)
        search_method = search_method(
            submitted_config_dict=submitted_config_dict,
            model_dir=model_dir,
        )
        mock_master_obj = MockMaster(all_metrics=all_metrics)
        search_runner = MockMasterSearchRunner(search_method, mock_master_obj, searcher_dir)
        search_runner.run(exp_config={}, context_dir="", includes=None)
    return search_runner


@pytest.mark.timeout(5)
def test_deepspeed_autotune_happy_path() -> None:
    """
    Simulate the Deepspeed Autotune Search Methods end to end and make sure
    nothing falls over
    """
    for search_method in _defaults.ALL_SEARCH_METHOD_CLASSES.values():
        search_runner = _run_searcher(
            search_method, [MODEL_INFO_PROFILE_METRIC_FIXTURE, {"throughput": 1.0}]
        )
        exp_num_trials = _defaults.AUTOTUNING_DICT["tuner_num_trials"]
        assert len(search_runner.state.trials_created) == exp_num_trials
        assert len(search_runner.state.trials_closed) == exp_num_trials
        assert len(search_runner.state.trial_progress) == exp_num_trials
        for trial_uuid in search_runner.state.trial_progress:
            assert search_runner.state.trial_progress[trial_uuid] == 1.0


@pytest.mark.timeout(5)
def test_continuous_failures() -> None:
    """
    Make sure that DSAT Search Methods can handle continuous failures.
    Note that the `ERROR_METRIC_NAME` triggered `v1TrialExitedEarly` event
    will happen for all trials after the first model profile info and single
    successful run
    """
    for search_method in _defaults.ALL_SEARCH_METHOD_CLASSES.values():
        all_mock_metrics = [
            MODEL_INFO_PROFILE_METRIC_FIXTURE,
            {"throughput": 1.0},
            {ERROR_METRIC_NAME: 1.0},
        ]
        search_runner = _run_searcher(search_method, all_mock_metrics)
        exp_num_trials = _defaults.AUTOTUNING_DICT["tuner_num_trials"]
        assert len(search_runner.state.trials_created) == exp_num_trials
        assert len(search_runner.state.failures) == exp_num_trials - 2
        assert len(search_runner.state.trials_closed) == exp_num_trials
        assert len(search_runner.state.trial_progress) == exp_num_trials


@pytest.mark.timeout(5)
def test_one_off_failure() -> None:
    """Make sure that DSAT Search Methods can properly handle a single failure"""
    for search_method in _defaults.ALL_SEARCH_METHOD_CLASSES.values():
        all_mock_metrics = [
            MODEL_INFO_PROFILE_METRIC_FIXTURE,
            {ERROR_METRIC_NAME: 1.0},
            {"throughput": 1.0},
        ]
        search_runner = _run_searcher(search_method, all_mock_metrics)
        exp_num_trials = _defaults.AUTOTUNING_DICT["tuner_num_trials"]
        assert len(search_runner.state.trials_created) == exp_num_trials
        assert len(search_runner.state.failures) == 1
        assert len(search_runner.state.trials_closed) == exp_num_trials
        assert len(search_runner.state.trial_progress) == exp_num_trials


@pytest.mark.timeout(5)
def test_simple_model_profile_info_run_fails() -> None:
    """Run the random search method where the model profile info run fails"""
    search_runner = _run_searcher(
        RandomDSATSearchMethod,
        [
            {ERROR_METRIC_NAME: 1.0},
        ],
    )
    assert len(search_runner.state.trials_created) == 1
    assert len(search_runner.state.failures) == 1
    assert len(search_runner.state.trials_closed) == 1
    assert len(search_runner.state.trial_progress) == 1


class MockMaster:
    """
    Sends v1 metrics back to the Search Runner in the manner defined with the
    `all_metrics` list of dictionaries.

    The metrics are sent as a `v1ValidationCompleted` metric event. When the key for
    the metric is instead `ERROR_METRIC_NAME`, this signals to the `MockMaster` to
    instead send a `v1TrialExitedEarly` event to the Search Runner.

    The last element of the `all_metrics` list will be repeated until the Search Runner
    quits.
    """

    def __init__(self, all_metrics: List[Union[float, Dict[str, Any]]]) -> None:
        self.events_queue: Deque[bindings.v1SearcherEvent] = deque([])
        self.events_count = 0
        self.all_metrics = all_metrics
        self.metric_index = 0

    def handle_post_operations(
        self, event: bindings.v1SearcherEvent, operations: List[searcher.Operation]
    ) -> None:
        self._remove_upto(event)
        self._process_operations(operations)

    def _remove_upto(self, event: bindings.v1SearcherEvent) -> None:
        while len(self.events_queue) > 0:
            e = self.events_queue.popleft()
            if e.id == event.id:
                return

        raise RuntimeError(f"event not found in events queue: {event}")

    def _process_operations(self, operations: List[searcher.Operation]) -> None:
        for op in operations:
            self._append_events_for_op(op)  # validate_after returns two events.

    def handle_get_events(self) -> Optional[Sequence[bindings.v1SearcherEvent]]:
        return list(self.events_queue)

    def _append_events_for_op(self, op: searcher.Operation) -> None:
        if isinstance(op, searcher.ValidateAfter):
            index = min(self.metric_index, len(self.all_metrics) - 1)
            metric = self.all_metrics[index]
            self.metric_index += 1
            if ERROR_METRIC_NAME in metric:
                trial_exited_early = bindings.v1TrialExitedEarly(
                    requestId=str(op.request_id),
                    exitedReason=bindings.v1TrialExitedEarlyExitedReason.EXITED_REASON_UNSPECIFIED,
                )
                self.events_count += 1
                event = bindings.v1SearcherEvent(
                    id=self.events_count, trialExitedEarly=trial_exited_early
                )
                self.events_queue.append(event)

                trial_closed = bindings.v1TrialClosed(requestId=str(op.request_id))
                self.events_count += 1
                event = bindings.v1SearcherEvent(id=self.events_count, trialClosed=trial_closed)
                self.events_queue.append(event)
            else:
                validation_completed = bindings.v1ValidationCompleted(
                    requestId=str(op.request_id),
                    metric=metric,
                    validateAfterLength=str(op.length),
                )

                self.events_count += 1
                event = bindings.v1SearcherEvent(
                    id=self.events_count, validationCompleted=validation_completed
                )
                self.events_queue.append(event)

                # Send 1.0 to signal it was completed
                trial_progress = bindings.v1TrialProgress(
                    requestId=str(op.request_id), partialUnits=1.0
                )
                self.events_count += 1
                event = bindings.v1SearcherEvent(id=self.events_count, trialProgress=trial_progress)
                self.events_queue.append(event)

        elif isinstance(op, searcher.Create):
            trial_created = bindings.v1TrialCreated(requestId=str(op.request_id))
            self.events_count += 1
            event = bindings.v1SearcherEvent(id=self.events_count, trialCreated=trial_created)
            self.events_queue.append(event)

        elif isinstance(op, searcher.Progress):  # no events
            pass

        elif isinstance(op, searcher.Close):
            trial_closed = bindings.v1TrialClosed(requestId=str(op.request_id))
            self.events_count += 1
            event = bindings.v1SearcherEvent(id=self.events_count, trialClosed=trial_closed)
            self.events_queue.append(event)

        elif isinstance(op, searcher.Shutdown):
            exp_state = bindings.experimentv1State.STATE_COMPLETED
            exp_inactive = bindings.v1ExperimentInactive(experimentState=exp_state)
            self.events_count += 1
            event = bindings.v1SearcherEvent(id=self.events_count, experimentInactive=exp_inactive)
            self.events_queue.append(event)
