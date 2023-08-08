import yaml from 'js-yaml';
import React, { useMemo } from 'react';

import Icon from 'components/kit/Icon';
import Spinner from 'components/kit/Spinner';
import { useLoadable } from 'hooks/useLoadable';
import { paths } from 'routes/utils';
import { getExperimentFileFromTree, getExperimentFileTree } from 'services/api';
import { V1FileNode } from 'services/api-ts-sdk';
import { RawJson } from 'types';
import { ExperimentBase, TreeNode } from 'types';
import handleError from 'utils/error';
import { isSingleTrialExperiment } from 'utils/experiment';
import { Loadable, Loaded, NotLoaded } from 'utils/loadable';

import css from './ExperimentCodeViewer.module.scss';

const CodeEditor = React.lazy(() => import('components/kit/CodeEditor'));

const configIcon = <Icon name="settings" title="settings" />;

export interface Props {
  experiment: ExperimentBase;
  onSelectFile?: (arg0: string) => void;
  selectedFilePath?: string;
}

const ExperimentCodeViewer: React.FC<Props> = ({
  experiment,
  onSelectFile,
  selectedFilePath,
}: Props) => {
  const expFiles = useLoadable(async () => {
    const convertV1FileNodeToTreeNode = (node: V1FileNode): TreeNode => ({
      children: node.files?.map((n) => convertV1FileNodeToTreeNode(n)) ?? [],
      content: NotLoaded,
      download: paths.experimentFileFromTree(experiment.id, String(node.path)),
      get: (path: string) => getExperimentFileFromTree({ experimentId: experiment.id, path }),
      isLeaf: !node.isDir,
      key: node.path ?? '',
      title: node.name,
    });
    const fileTree = await getExperimentFileTree({ experimentId: experiment.id });
    return fileTree.map(convertV1FileNodeToTreeNode);
  }, [experiment.id]);

  const submittedConfig = useMemo(() => {
    if (!experiment.originalConfig) return;

    const { hyperparameters, ...restConfig } = yaml.load(experiment.originalConfig) as RawJson;

    // don't ask me why this works.. it gets rid of the JSON though
    return yaml.dump({ ...restConfig, hyperparameters });
  }, [experiment.originalConfig]);

  const runtimeConfig = useMemo(() => {
    if (!experiment.configRaw) return;

    const {
      environment: { registry_auth, ...restEnvironment },
      workspace,
      project,
      ...restConfig
    } = experiment.configRaw;
    return yaml.dump({ environment: restEnvironment, ...restConfig });
  }, [experiment.configRaw]);

  const fileOpts = [
    submittedConfig
      ? {
          content: Loaded(submittedConfig),
          download: `${experiment.id}_submitted_configuration.yaml`,
          icon: configIcon,
          isLeaf: true,
          key: 'Submitted Configuration',
          title: 'Submitted Configuration',
        }
      : null,
    runtimeConfig
      ? {
          content: Loaded(runtimeConfig),
          download: `${experiment.id}_runtime_configuration.yaml`,
          icon: configIcon,
          isLeaf: true,
          key: 'Runtime Configuration',
          title: 'Runtime Configuration',
        }
      : null,
    ...Loadable.getOrElse([], expFiles),
  ].filter((valid) => !!valid) as TreeNode[];

  const cssClasses = [
    css.codeContainer,
    isSingleTrialExperiment(experiment) || css.multitrialContainer,
  ];

  return (
    <React.Suspense fallback={<Spinner spinning tip="Loading code viewer..." />}>
      <Spinner spinning={expFiles === NotLoaded} tip="Loading file tree...">
        <div className={cssClasses.join(' ')}>
          <CodeEditor
            files={fileOpts}
            readonly={true}
            selectedFilePath={selectedFilePath}
            onError={handleError}
            onSelectFile={onSelectFile}
          />
        </div>
      </Spinner>
    </React.Suspense>
  );
};

export default ExperimentCodeViewer;
