import { ExclamationCircleOutlined } from '@ant-design/icons';
import { ModalFuncProps } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';

import { detApi } from 'services/apiConfig';
import { readStream } from 'services/utils';
import useModal, { ModalHooks as Hooks, ModalCloseReason } from 'shared/hooks/useModal/useModal';
import { pluralizer } from 'shared/utils/string';
import handleError from 'utils/error';

import { Modal } from './kit/Modal';

export interface Props {
  onClose?: (reason?: ModalCloseReason) => void;
  checkpoints: string | string[];
  initialModalProps?: ModalFuncProps;
}

const CheckpointDeleteModal: React.FC<Props> = ({ onClose, checkpoints }: Props) => {
  const numCheckpoints = useMemo(() => {
    if (Array.isArray(checkpoints)) return checkpoints.length;
    return 1;
  }, [checkpoints]);

  const handleCancel = useCallback(() => onClose?.(ModalCloseReason.Cancel), [onClose]);

  const handleDelete = useCallback(() => {
    readStream(
      detApi.Checkpoint.deleteCheckpoints({
        checkpointUuids: Array.isArray(checkpoints) ? checkpoints : [checkpoints],
      }),
    );
    onClose?.(ModalCloseReason.Ok);
  }, [checkpoints, onClose]);

  return (
    <Modal
      cancel
      danger
      icon="warning-large"
      submit={{
        handler: () => handleDelete(),
        text: 'Request Delete',
      }}
      title="Confirm Checkpoint Deletion">
      {`Are you sure you want to request deletion for 
${numCheckpoints} ${pluralizer(numCheckpoints, 'checkpoint')}?
This action may complete or fail without further notification.`}
    </Modal>
  );
};

export default CheckpointDeleteModal;
