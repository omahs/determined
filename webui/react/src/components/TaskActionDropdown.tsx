import Button from 'determined-ui/Button';
import Dropdown, { MenuItem } from 'determined-ui/Dropdown';
import Icon from 'determined-ui/Icon';
import useConfirm from 'determined-ui/useConfirm';
import React from 'react';

import css from 'components/ActionDropdown/ActionDropdown.module.scss';
import usePermissions from 'hooks/usePermissions';
import { paths } from 'routes/utils';
import { killTask } from 'services/api';
import { ExperimentAction as Action, AnyTask, CommandTask, DetailedUser } from 'types';
import handleError, { ErrorLevel, ErrorType } from 'utils/error';
import { capitalize } from 'utils/string';
import { isTaskKillable } from 'utils/task';

import Link from './Link';

interface Props {
  children?: React.ReactNode;
  curUser?: DetailedUser;
  onComplete?: (action?: Action) => void;
  onVisibleChange?: (visible: boolean) => void;
  task: AnyTask;
}

const TaskActionDropdown: React.FC<Props> = ({ task, onComplete, children }: Props) => {
  const { canModifyWorkspaceNSC } = usePermissions();
  const isKillable = isTaskKillable(
    task,
    canModifyWorkspaceNSC({ workspace: { id: task.workspaceId } }),
  );

  const confirm = useConfirm();

  const menuItems: MenuItem[] = [
    {
      key: Action.ViewLogs,
      label: <Link path={paths.taskLogs(task as CommandTask)}>View Logs</Link>,
    },
  ];

  if (isKillable) menuItems.unshift({ key: Action.Kill, label: 'Kill' });

  const handleDropdown = (key: string) => {
    try {
      switch (key) {
        case Action.Kill:
          confirm({
            content: 'Are you sure you want to kill this task?',
            danger: true,
            okText: 'Kill',
            onConfirm: async () => {
              await killTask(task as CommandTask);
              onComplete?.(key);
            },
            onError: handleError,
            title: 'Confirm Task Kill',
          });
          break;
        case Action.ViewLogs:
          break;
      }
    } catch (e) {
      handleError(e, {
        level: ErrorLevel.Error,
        publicMessage: `Unable to ${key} task ${task.id}.`,
        publicSubject: `${capitalize(key)} failed.`,
        silent: false,
        type: ErrorType.Server,
      });
    }
    // TODO show loading indicator when we have a button component that supports it.
  };

  return children ? (
    <Dropdown isContextMenu menu={menuItems} onClick={handleDropdown}>
      {children}
    </Dropdown>
  ) : (
    <div className={css.base} title="Open actions menu">
      <Dropdown menu={menuItems} placement="bottomRight" onClick={handleDropdown}>
        <Button
          icon={<Icon name="overflow-vertical" size="small" title="Action menu" />}
          type="text"
        />
      </Dropdown>
    </div>
  );
};

export default TaskActionDropdown;
