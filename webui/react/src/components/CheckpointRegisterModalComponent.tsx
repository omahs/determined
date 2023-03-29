import { ModalFuncProps, Select } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import Input from 'components/kit/Input';
import Tags, { tagsActionHelper } from 'components/kit/Tags';
import Link from 'components/Link';
import EditableMetadata from 'components/Metadata/EditableMetadata';
import usePermissions from 'hooks/usePermissions';
import { paths } from 'routes/utils';
import { getModels, postModelVersion } from 'services/api';
import { V1GetModelsRequestSortBy } from 'services/api-ts-sdk';
import useModal, { ModalHooks as Hooks, ModalCloseReason } from 'shared/hooks/useModal/useModal';
import usePrevious from 'shared/hooks/usePrevious';
import { isEqual } from 'shared/utils/data';
import { ErrorType } from 'shared/utils/error';
import { validateDetApiEnum } from 'shared/utils/service';
import { pluralizer } from 'shared/utils/string';
import { Metadata, ModelItem } from 'types';
import { notification } from 'utils/dialogApi';
import handleError from 'utils/error';

import css from './CheckpointRegisterModal.module.scss';

import { Modal } from 'components/kit/Modal';

interface Props {
  onClose?: (reason?: ModalCloseReason, checkpoints?: string[]) => void;
  checkpoints: string | string[];
  selectedModelName?: string;
}


interface ModalState {
  checkpoints?: string[];
  expandDetails: boolean;
  metadata: Metadata;
  models: ModelItem[];
  selectedModelName?: string;
  tags: string[];
  versionDescription: string;
  versionName: string;
}

const INITIAL_MODAL_STATE = {
  expandDetails: false,
  metadata: {},
  models: [],
  tags: [],
  versionDescription: '',
  versionName: '',
};

