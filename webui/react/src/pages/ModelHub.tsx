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
import type { UploadProps } from 'antd';
import { message, Upload } from 'antd';
const {Dragger} = Upload;
import Papa from "papaparse";

const ModelDetails: React.FC = () => {
  const pageRef = useRef<HTMLElement>(null);

  const [showLoad, setShowLoad] = useState<boolean>();
  const [files, setFiles] = useState<RcFile[]>();
  const [filename, setFilename] = useState<string>();
  const [filesize, setFileSize] = useState<number>();
  const [samples, setSamples] = useState<number>();
    // State to store parsed data
    const [parsedData, setParsedData] = useState<any>([]);

    //State to store table Column name
    const [tableRows, setTableRows] = useState<any>([]);

    console.log("tableRows")
  console.log(tableRows)
  console.log(parsedData)
  const humanFileSize = (bytes: any, si=true, dp=1): string => {
    const thresh = si ? 1000 : 1024;
  
    if (Math.abs(bytes) < thresh) {
      return bytes + ' B';
    }
  
    const units = si 
      ? ['kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'] 
      : ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];
    let u = -1;
    const r = 10**dp;
  
    do {
      bytes /= thresh;
      ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);
  
  
    return bytes.toFixed(dp) + ' ' + units[u];
  }

  const cols = tableRows.map((row: any) => {return {
    dataIndex: row,
    key: row,
    // render: modelVersionNumberRenderer,
    sorter: true,
    title: row,
  }})

  const changeHandler = (event: any) => {
    console.log("event",event)
    const file = event.file
    setFilename(file.name)
    setFileSize(file.size)
    // Passing file data (event.target.files[0]) to parse using Papa.parse
    Papa.parse(event.file, {
      header: true,
      skipEmptyLines: true,
      complete: function (results) {
        const rowsArray: any = [];
        const valuesArray: any = [];

        // Iterating data to get column name and their values
        results.data.map((d: any) => {
          rowsArray.push(Object.keys(d));
        });

        // Parsed Data Response in array format
        setParsedData(results.data);

        // Filtered Column Names
        setTableRows(rowsArray[0].slice(1));

      },
    });
  };

  // const props: UploadProps = {
  //   name: 'file',
  //   multiple: true,
  //   action: 'https://www.mocky.io/v2/5cc8019d300000980a055e76',
  //   showUploadList: {false} beforeUpload={(file, fileList) => {
  //     setFiles(fileList)
  //     return false;
  //   }}
  //   onDrop(e) {
  //     console.log('Dropped files', e.dataTransfer.files);
  //   },
  // };

  // files?.forEach(f => {
  //   const n = f.webkitRelativePath;
  //   const b = n.split("/");
  //   labels.add(b[1]);
  // })
  return (
    <Page
      docTitle="Model Details"
      id="modelHub"
      title="Model Hub"
      >
        { !showLoad && (
          <div className={css.center}>
        <Button type="primary" onClick={() => setShowLoad(true)} size='large'>
          Fine-tune a model
        </Button>
        </div>
        )
        }
        {showLoad && tableRows.length == 0 && (
          <div className={css.center}>
        <Dragger
        onChange={changeHandler}
        showUploadList={false} 
        beforeUpload={(file, fileList) => {
          setFiles(fileList)
          return false;
        }}>
          <div className={css.dragger}>
       <p className="ant-upload-drag-icon">
      <UploadOutlined />
    </p>
    <p className="ant-upload-text">Drag and drop your dataset here.</p>
    <Button type='link'>
      Learn more how your data should be structured
    </Button>
    </div>
  </Dragger> </div>)
        }
      { tableRows.length > 0 && (<>
      <div className={css.fileInfo}>
        <h3 style={{fontSize:"15px"}}>Your dataset</h3>
      <p>Filename</p>
      <p> &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;Samples</p>
      <p> &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;Size</p>
      <br />
      <p>{filename}</p>
      <p> &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;{parsedData.length}</p>
      <p> &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;{humanFileSize(filesize)}</p>
      </div>
      <InteractiveTable 
      containerRef={pageRef}
      columns={cols}
      dataSource={parsedData}
      settings={{
        columns: tableRows
      } as InteractiveTableSettings}
      updateSettings={() => {}}
      /></>)}
    </Page>
  );
};

export default ModelDetails;
