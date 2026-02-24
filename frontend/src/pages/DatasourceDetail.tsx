import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useIntl } from 'react-intl';
import {
  Card,
  Typography,
  Table,
  Tag,
  Button,
  Space,
  Spin,
  message,
  Breadcrumb,
} from 'antd';
import {
  DatabaseOutlined,
  TableOutlined,
  ColumnWidthOutlined,
  ArrowLeftOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useStore } from '../store';
import { datasourcesApi } from '../api';
import type { TableInfo, ColumnInfo } from '../api';

const { Title, Text } = Typography;

const getTypeInfo = (type: string) => {
  const typeMap: Record<string, { label: string; color: string }> = {
    postgresql: { label: 'PostgreSQL', color: 'blue' },
    mysql: { label: 'MySQL', color: 'green' },
    clickhouse: { label: 'ClickHouse', color: 'purple' },
    starrocks: { label: 'StarRocks', color: 'orange' },
  };
  return typeMap[type] || { label: type, color: 'blue' };
};

const DatasourceDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const intl = useIntl();
  const datasourceId = Number(id);

  const { datasources, fetchDatasources } = useStore();

  const [tables, setTables] = useState<TableInfo[]>([]);
  const [tablesLoading, setTablesLoading] = useState(false);
  const [columnsCache, setColumnsCache] = useState<Record<string, ColumnInfo[]>>({});
  const [loadingColumns, setLoadingColumns] = useState<Record<string, boolean>>({});

  const datasource = datasources.find((ds) => ds.id === datasourceId);

  useEffect(() => {
    if (datasources.length === 0) {
      fetchDatasources();
    }
  }, [datasources.length, fetchDatasources]);

  const loadTables = async () => {
    setTablesLoading(true);
    try {
      const response = await datasourcesApi.getTables(datasourceId);
      setTables(response.data);
    } catch (error: any) {
      message.error(error.response?.data?.message || intl.formatMessage({ id: 'common.failedToLoad' }));
    } finally {
      setTablesLoading(false);
    }
  };

  useEffect(() => {
    if (datasourceId) {
      loadTables();
    }
  }, [datasourceId]);

  const loadColumns = async (tableName: string) => {
    if (columnsCache[tableName]) {
      return;
    }

    setLoadingColumns((prev) => ({ ...prev, [tableName]: true }));
    try {
      const response = await datasourcesApi.getTableColumns(datasourceId, tableName);
      setColumnsCache((prev) => ({ ...prev, [tableName]: response.data }));
    } catch (error: any) {
      message.error(error.response?.data?.message || `Failed to load columns for ${tableName}`);
    } finally {
      setLoadingColumns((prev) => ({ ...prev, [tableName]: false }));
    }
  };

  const columnsColumns = [
    {
      title: 'Name',
      dataIndex: 'Name',
      key: 'Name',
      width: 200,
      render: (text: string) => (
        <Space>
          <ColumnWidthOutlined />
          <Text strong>{text}</Text>
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'Type',
      key: 'Type',
      width: 150,
      render: (text: string) => <Tag color="blue">{text}</Tag>,
    },
    {
      title: 'Comment',
      dataIndex: 'Comment',
      key: 'Comment',
    },
  ];

  const tableColumns = [
    {
      title: 'Table Name',
      dataIndex: 'Name',
      key: 'Name',
      render: (text: string) => (
        <Space>
          <TableOutlined />
          <Text strong>{text}</Text>
        </Space>
      ),
    },
    {
      title: 'Comment',
      dataIndex: 'Comment',
      key: 'Comment',
    },
  ];

  if (!datasource) {
    return (
      <div style={{ padding: '24px' }}>
        <Spin tip="Loading datasource..." />
      </div>
    );
  }

  return (
    <div style={{ padding: '24px' }}>
      <Card>
        <Breadcrumb
          style={{ marginBottom: '16px' }}
          items={[
            {
              title: <Link to="/datasources">Data Sources</Link>,
            },
            {
              title: datasource.name,
            },
          ]}
        />

        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '24px',
          }}
        >
          <div>
            <Space>
              <DatabaseOutlined style={{ fontSize: '24px', color: '#1890ff' }} />
              <Title level={3} style={{ margin: 0 }}>
                {datasource.name}
              </Title>
              <Tag color={getTypeInfo(datasource.type).color}>
                {getTypeInfo(datasource.type).label}
              </Tag>
            </Space>
            <Text type="secondary">
              {datasource.host}:{datasource.port} / {datasource.database_name}
            </Text>
          </div>
          <Space>
            <Button
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/datasources')}
            >
              Back
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={loadTables}
              loading={tablesLoading}
            >
              Refresh Tables
            </Button>
          </Space>
        </div>

        <div style={{ marginBottom: '24px' }}>
          <Text strong style={{ marginBottom: '8px', display: 'block' }}>
            Tables ({tables.length})
          </Text>
          <Table
            columns={tableColumns}
            dataSource={tables}
            rowKey="Name"
            loading={tablesLoading}
            expandable={{
              expandedRowRender: (record: TableInfo) => {
                const tableColumns = columnsCache[record.Name] || [];
                const isLoading = loadingColumns[record.Name];

                if (isLoading) {
                  return <Spin tip="Loading columns..." />;
                }

                return (
                  <Table
                    columns={columnsColumns}
                    dataSource={tableColumns}
                    rowKey="name"
                    size="small"
                    pagination={false}
                    locale={{ emptyText: 'No columns found' }}
                  />
                );
              },
              onExpand: (expanded: boolean, record: TableInfo) => {
                if (expanded) {
                  loadColumns(record.Name);
                }
              },
            }}
            pagination={{
              pageSize: 10,
              showSizeChanger: true,
            }}
            locale={{
              emptyText: 'No tables found. Make sure the datasource connection is valid.',
            }}
            size="middle"
          />
        </div>
      </Card>
    </div>
  );
};

export default DatasourceDetailPage;