const CheckpointRegisterModalComponent: React.FC<Props> = ({ checkpoints, selectedModelName, onClose }: Props) => {
  const [canceler] = useState(new AbortController());
  const [modalState, setModalState] = useState<ModalState>(INITIAL_MODAL_STATE);

  const { canCreateModelVersion } = usePermissions();

  const handleClose = useCallback(
    (reason?: ModalCloseReason) => {
      setModalState(INITIAL_MODAL_STATE);
      onClose?.(reason);
    },
    [onClose],
  );

  const { modalClose, modalOpen: openOrUpdate, ...modalHook } = useModal({ onClose: handleClose });

  const selectedModelNumVersions = useMemo(() => {
    return (
      modalState.models.find((model) => model.name === modalState.selectedModelName)?.numVersions ??
      0
    );
  }, [modalState.models, modalState.selectedModelName]);

  const modelOptions = useMemo(() => {
    return modalState.models
      .filter((model) => canCreateModelVersion({ model }))
      .map((model) => ({ id: model.id, name: model.name }));
  }, [modalState.models, canCreateModelVersion]);

  const registerModelVersion = useCallback(
    async (state: ModalState) => {
      const { selectedModelName, versionDescription, tags, metadata, versionName, checkpoints } =
        state;
      if (!selectedModelName || !checkpoints) return;
      try {
        if (checkpoints.length === 1) {
          const response = await postModelVersion({
            body: {
              checkpointUuid: checkpoints[0],
              comment: versionDescription,
              labels: tags,
              metadata,
              modelName: selectedModelName,
              name: versionName,
            },
            modelName: selectedModelName,
          });

          if (!response) return;

          modalClose(ModalCloseReason.Ok);

          notification.open({
            btn: null,
            description: (
              <div className={css.toast}>
                <p>{`"${versionName || `Version ${selectedModelNumVersions + 1}`}"`} registered</p>
                <Link path={paths.modelVersionDetails(selectedModelName, response.version)}>
                  View Model Version
                </Link>
              </div>
            ),
            message: '',
          });
        } else {
          for (const checkpointUuid of checkpoints) {
            await postModelVersion({
              body: {
                checkpointUuid,
                comment: versionDescription,
                labels: tags,
                metadata,
                modelName: selectedModelName,
              },
              modelName: selectedModelName,
            });
          }
          modalClose(ModalCloseReason.Ok);

          notification.open({
            btn: null,
            description: (
              <div className={css.toast}>
                <p>{checkpoints.length} versions registered</p>
                <Link path={paths.modelDetails(selectedModelName)}>View Model</Link>
              </div>
            ),
            message: '',
          });
        }
      } catch (e) {
        handleError(e, {
          publicSubject: `Unable to register ${pluralizer(checkpoints.length, 'checkpoint')}.`,
          silent: true,
          type: ErrorType.Api,
        });
      }
    },
    [modalClose, selectedModelNumVersions],
  );

  const handleOk = useCallback(
    async (state: ModalState) => {
      await registerModelVersion(state);
    },
    [registerModelVersion],
  );

  const updateModel = useCallback((value?: string) => {
    setModalState((prev) => ({ ...prev, selectedModelName: value }));
  }, []);

  const updateVersionName = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setModalState((prev) => ({ ...prev, versionName: e.target.value }));
  }, []);

  const updateVersionDescription = useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setModalState((prev) => ({ ...prev, versionDescription: e.target.value }));
  }, []);

  const openDetails = useCallback(() => {
    setModalState((prev) => ({ ...prev, expandDetails: true }));
  }, []);

  const updateMetadata = useCallback((value: Metadata) => {
    setModalState((prev) => ({ ...prev, metadata: value }));
  }, []);

  const updateTags = useCallback((value: string[]) => {
    setModalState((prev) => ({ ...prev, tags: value }));
  }, []);

  const launchNewModelModal = useCallback(
    (state: ModalState) => {
      modalClose(ModalCloseReason.Cancel);
      onClose?.(ModalCloseReason.Cancel, state.checkpoints);
    },
    [modalClose, onClose],
  );

  const fetchModels = useCallback(async () => {
    try {
      const response = await getModels(
        {
          archived: false,
          orderBy: 'ORDER_BY_DESC',
          sortBy: validateDetApiEnum(
            V1GetModelsRequestSortBy,
            V1GetModelsRequestSortBy.LASTUPDATEDTIME,
          ),
        },
        { signal: canceler.signal },
      );
      setModalState((prev) => {
        if (isEqual(prev.models, response.models)) return prev;
        return { ...prev, models: response.models };
      });
    } catch (e) {
      handleError(e, {
        publicSubject: 'Unable to fetch models.',
        silent: true,
        type: ErrorType.Api,
      });
    }
  }, [canceler.signal]);
  
  // const getModalProps = useCallback(
  //   (state: ModalState): Partial<ModalFuncProps> => {
  //     const { selectedModelName } = state;

  //     const modalProps = {
  //       className: css.base,
  //       closable: true,
  //       content: getModalContent(state),
  //       icon: null,
  //       maskClosable: true,
  //       okButtonProps: { disabled: selectedModelName == null },
  //       okText: 'Register Checkpoint',
  //       onCancel: handleCancel,
  //       onOk: () => handleOk(state),
  //       title: 'Register Checkpoint',
  //     };

  //     return modalProps;
  //   },
  //   [getModalContent, handleCancel, handleOk],
  // );

  useEffect(() => {
    fetchModels();
  }, [fetchModels]);

  useEffect(() => {
    return () => canceler.abort();
  }, [canceler]);

  const handleCancel = useCallback(() => modalClose(), [modalClose]);

      const {
        versionDescription,
        tags,
        metadata,
        versionName,
        expandDetails,
      } = modalState;

      // We always render the form regardless of mode to provide a reference to it.
      return (
        <Modal
         cancel
         submit={{
          text: 'Register Checkpoint',
          handler: () => handleOk(modalState),
         }}
         title='Register Checkpoint'
        >
        <div className={css.base}>
          <p className={css.directions}>Save this checkpoint to the Model Registry</p>
          <div>
            <div className={css.selectModelRow}>
              <h2>Select Model</h2>
              <p onClick={() => launchNewModelModal(modalState)}>New Model</p>
            </div>
            <Select
              optionFilterProp="label"
              options={modelOptions.map((option) => ({ label: option.name, value: option.name }))}
              placeholder="Select a model..."
              showSearch
              style={{ width: '100%' }}
              value={modalState.selectedModelName}
              onChange={updateModel}
            />
          </div>
          {modalState.selectedModelName && (
            <>
              <div className={css.separator} />
              <div>
                <h2>Version Name</h2>
                <Input
                  disabled={modalState.checkpoints?.length != null && modalState.checkpoints.length > 1}
                  placeholder={`Version ${selectedModelNumVersions + 1}`}
                  value={versionName}
                  onChange={updateVersionName}
                />
                {modalState.checkpoints?.length != null && modalState.checkpoints.length > 1 && (
                  <p>Cannot specify version name when batch registering.</p>
                )}
              </div>
              <div>
                <h2>
                  Description <span>(optional)</span>
                </h2>
                <Input.TextArea value={versionDescription} onChange={updateVersionDescription} />
              </div>
              {expandDetails ? (
                <>
                  <div>
                    <h2>
                      Metadata <span>(optional)</span>
                    </h2>
                    <EditableMetadata
                      editing={true}
                      metadata={metadata}
                      updateMetadata={updateMetadata}
                    />
                  </div>
                  <div>
                    <h2>
                      Tags <span>(optional)</span>
                    </h2>
                    <Tags tags={tags} onAction={tagsActionHelper(tags, updateTags)} />
                  </div>
                </>
              ) : (
                <p className={css.expandDetails} onClick={openDetails}>
                  Add More Details...
                </p>
              )}
            </>
          )}
        </div>
        </Modal>
      );

};

export default CheckpointRegisterModalComponent;
