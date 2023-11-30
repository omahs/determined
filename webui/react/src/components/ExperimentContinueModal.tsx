import { InputNumber } from 'antd';
import Button from 'hew/Button';
import Form, { hasErrors } from 'hew/Form';
import Input from 'hew/Input';
import Message from 'hew/Message';
import { Modal } from 'hew/Modal';
import Spinner from 'hew/Spinner';
import { Body } from 'hew/Typography';
import { Loaded } from 'hew/utils/loadable';
import yaml from 'js-yaml';
import _ from 'lodash';
import React, { useCallback, useEffect, useId, useState } from 'react';

import { paths } from 'routes/utils';
import { continueExperiment, createExperiment } from 'services/api';
import { V1LaunchWarning } from 'services/api-ts-sdk';
import { ExperimentBase, RawJson, RunState, TrialHyperparameters, TrialItem, ValueOf } from 'types';
import handleError, {
  DetError,
  ErrorLevel,
  ErrorType,
  handleWarning,
  isDetError,
  isError,
} from 'utils/error';
import { trialHParamsToExperimentHParams, upgradeConfig } from 'utils/experiment';
import { routeToReactUrl } from 'utils/routes';

export const FULL_CONFIG_BUTTON_TEXT = 'Show Full Config';
export const SIMPLE_CONFIG_BUTTON_TEXT = 'Show Simple Config';
const FORM_ID = 'create-experiment-form';

export const CreateExperimentType = {
  ContinueTrial: 'Continue Experiment',
  ReactivateExperiment: 'Reactivate Experiment',
} as const;

export type CreateExperimentType = ValueOf<typeof CreateExperimentType>;

const EXPERIMENT_NAME = 'name';
const MAX_LENGTH = 'maxLength';

interface Props {
  onClose?: () => void;
  experiment: ExperimentBase;
  trial?: TrialItem;
  type: CreateExperimentType;
}

interface ModalState {
  config: RawJson;
  configError?: string;
  configString: string;
  error?: string;
  experiment?: ExperimentBase;
  isAdvancedMode: boolean;
  open: boolean;
  trial?: TrialItem;
  type: CreateExperimentType;
}

const getExperimentName = (config: RawJson) => {
  return config.name || '';
};

// For unitless searchers, this will return undefined.
const getMaxLengthType = (config: RawJson) => {
  return (Object.keys(config.searcher?.max_length || {}) || [])[0];
};

const getMaxLengthValue = (config: RawJson) => {
  const value = (Object.keys(config.searcher?.max_length || {}) || [])[0];
  return value
    ? parseInt(config.searcher?.max_length[value])
    : parseInt(config.searcher?.max_length);
};

const trialContinueConfig = (
  experimentConfig: RawJson,
  trialHparams: TrialHyperparameters,
  trialId: number,
  workspaceName: string,
  projectName: string,
): RawJson => {
  const newConfig = structuredClone(experimentConfig);
  return {
    ...newConfig,
    hyperparameters: trialHParamsToExperimentHParams(trialHparams),
    project: projectName,
    searcher: {
      max_length: experimentConfig.searcher.max_length,
      metric: experimentConfig.searcher.metric,
      name: 'single',
      smaller_is_better: experimentConfig.searcher.smaller_is_better,
      source_trial_id: trialId,
    },
    workspace: workspaceName,
  };
};

const CodeEditor = React.lazy(() => import('hew/CodeEditor'));

const DEFAULT_MODAL_STATE = {
  config: {},
  configString: '',
  isAdvancedMode: false,
  open: false,
  type: CreateExperimentType.Fork,
};

