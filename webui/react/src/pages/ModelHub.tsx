import { Typography } from 'antd';
import { FilterValue, SorterResult, TablePaginationConfig } from 'antd/lib/table/interface';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import Form from 'components/kit/Form';
import Button from 'components/kit/Button';
import Input from 'components/kit/Input';
import { useModal } from 'components/kit/Modal';
import Tags, { tagsActionHelper } from 'components/kit/Tags';
import MetadataCard from 'components/Metadata/MetadataCard';
import ModelDownloadModal from 'components/ModelDownloadModal';
import ModelVersionDeleteModal from 'components/ModelVersionDeleteModal';
import NotesCard from 'components/NotesCard';
import Page from 'components/Page';
import { Upload } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import PageNotFound from 'components/PageNotFound';
import InteractiveTable, {
  ColumnDef,
  InteractiveTableSettings,
} from 'components/Table/InteractiveTable';
import {
  defaultRowClassName,
  getFullPaginationConfig,
  modelVersionNameRenderer,
  modelVersionNumberRenderer,
  relativeTimeRenderer,
  userRenderer,
} from 'components/Table/Table';
import usePermissions from 'hooks/usePermissions';
import { UpdateSettings, useSettings } from 'hooks/useSettings';
import {
  archiveModel,
  getModelDetails,
  patchModel,
  patchModelVersion,
  unarchiveModel,
} from 'services/api';
import { V1GetModelVersionsRequestSortBy } from 'services/api-ts-sdk';
import Message, { MessageType } from 'shared/components/Message';
import Spinner from 'shared/components/Spinner/Spinner';
import usePolling from 'shared/hooks/usePolling';
import { isEqual } from 'shared/utils/data';
import { ErrorType } from 'shared/utils/error';
import { isAborted, isNotFound, validateDetApiEnum } from 'shared/utils/service';
import usersStore from 'stores/users';
import { useEnsureWorkspacesFetched, useWorkspaces } from 'stores/workspaces';
import { Metadata, ModelVersion, ModelVersions } from 'types';
import handleError from 'utils/error';
import { Loadable, NotLoaded } from 'utils/loadable';
import { useObservable } from 'utils/observable';
import Select, { Option }  from 'components/kit/Select';
import settingsConfig, {
  DEFAULT_COLUMN_WIDTHS,
  isOfSortKey,
  Settings,
} from './ModelDetails/ModelDetails.settings';
import ModelHeader from './ModelDetails/ModelHeader';
import ModelVersionActionDropdown from './ModelDetails/ModelVersionActionDropdown';
import css from './ModelHub.module.scss';
import { RcFile } from 'antd/es/upload';


const ModelDetails: React.FC = () => {
  
  const [files, setFiles] = useState<RcFile[]>();

  const labels = new Set<string>();

  files?.forEach(f => {
    const n = f.webkitRelativePath;
    const b = n.split("/");
    labels.add(b[1]);
  })
  return (
    <Page
      docTitle="Model Details"
      id="modelHub"
      >
      <div className={css.base}>
        <>
        <h1>Select Domain</h1>
        <Form>
          <Form.Item label='Select Domain'>
        <Select>
          <Option> Computer Vision </Option>
          <Option> NLP </Option>
          <Option> Audio/Speech </Option>
        </Select>
        </Form.Item>
        <Form.Item label='Select Task'>
        <Select>
          <Option> Image Classification</Option>
          <Option> Semantic Segmentation </Option>
          <Option> Object Detection </Option>
        </Select>
        </Form.Item>
        <Form.Item label='Select Dataset Folder'>
        <Upload showUploadList={false} beforeUpload={(file, fileList) => {
          setFiles(fileList)
          return false;
        }} directory>
    <Button icon={<UploadOutlined />}>Upload Directory</Button>
  </Upload>
        </Form.Item>
        </Form>
        <h2>Classes</h2>
        
        {Array.from(labels).map((f) => <p>{f}</p>)}
    </>
      </div>
    </Page>
  );
};

export default ModelDetails;
