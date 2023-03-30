import React, { useCallback } from 'react';

import CheckpointModalComponent from 'components/CheckpointModalComponent';
import CheckpointRegisterModalComponent from 'components/CheckpointRegisterModalComponent';
import Button from 'components/kit/Button';
import { useModal } from 'components/kit/Modal';
import Tooltip from 'components/kit/Tooltip';
import useModalModelCreate from 'hooks/useModal/Model/useModalModelCreate';
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
  const CheckpointModal = useModal(CheckpointModalComponent);
  const CheckpointRegisterModal = useModal(CheckpointRegisterModalComponent);

  const handleOnCloseCreateModel = useCallback(
    (reason?: ModalCloseReason, checkpoints?: string[], modelName?: string) => {
      if (checkpoints) CheckpointRegisterModal.open();
    },
    [],
  );

  const { contextHolder: modalModelCreateContextHolder, modalOpen: openModalCreateModel } =
    useModalModelCreate({ onClose: handleOnCloseCreateModel });

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
      {modalModelCreateContextHolder}
      {checkpoint.uuid && <CheckpointRegisterModal.Component checkpoints={checkpoint.uuid} />}
      <CheckpointModal.Component
        checkpoint={checkpoint}
        config={experiment.config}
        onClose={handleOnCloseCheckpoint}
      />
    </>
  );
};

export default CheckpointModalTrigger;