const ExperimentContinueModalComponent = ({
  onClose,
  experiment,
  trial,
  type,
}: Props): JSX.Element => {
  const idPrefix = useId();
  const [registryCredentials, setRegistryCredentials] = useState<RawJson>();
  const [modalState, setModalState] = useState<ModalState>(DEFAULT_MODAL_STATE);
  const [disabled, setDisabled] = useState<boolean>(true);

  const isFork = type === CreateExperimentType.Fork;

  const titleLabel = isFork ? 'Reactivate Current Trial' : 'Continue Trial in New Experiment';

  const requiredFields = [EXPERIMENT_NAME, MAX_LENGTH];

  const handleModalClose = () => {
    setModalState(DEFAULT_MODAL_STATE);
    onClose?.();
  };

  const [form] = Form.useForm();

  const handleFieldsChange = () => {
    setModalState((prev) => {
      if (prev.error) return { ...prev, error: undefined };
      const values = form.getFieldsValue();
      if (!prev.isAdvancedMode) {
        prev.config.name = values[EXPERIMENT_NAME];
      }
      if (values[MAX_LENGTH]) {
        const maxLengthType = getMaxLengthType(prev.config);
        if (maxLengthType) {
          prev.config.searcher.max_length[maxLengthType] = prev.config.searcher.max_length + parseInt(values[MAX_LENGTH]);
        } else {
          prev.config.searcher.max_length = prev.config.searcher.max_length + parseInt(values[MAX_LENGTH]);
        }
      }
      prev.configString = yaml.dump(prev.config);
      return prev;
    });

    const hasError = hasErrors(form);
    const values = form.getFieldsValue();
    const missingRequiredFields = Object.entries(values).some(([key, value]) => {
      return requiredFields.includes(key) && !value;
    });
    setDisabled(hasError || missingRequiredFields);
  };

  const handleEditorChange = useCallback((newConfigString: string) => {
    // Update config string and config error upon each keystroke change.
    setModalState((prev) => {
      if (!prev) return prev;

      const newModalState = { ...prev, configString: newConfigString };

      // Validate the yaml syntax by attempting to load it.
      try {
        newModalState.config = yaml.load(newConfigString) as RawJson;
        newModalState.configError = undefined;
        newModalState.error = undefined;
      } catch (e) {
        if (isError(e)) newModalState.configError = e.message;
      }

      return newModalState;
    });
  }, []);

  const toggleMode = useCallback(async () => {
    setModalState((prev) => {
      if (!prev) return prev;

      return {
        ...prev,
        configError: undefined,
        error: undefined,
        isAdvancedMode: !prev.isAdvancedMode,
      };
    });
    // avoid calling form.setFields inside setModalState:
    if (modalState.isAdvancedMode && form) {
      try {
        const newConfig = (yaml.load(modalState.configString) || {}) as RawJson;
        const isFork = modalState.type === CreateExperimentType.Fork;

        form.setFields([
          { name: 'name', value: getExperimentName(newConfig) },
          {
            name: 'maxLength',
            value: !isFork ? getMaxLengthValue(newConfig) : undefined,
          },
        ]);
      } catch (e) {
        handleError(e, { publicMessage: 'failed to load previous yaml config' });
      }
    }
    await form.validateFields();
  }, [form, modalState]);

  const getConfigFromForm = useCallback(
    (config: RawJson) => {
      if (!form) return yaml.dump(config);

      const formValues = form.getFieldsValue();
      const newConfig = structuredClone(config);

      if (formValues[MAX_LENGTH]) {
        const maxLengthType = getMaxLengthType(newConfig);
        if (maxLengthType === undefined) {
          // Unitless searcher config.
          newConfig.searcher.max_length = parseInt(formValues[MAX_LENGTH]);
        } else {
          newConfig.searcher.max_length = { [maxLengthType]: parseInt(formValues[MAX_LENGTH]) };
        }
      }
      return yaml.dump(newConfig);
    },
    [form],
  );

  const submitExperiment = useCallback(
    async (newConfig: string) => {
      const isFork = modalState.type === CreateExperimentType.Fork;
      if (!modalState.experiment || (!isFork && !modalState.trial)) return;
      if (!isFork) {
        try {
          const { experiment: newExperiment, warnings } = await createExperiment({
            activate: true,
            experimentConfig: newConfig,
            parentId: modalState.experiment.id,
          });
          const currentSlotsExceeded = warnings
            ? warnings.includes(V1LaunchWarning.CURRENTSLOTSEXCEEDED)
            : false;
          if (currentSlotsExceeded) {
            handleWarning({
              level: ErrorLevel.Warn,
              publicMessage:
                'The requested job requires more slots than currently available. You may need to increase cluster resources in order for the job to run.',
              publicSubject: 'Current Slots Exceeded',
              silent: false,
              type: ErrorType.Server,
            });
          }
          // Route to reload path to forcibly remount experiment page.
          const newPath = paths.experimentDetails(newExperiment.id);
          routeToReactUrl(paths.reload(newPath));
        } catch (e) {
          let errorMessage = `Unable to ${modalState.type.toLowerCase()} with the provided config.`;
          if (isError(e) && e.name === 'YAMLException') {
            errorMessage = e.message;
          } else if (isDetError(e)) {
            errorMessage = e.publicMessage || e.message;
          }

          setModalState((prev) => ({ ...prev, error: errorMessage }));

          // We throw an error to prevent the modal from closing.
          throw new DetError(errorMessage, { publicMessage: errorMessage, silent: true });
        }
      }
      else {

        try {
          await continueExperiment({
            overrideConfig: newConfig,
            id: modalState.experiment.id,
          });
        } catch (e) {
          let errorMessage = `Unable to ${modalState.type.toLowerCase()} with the provided config.`;
          if (isError(e) && e.name === 'YAMLException') {
            errorMessage = e.message;
          } else if (isDetError(e)) {
            errorMessage = e.publicMessage || e.message;
          }
          setModalState((prev) => ({ ...prev, error: errorMessage }));

          // We throw an error to prevent the modal from closing.
          throw new DetError(errorMessage, { publicMessage: errorMessage, silent: true });
        }
      }
    },
    [modalState],
  );

  const handleSubmit = async () => {
    const error = modalState.error || modalState.configError;
    if (error) throw new Error(error);

    const { isAdvancedMode } = modalState;
    let userConfig, fullConfig;
    if (isAdvancedMode) {
      userConfig = (yaml.load(modalState.configString) || {}) as RawJson;
    } else {
      await form.validateFields();
      userConfig = modalState.config;
    }

    // Add back `registry_auth` if it was stripped and no new `registry_auth` was provided.
    if (!userConfig?.environment?.registry_auth && registryCredentials) {
      const { environment, ...restConfig } = userConfig;
      fullConfig = {
        environment: { registry_auth: registryCredentials, ...environment },
        ...restConfig,
      };
    } else {
      fullConfig = userConfig;
    }

    const configString = isAdvancedMode ? yaml.dump(fullConfig) : getConfigFromForm(fullConfig);
    await submitExperiment(configString);
  };

  useEffect(() => {
    let config = upgradeConfig(experiment.configRaw);

    if (!isFork && trial) {
      config = trialContinueConfig(
        config,
        trial.hyperparameters,
        trial.id,
        experiment.workspaceName,
        experiment.projectName,
      );
      config.description =
        `Continuation of trial ${trial.id}, experiment ${experiment.id}` +
        (config.description ? ` (${config.description})` : '');
    } else if (isFork) {
      if (config.description) config.description = `Fork of ${config.description}`;
    }

    const {
      environment: { registry_auth, ...restEnvironment },
      project: stripIt,
      workspace: stripItToo,
      ...restConfig
    } = config;
    setRegistryCredentials(registry_auth);

    const publicConfig = {
      environment: restEnvironment,
      project: experiment.projectName,
      workspace: experiment.workspaceName,
      ...restConfig,
    };
    setModalState((prev) => {
      const newModalState = {
        ...prev,
        config: publicConfig,
        configString: prev.configString || yaml.dump(publicConfig),
        experiment,
        open: true,
        trial,
        type,
      };
      return _.isEqual(prev, newModalState) ? prev : newModalState;
    });
    setDisabled(!experiment.name); // initial disabled state set here, gets updated later in handleFieldsChange
  }, [experiment, trial, type, isFork, form]);

  if (!experiment || (!isFork && !trial)) return <></>;

  const hideSimpleConfig = isFork && experiment.state !== RunState.Completed;

  return (
    <Modal
      cancel
      size={
        !hideSimpleConfig
          ? modalState.isAdvancedMode
            ? isFork
              ? 'medium'
              : 'large'
            : 'small'
          : 'large'
      }
      submit={{
        disabled,
        form: idPrefix + FORM_ID,
        handleError,
        handler: handleSubmit,
        text: !isFork ? 'Launch Experiment' : 'Reactivate Trial',
      }}
      title={titleLabel}
      onClose={handleModalClose}>
      <>
        {modalState.error && <Message icon="error" title={modalState.error} />}
        {modalState.configError && modalState.isAdvancedMode && (
          <Message icon="error" title={modalState.configError} />
        )}
        {(modalState.isAdvancedMode || hideSimpleConfig) && (
          <React.Suspense fallback={<Spinner spinning tip="Loading text editor..." />}>
            <CodeEditor
              file={Loaded(modalState.configString)}
              files={[{ key: 'config.yaml' }]}
              height="40vh"
              onChange={handleEditorChange}
              onError={handleError}
            />
          </React.Suspense>
        )}
        {!isFork && (
          <Body>Start a new experiment from the current trial&rsquo;s latest checkpoint.</Body>
        )}
        <Form
          form={form}
          hidden={modalState.isAdvancedMode}
          id={idPrefix + FORM_ID}
          labelCol={{ span: 8 }}
          name="basic"
          onFieldsChange={handleFieldsChange}>
          {!isFork && (
            <Form.Item
              initialValue={experiment.name}
              label="Experiment name"
              name={EXPERIMENT_NAME}
              rules={[{ message: 'Please provide a new experiment name.', required: true }]}>
              <Input />
            </Form.Item>
          )}
          {!isFork && (
            <Form.Item
              label={'Max Batches'}
              name={MAX_LENGTH}
              rules={[
                {
                  required: true,
                  validator: (_rule, value) => {
                    let errorMessage = '';
                    if (!value) errorMessage = 'Please provide a max length.';
                    if (value < 1) errorMessage = 'Max length must be at least 1.';
                    return errorMessage ? Promise.reject(errorMessage) : Promise.resolve();
                  },
                },
              ]}>
              <Input type="number" />
            </Form.Item>
          )}
          {isFork && !hideSimpleConfig && (
            <Form.Item
              label={'Additional Batches'}
              name={MAX_LENGTH}
              rules={[
                {
                  required: false,
                  validator: (_rule, value) => {
                    let errorMessage = '';
                    if (value < 0) errorMessage = 'Additional batches must be at least 0.';
                    if (value && !Number.isInteger(value))
                      errorMessage = 'Additional batches must be an integer.';
                    return errorMessage ? Promise.reject(errorMessage) : Promise.resolve();
                  },
                },
              ]}>
              <InputNumber style={{ width: '100%' }} type="number" />
            </Form.Item>
          )}
        </Form>
        <div>
          {!hideSimpleConfig && (
            <Button onClick={toggleMode}>
              {modalState.isAdvancedMode ? SIMPLE_CONFIG_BUTTON_TEXT : FULL_CONFIG_BUTTON_TEXT}
            </Button>
          )}
        </div>
      </>
    </Modal>
  );
};

export default ExperimentContinueModalComponent;