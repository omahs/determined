import { Form, Modal } from 'antd';
import dayjs, { Dayjs } from 'dayjs';
import DatePicker from 'hew/DatePicker';
import Select, { Option } from 'hew/Select';
import { useTheme } from 'hew/Theme';
import React from 'react';

import { handlePath, serverAddress } from 'routes/utils';
import { ValueOf } from 'types';

export const CSVGroupBy = {
  Allocations: '/resources/allocation/allocations-csv?',
  Workloads: '/resources/allocation/raw?',
} as const;

export type CSVGroupBy = ValueOf<typeof CSVGroupBy>;

interface Props {
  afterDate: Dayjs;
  beforeDate: Dayjs;
  groupBy: CSVGroupBy;
  onVisibleChange: (visible: boolean) => void;
}

const ClusterHistoricalUsageCsvModal: React.FC<Props> = ({
  afterDate,
  beforeDate,
  groupBy,
  onVisibleChange,
}: Props) => {
  const [form] = Form.useForm();

  const {
    themeSettings: { className: themeClass },
  } = useTheme();

  const handleOk = (event: React.MouseEvent): void => {
    const formAfterDate = form.getFieldValue('afterDate');
    const formBeforeDate = form.getFieldValue('beforeDate');
    const groupByEndpoint = form.getFieldValue('groupBy');
    const searchParams = new URLSearchParams();

    searchParams.append('timestamp_after', formAfterDate.startOf('day').toISOString());
    searchParams.append('timestamp_before', formBeforeDate.endOf('day').toISOString());
    handlePath(event, {
      external: true,
      path: serverAddress(groupByEndpoint + searchParams.toString()),
    });
    onVisibleChange(false);
  };

  const isAfterDateDisabled = (currentDate: Dayjs) => {
    const formBeforeDate = form.getFieldValue('beforeDate');
    return currentDate.isAfter(formBeforeDate);
  };

  const isBeforeDateDisabled = (currentDate: Dayjs) => {
    const formAfterDate = form.getFieldValue('afterDate');
    return currentDate.isBefore(formAfterDate) || currentDate.isAfter(dayjs());
  };

  return (
    <Modal
      okText="Proceed to Download"
      open={true}
      title="Download Resource Usage Data in CSV"
      wrapClassName={themeClass}
      onCancel={() => onVisibleChange(false)}
      onOk={handleOk}>
      <Form form={form} initialValues={{ afterDate, beforeDate, groupBy }} labelCol={{ span: 8 }}>
        <Form.Item label="Start" name="afterDate">
          <DatePicker allowClear={false} disabledDate={isAfterDateDisabled} width={150} />
        </Form.Item>
        <Form.Item label="End" name="beforeDate">
          <DatePicker allowClear={false} disabledDate={isBeforeDateDisabled} width={150} />
        </Form.Item>
        <Form.Item label="Group by" name="groupBy">
          <Select searchable={false} width={'150px'}>
            <Option value={CSVGroupBy.Workloads}>Workloads</Option>
            <Option value={CSVGroupBy.Allocations}>Allocations</Option>
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ClusterHistoricalUsageCsvModal;
