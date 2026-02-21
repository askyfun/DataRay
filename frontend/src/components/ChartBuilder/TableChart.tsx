import { useMemo } from 'react';
import { Table, Spin, Empty, Typography } from 'antd';
import type { TableProps } from 'antd';
import type { QueryConfig } from '../../store';

const { Text } = Typography;

interface TableChartProps {
  data: any[];
  loading: boolean;
  queryConfig?: QueryConfig;
  columns?: string[];
  pagination?: {
    page: number;
    pageSize: number;
    total: number;
  };
  onPageChange?: (page: number, pageSize: number) => void;
}

const TableChart: React.FC<TableChartProps> = ({ 
  data, 
  loading, 
  queryConfig,
  columns: propColumns,
  pagination,
  onPageChange 
}) => {
  const columns: TableProps<any>['columns'] = useMemo(() => {
    const keys = propColumns && propColumns.length > 0 ? propColumns : (data && data.length > 0 ? Object.keys(data[0]) : []);

    return keys.map((key) => ({
      title: key,
      dataIndex: key,
      key,
      sorter: true,
      ellipsis: true,
    }));
  }, [data, propColumns]);

  const handleTableChange: TableProps<any>['onChange'] = (tablePagination, _filters, sorter) => {
    if (onPageChange && tablePagination) {
      onPageChange(tablePagination.current || 1, tablePagination.pageSize || 10);
    }
    
    if (queryConfig && sorter && !Array.isArray(sorter)) {
      const sortField = sorter.field as string | undefined;
      const sortOrder = sorter.order === 'ascend' ? 'asc' : sorter.order === 'descend' ? 'desc' : undefined;

      if (sortField && sortOrder) {
        queryConfig.sort = { field: sortField, order: sortOrder };
      }
    }
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>
          <Text type="secondary">加载数据中...</Text>
        </div>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Empty
        description="暂无数据"
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        style={{ padding: '100px 0' }}
      />
    );
  }

  return (
    <Table
      dataSource={data.map((item, index) => ({ ...item, key: index }))}
      columns={columns}
      pagination={
        pagination
          ? {
              current: pagination.page,
              pageSize: pagination.pageSize,
              total: pagination.total,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
              showTotal: (total) => `总计 ${total} 条`,
              showQuickJumper: true,
            }
          : {
              defaultPageSize: 10,
              showSizeChanger: true,
              pageSizeOptions: ['10', '20', '50', '100'],
              showTotal: (total, range) => `${range[0]}-${range[1]} of ${total} items`,
            }
      }
      bordered
      rowClassName={() => 'table-striped'}
      size="middle"
      scroll={{ x: 'max-content' }}
      onChange={handleTableChange}
    />
  );
};

export default TableChart;
