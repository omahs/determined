import React, { useCallback, useMemo, useState } from 'react';

import CheckpointModalComponent from 'components/CheckpointModalComponent';
import CheckpointRegisterModalComponent from 'components/CheckpointRegisterModalComponent';
import Card from 'components/kit/Card';
import { useModal } from 'components/kit/Modal';
import ModelCreateModal from 'components/ModelCreateModal';
import OverviewStats from 'components/OverviewStats';
import Section from 'components/Section';
import TimeAgo from 'components/TimeAgo';
import { ModalCloseReason } from 'shared/hooks/useModal/useModal';
import { humanReadableBytes } from 'shared/utils/string';
import { CheckpointWorkloadExtended, ExperimentBase, TrialDetails } from 'types';

interface Props {
  experiment: ExperimentBase;
  trial?: TrialDetails;
}

const TrialInfoBox: React.FC<Props> = ({ trial, experiment }: Props) => {
  const CheckpointModal = useModal(CheckpointModalComponent);
  const CheckpointRegisterModal = useModal(CheckpointRegisterModalComponent);
  const bestCheckpoint: CheckpointWorkloadExtended | undefined = useMemo(() => {
    if (!trial) return;
    const cp = trial.bestAvailableCheckpoint;
    if (!cp) return;

    return {
      ...cp,
      experimentId: trial.experimentId,
      trialId: trial.id,
    };
  }, [trial]);

  const totalCheckpointsSize = useMemo(() => {
    const totalBytes = trial?.totalCheckpointSize;
    if (!totalBytes) return;
    return humanReadableBytes(totalBytes);
  }, [trial?.totalCheckpointSize]);

  const modelCreateModal = useModal(ModelCreateModal);

  const handleOnCloseCreateModel = useCallback(
    (reason?: ModalCloseReason, checkpoints?: string[], modelName?: string) => {
      if (checkpoints) CheckpointRegisterModal.open();
    },
    [],
  );

  const handleOnCloseCheckpoint = useCallback(
    (reason?: ModalCloseReason) => {
      if (reason === ModalCloseReason.Ok && bestCheckpoint?.uuid) {
        CheckpointRegisterModal.open();
      }
    },
    [bestCheckpoint],
  );

  return (
    <Section>
      <Card.Group size="small">
        {trial?.runnerState && (
          <OverviewStats title="Last Runner State">{trial.runnerState}</OverviewStats>
        )}
        {trial?.startTime && (
          <OverviewStats title="Started">
            <TimeAgo datetime={trial.startTime} />
          </OverviewStats>
        )}
        {totalCheckpointsSize && (
          <OverviewStats title="Checkpoints">{`${trial?.checkpointCount} (${totalCheckpointsSize})`}</OverviewStats>
        )}
        {bestCheckpoint && bestCheckpoint.uuid && (
          <>
            <OverviewStats title="Best Checkpoint" onClick={() => CheckpointModal.open()}>
              Batch {bestCheckpoint.totalBatches}
            </OverviewStats>
            <CheckpointModal.Component
              checkpoint={bestCheckpoint}
              config={experiment.config}
              onClose={handleOnCloseCheckpoint}
            />
            <CheckpointRegisterModal.Component checkpoints={bestCheckpoint.uuid} />
            <modelCreateModal.Component onClose={handleOnCloseCreateModel} />
          </>
        )}
      </Card.Group>
    </Section>
  );
};

export default TrialInfoBox;

export const TrialInfoBoxMultiTrial: React.FC<Props> = ({ experiment }: Props) => {
  const searcher = experiment.config.searcher;
  const checkpointsSize = useMemo(() => {
    const totalBytes = experiment?.checkpointSize;
    if (!totalBytes) return;
    return humanReadableBytes(totalBytes);
  }, [experiment]);
  return (
    <Section>
      <Card.Group size="small">
        {searcher?.metric && <OverviewStats title="Metric">{searcher.metric}</OverviewStats>}
        {searcher?.name && <OverviewStats title="Searcher">{searcher.name}</OverviewStats>}
        {experiment.numTrials > 0 && (
          <OverviewStats title="Trials">{experiment.numTrials}</OverviewStats>
        )}
        {checkpointsSize && (
          <OverviewStats title="Checkpoints">
            {`${experiment.checkpointCount} (${checkpointsSize})`}
          </OverviewStats>
        )}
      </Card.Group>
    </Section>
  );
};
