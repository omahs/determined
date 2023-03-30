import React, { useCallback, useState } from 'react';

import CheckpointModalComponent from 'components/CheckpointModalComponent';
import CheckpointRegisterModalComponent from 'components/CheckpointRegisterModalComponent';
import Button from 'components/kit/Button';
import { useModal } from 'components/kit/Modal';
import Tooltip from 'components/kit/Tooltip';
import ModelCreateModal from 'components/ModelCreateModal';
import Icon from 'shared/components/Icon/Icon';
import { ModalCloseReason } from 'shared/hooks/useModal/useModal';
import { CheckpointWorkloadExtended, CoreApiGenericCheckpoint, ExperimentBase } from 'types';

interface Props {
  checkpoint: CheckpointWorkloadExtended | CoreApiGenericCheckpoint;
  children?: React.ReactNode;
  experiment: ExperimentBase;
  title: string;
}

const CheckpointModalTrigger: React.FC<Props> = ({
  checkpoint,
  experiment,
  title,
  children,
}: Props) => {
  const [modelName, setModelName] = useState<string>();
  const CheckpointModal = useModal(CheckpointModalComponent);
  const CheckpointRegisterModal = useModal(CheckpointRegisterModalComponent);
  const modelCreateModal = useModal(ModelCreateModal);

  // const {
  //   contextHolder: modalCheckpointRegisterContextHolder,
  //   modalOpen: openModalCheckpointRegister,
  // } = useModalCheckpointRegister({
  //   onClose: (reason?: ModalCloseReason, checkpoints?: string[]) => {
  //     // TODO: fix the behavior along with checkpoint modal migration
  //     // It used to open checkpoint modal again after creating a model,
  //     // but it doesn't with new create model modal since we don't use context holder anymore.
  //     // This should be able to fix it along with checkpoint modal migration.
  //     if (checkpoints) modelCreateModal.open();
  //   },
  // });

  const handleOnCloseCreateModel = useCallback(
    (reason?: ModalCloseReason, checkpoints?: string[], modelName?: string) => {
      if (modelName) setModelName(modelName);
      if (checkpoints) CheckpointRegisterModal.open();
    },
    [],
  );

  const handleOnCloseCheckpoint = useCallback(
    (reason?: ModalCloseReason) => {
      if (reason === ModalCloseReason.Ok && checkpoint.uuid) {
        CheckpointRegisterModal.open();
      }
    },
    [checkpoint],
  );

  const handleModalCheckpointClick = useCallback(() => {
    CheckpointModal.open();
  }, []);

  const onClose = (reason?: ModalCloseReason, checkpoints?: string[]) => {
    modelCreateModal.open();
  };

  return (
    <>
      <span onClick={handleModalCheckpointClick}>
        {children !== undefined ? (
          children
        ) : (
          <Tooltip title="View Checkpoint">
            <Button aria-label="View Checkpoint" icon={<Icon name="checkpoint" />} />
          </Tooltip>
        )}
      </span>
      {checkpoint.uuid && (
        <CheckpointRegisterModal.Component
          checkpoints={checkpoint.uuid}
          selectedModelName={modelName}
          onClose={onClose}
        />
      )}
      <CheckpointModal.Component
        checkpoint={checkpoint}
        config={experiment.config}
        title={title}
        onClose={handleOnCloseCheckpoint}
      />
      <modelCreateModal.Component onClose={handleOnCloseCreateModel} />
    </>
  );
};

export default CheckpointModalTrigger;
