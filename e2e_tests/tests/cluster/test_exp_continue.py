
import pytest


from determined.common.api.bindings import experimentv1State
from tests import experiment as exp
from tests import config as conf


@pytest.mark.e2e_cpu
def test_continue_fixing_broken_config() -> None:
    exp_id = exp.create_experiment(
        conf.fixtures_path("no_op/single-medium-train-step.yaml"),
        conf.fixtures_path("no_op"),
        ["--config", "hyperparameters.metrics_sigma=-1.0"],
    )
    exp.wait_for_experiment_state(exp_id, experimentv1State.ERROR)
    # RUNNING exp.wait_for_experiment_state(exp_id, experimentv1State.ERROR)
