import argparse
import logging
import os
import pathlib
import pickle

import determined as det
from determined import searcher
from determined.pytorch.dsat import _defaults, _dsat_search_method, _utils
from determined.util import merge_dicts


def build_exp_conf_from_args(args: argparse.Namespace) -> _dsat_search_method.BaseDSATSearchMethod:
    """
    Helper function which alters the user-submitted configuration and args into a configuration
    for the DS AT custom searchers.
    """
    exp_config = _utils.get_dict_from_yaml_or_json_path(
        args.config_path
    )  # add the search runner's experiment id to the description of the corresonding Trial
    additional_description = f"(#{args.experiment_id}) generated"
    existing_description = exp_config.get("description")
    if existing_description is not None:
        exp_config["description"] = f"{additional_description} - {exp_config['description']}"
    else:
        exp_config["description"] = additional_description

    # Overwrite the searcher section.
    exp_config["searcher"] = {
        "name": "custom",
        "metric": args.metric,
        "smaller_is_better": _utils.smaller_is_better(args.metric),
    }
    # Add all necessary autotuning keys from defaults and user-supplied args.
    autotuning_config = _defaults.AUTOTUNING_DICT
    autotuning_config["start_profile_step"] = args.start_profile_step
    autotuning_config["end_profile_step"] = args.end_profile_step

    exp_config["hyperparameters"] = merge_dicts(
        exp_config["hyperparameters"], {_defaults.OVERWRITE_KEY: autotuning_config}
    )
    # Add an internal key to the HP dict which enables the DSAT code path for Trial classes.
    exp_config["hyperparameters"][_defaults.USE_DSAT_MODE_KEY] = True

    return exp_config


def main(core_context: det.core.Context) -> None:
    with pathlib.Path(_defaults.ARGS_PKL_PATH).open("rb") as f:
        args = pickle.load(f)
    # On-cluster, the relative paths to the below files just come from the base names.
    args.config_path = os.path.basename(args.config_path)
    args.model_dir = os.path.basename(args.model_dir)
    args.include = [os.path.basename(p) for p in args.include] if args.include is not None else []
    args.experiment_id = det.get_cluster_info()._trial_info.experiment_id

    exp_config = build_exp_conf_from_args(args)

    assert (
        args.tuner_type in _defaults.ALL_SEARCH_METHOD_CLASSES
    ), f"tuner-type must be one of {list(_defaults.ALL_SEARCH_METHOD_CLASSES)}, not {args.tuner_type}"
    search_method_class = _defaults.ALL_SEARCH_METHOD_CLASSES[args.tuner_type]
    search_method = search_method_class(args=args, exp_config=exp_config)

    search_runner = searcher.RemoteSearchRunner(search_method, context=core_context)

    search_runner.run(exp_config=exp_config, model_dir=args.model_dir, includes=args.include)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO, format=det.LOG_FORMAT)
    with det.core.init() as core_context:
        main(core_context)
