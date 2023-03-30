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

  const handleOnCloseCreateModel = useCallback(
    (reason?: ModalCloseReason, checkpoints?: string[], modelName?: string) => {
      if (modelName) setModelName(modelName);
      if (checkpoints) CheckpointRegisterModal.open();
    },
    [CheckpointRegisterModal],
  );

  const handleOnCloseCheckpoint = useCallback(
    (reason?: ModalCloseReason) => {
      if (reason === ModalCloseReason.Ok && checkpoint.uuid) {
        CheckpointRegisterModal.open();
      }
    },
    [checkpoint, CheckpointRegisterModal],
  );

  const handleModalCheckpointClick = useCallback(() => {
    CheckpointModal.open();
  }, [CheckpointModal]);

  const onClose = () => {
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
