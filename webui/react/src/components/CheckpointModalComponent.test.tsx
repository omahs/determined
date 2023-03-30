import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, { useCallback } from 'react';

import Button from 'components/kit/Button';
import { StoreProvider as UIProvider } from 'shared/contexts/stores/UI';
import { ModalCloseReason } from 'shared/hooks/useModal/useModal';
import { generateTestExperimentData } from 'storybook/shared/generateTestData';
import { useModal } from 'components/kit/Modal';
import CheckpointModalComponent from 'components/CheckpointModalComponent';
const TEST_MODAL_TITLE = 'Checkpoint Modal Test';
const MODAL_TRIGGER_TEXT = 'Open Checkpoint Modal';
const REGISTER_CHECKPOINT_TEXT = 'Register Checkpoint';

vi.mock('services/api', () => ({
  getModels: () => {
    return Promise.resolve({ models: [] });
  },
}));

const { experiment, checkpoint } = generateTestExperimentData();

const Container: React.FC = () => {
  const CheckpointModal = useModal(CheckpointModalComponent);
  <CheckpointModal.Component
  checkpoint={checkpoint}
  config={experiment.config}
/>
  const handleClick = useCallback(() => CheckpointModal.open(), []);

  return (
    <UIProvider>
      <Button onClick={handleClick}>{MODAL_TRIGGER_TEXT}</Button>
    </UIProvider>
  );
};

const setup = async () => {
  const user = userEvent.setup();

  render(<Container />);

  await user.click(screen.getByText(MODAL_TRIGGER_TEXT));

  return user;
};

describe('useModalCheckpoint', () => {
  it('should open modal', async () => {
    await setup();

    expect(await screen.findByText(TEST_MODAL_TITLE)).toBeInTheDocument();
  });

  it('should close modal', async () => {
    const onClose = vi.fn();
    const user = await setup();

    await screen.findByText(TEST_MODAL_TITLE);

    await user.click(screen.getByRole('button', { name: /cancel/i }));

    expect(onClose).toHaveBeenCalledWith(ModalCloseReason.Cancel);

    await waitFor(() => {
      expect(screen.queryByText(TEST_MODAL_TITLE)).not.toBeInTheDocument();
    });
  });

  it('should call `onClose` handler with Okay', async () => {
    const onClose = vi.fn();
    const user = await setup();

    await screen.findByText(TEST_MODAL_TITLE);

    await user.click(screen.getByRole('button', { name: REGISTER_CHECKPOINT_TEXT }));

    expect(onClose).toHaveBeenCalledWith(ModalCloseReason.Ok);
  });
});
