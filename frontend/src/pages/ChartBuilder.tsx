import { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import ReactECharts from 'echarts-for-react';
import { DndContext, DragEndEvent, closestCenter, PointerSensor, useSensor, useSensors } from '@dnd-kit/core';
import {
  Layout,
  Card,
  Typography,
  Select,
  Button,
  Space,
  Empty,
  Spin,
  Divider,
  message,
  Switch,
  Input,
  Modal,
} from 'antd';
import {
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
  TableOutlined,
  AreaChartOutlined,
  DotChartOutlined,
  AppstoreOutlined,
  SaveOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import { useStore, ChartField, ChartConfig } from '../store';
import { ChartQueryAggregation, ChartQueryRequest } from '../api';
import FilterBuilder from '../components/ChartBuilder/FilterBuilder';
import TableChart from '../components/ChartBuilder/TableChart';
import DraggableField from '../components/ChartBuilder/DraggableField';
import QueryConfigRow from '../components/ChartBuilder/QueryConfigRow';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

interface ChartCanvasProps {
  config: ChartConfig;
  data: any[];
  loading: boolean;
}

const ChartCanvas: React.FC<ChartCanvasProps> = ({ config, data, loading }) => {
  const getChartOption = useCallback(() => {
    const { queryConfig } = useStore.getState();
    const dimensionFields = queryConfig.dimensionGroups.flatMap(g => g.fields);
    const metricFields = queryConfig.metricGroups.flatMap(g => g.fields);
    
    if (dimensionFields.length === 0 || metricFields.length === 0 || data.length === 0) {
      return null;
    }

    const xAxisField = dimensionFields[0];
    const xAxisData = data.map((item) => item[xAxisField]);

    const commonOptions = {
      title: {
        text: config.title,
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
    };

    switch (config.chartType) {
      case 'line':
        return {
          ...commonOptions,
          xAxis: {
            type: 'category',
            data: xAxisData,
          },
          yAxis: {
            type: 'value',
          },
          series: metricFields.map((yField) => ({
            name: yField,
            type: 'line',
            data: data.map((item) => item[yField]),
          })),
        };

      case 'bar':
        return {
          ...commonOptions,
          xAxis: {
            type: 'category',
            data: xAxisData,
          },
          yAxis: {
            type: 'value',
          },
          series: metricFields.map((yField) => ({
            name: yField,
            type: 'bar',
            data: data.map((item) => item[yField]),
          })),
        };

      case 'pie': {
        const pieData = data.map((item) => ({
          name: item[xAxisField] || item.name,
          value: item[metricFields[0] || ''] || item.value,
        }));
        return {
          ...commonOptions,
          tooltip: {
            trigger: 'item',
            formatter: '{b}: {c} ({d}%)',
          },
          legend: {
            orient: 'vertical',
            left: 'left',
          },
          series: [
            {
              name: metricFields[0] || 'Value',
              type: 'pie',
              radius: '50%',
              data: pieData,
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)',
                },
              },
              label: {
                formatter: '{b}: {d}%',
              },
            },
          ],
        };
      }

      case 'area':
        return {
          ...commonOptions,
          xAxis: {
            type: 'category',
            data: xAxisData,
          },
          yAxis: {
            type: 'value',
          },
          series: metricFields.map((yField) => ({
            name: yField,
            type: 'line',
            areaStyle: {},
            data: data.map((item) => item[yField]),
          })),
        };

      case 'scatter':
        return {
          ...commonOptions,
          series: [
            {
              type: 'scatter',
              data: data.map((item) => [item[xAxisField], item[metricFields[0] || '']]),
            },
          ],
        };

      default:
        return null;
    }
  }, [config, data]);

  const chartOption = getChartOption();

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>
          <Text type="secondary">加载图表数据中...</Text>
        </div>
      </div>
    );
  }

  if (!chartOption) {
    return (
      <Empty
        description="请配置维度和指标以生成图表"
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        style={{ padding: '100px 0' }}
      />
    );
  }

  return (
    <ReactECharts
      option={chartOption}
      style={{ height: '100%', width: '100%' }}
      opts={{ renderer: 'canvas' }}
    />
  );
};

interface FieldListPanelProps {
  fields: ChartField[];
  loading: boolean;
}

const FieldListPanel: React.FC<FieldListPanelProps> = ({ fields, loading }) => {
  const dimensions = fields.filter((f) => f.type === 'dimension');
  const metrics = fields.filter((f) => f.type === 'metric');

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '40px 0' }}>
        <Spin />
        <div style={{ marginTop: 8 }}>
          <Text type="secondary">加载字段中...</Text>
        </div>
      </div>
    );
  }

  if (fields.length === 0) {
    return (
      <Empty
        description="请先选择数据集"
        image={Empty.PRESENTED_IMAGE_SIMPLE}
      />
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <Text strong type="secondary" style={{ display: 'block', marginBottom: 6 }}>
          维度 ({dimensions.length})
        </Text>
        {dimensions.map((field) => (
          <DraggableField key={field.id} field={field} />
        ))}
      </div>

      <Divider style={{ margin: '8px 0' }} />

      <div>
        <Text strong type="secondary" style={{ display: 'block', marginBottom: 6 }}>
          指标 ({metrics.length})
        </Text>
        {metrics.map((field) => (
          <DraggableField key={field.id} field={field} />
        ))}
      </div>
    </div>
  );
};

interface ConfigPanelProps {
  config: ChartConfig;
  onConfigChange: (config: Partial<ChartConfig>) => void;
}

const chartTypeOptions = [
  { type: 'table', icon: <TableOutlined />, label: '表格' },
  { type: 'bar', icon: <BarChartOutlined />, label: '柱状图' },
  { type: 'line', icon: <LineChartOutlined />, label: '折线图' },
  { type: 'pie', icon: <PieChartOutlined />, label: '饼图' },
  { type: 'area', icon: <AreaChartOutlined />, label: '面积图' },
  { type: 'scatter', icon: <DotChartOutlined />, label: '散点图' },
  { type: 'pivot', icon: <AppstoreOutlined />, label: '透视表' },
] as const;

const ConfigPanel: React.FC<ConfigPanelProps> = ({
  config,
  onConfigChange,
}) => {
  return (
    <div>
      <Card title="可视化类型" size="small" style={{ marginBottom: 12 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8 }}>
          {chartTypeOptions.map((opt) => (
            <Button
              key={opt.type}
              type={config.chartType === opt.type ? 'primary' : 'default'}
              icon={opt.icon}
              onClick={() => onConfigChange({ chartType: opt.type })}
              style={{ height: 40 }}
            >
              {opt.label}
            </Button>
          ))}
        </div>
      </Card>

      <Card title="图表配置" size="small" style={{ marginBottom: 12 }}>
        <Space direction="vertical" style={{ width: '100%' }} size="small">
          <div>
            <Text strong>图表标题</Text>
            <Input
              style={{ width: '100%', marginTop: 4 }}
              value={config.title}
              onChange={(e) => onConfigChange({ title: e.target.value })}
              placeholder="输入图表标题"
            />
          </div>
        </Space>
      </Card>

      <Card title="当前配置" size="small" style={{ marginBottom: 12 }}>
        <div style={{ fontSize: 12 }}>
          <div style={{ marginBottom: 6 }}>
            <Text type="secondary">类型: </Text>
            <Text strong>{config.chartType.toUpperCase()}</Text>
          </div>
          <div style={{ marginBottom: 6 }}>
            <Text type="secondary">标题: </Text>
            <Text strong>{config.title}</Text>
          </div>
        </div>
      </Card>
    </div>
  );
};

const ChartBuilder: React.FC = () => {
  const [searchParams] = useSearchParams();
  const [selectedDatasetId, setSelectedDatasetId] = useState<number | null>(null);
  const [editingChartId, setEditingChartId] = useState<number | null>(null);
  const [sqlModalVisible, setSqlModalVisible] = useState(false);

  const {
    datasets,
    fetchDatasets,
    chartBuilderFields,
    chartBuilderFieldsLoading,
    chartBuilderConfig,
    chartData,
    chartDataLoading,
    setChartBuilderConfig,
    fetchDatasetFields,
    fetchQueryData,
    resetChartBuilder,
    addChart,
    updateChart,
    queryConfig,
    setQueryConfig,
    addFilter,
    removeFilter,
    updateFilter,
    addDimensionField,
    removeDimensionField,
    addMetricField,
    removeMetricField,
    setMetricAggregation,
    setMetricAlias,
    autoQuery,
    toggleAutoQuery,
    metricAggregations,
    metricAliases,
    executeChartQuery,
    tablePagination,
    tableColumns,
    setTablePagination,
    chartQueryResponse,
  } = useStore();

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 5,
      },
    })
  );

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;
    
    if (!over) return;

    const fieldData = active.data.current;
    const overData = over.data.current;

    if (!fieldData || fieldData.type !== 'field') return;

    const field = fieldData.field as ChartField;
    const dropZoneType = overData?.type as 'dimension' | 'metric' | 'filter';

    if (dropZoneType === 'dimension') {
      if (field.type === 'dimension') {
        addDimensionField(field);
      } else {
        message.warning('请将指标拖入指标区域');
      }
    } else if (dropZoneType === 'metric') {
      if (field.type === 'metric') {
        addMetricField(field);
      } else {
        message.warning('请将维度拖入维度区域');
      }
    } else if (dropZoneType === 'filter') {
      addFilter({
        id: `filter-${Date.now()}`,
        field: field.id,
        operator: 'eq',
        value: '',
        logic: 'and',
      });
    }
  }, [addDimensionField, addMetricField, addFilter]);

  const getDimensionFields = useCallback(() => {
    const dimensionIds = queryConfig.dimensionGroups.flatMap(g => g.fields);
    return chartBuilderFields.filter(f => dimensionIds.includes(f.id));
  }, [queryConfig.dimensionGroups, chartBuilderFields]);

  const getMetricFields = useCallback(() => {
    const metricIds = queryConfig.metricGroups.flatMap(g => g.fields);
    return chartBuilderFields.filter(f => metricIds.includes(f.id));
  }, [queryConfig.metricGroups, chartBuilderFields]);

  const buildChartQueryRequest = useCallback((): ChartQueryRequest | null => {
    if (!selectedDatasetId) return null;

    const dimensionFields = getDimensionFields();
    const metricFields = getMetricFields();

    if (dimensionFields.length === 0 && metricFields.length === 0) {
      return null;
    }

    const dims = dimensionFields.map(f => f.name);
    const metrics = metricFields.map(f => ({
      field: f.name,
      agg: (metricAggregations[f.id] || 'sum') as ChartQueryAggregation,
      alias: metricAliases[f.id] || f.name,
    }));

    const filters = queryConfig.filters.map(f => {
      const field = chartBuilderFields.find(field => field.id === f.field);
      return {
        field: field?.name || f.field,
        op: f.operator as any,
        value: f.value,
        value_end: (f as any).valueEnd,
        logic: f.logic,
      };
    });

    return {
      dataset_id: selectedDatasetId,
      chart_type: chartBuilderConfig.chartType,
      dims,
      metrics,
      filters,
      pagination: chartBuilderConfig.chartType === 'table' ? {
        page: tablePagination.page,
        page_size: tablePagination.pageSize,
      } : undefined,
    };
  }, [selectedDatasetId, getDimensionFields, getMetricFields, metricAggregations, metricAliases, queryConfig.filters, chartBuilderConfig.chartType, tablePagination, chartBuilderFields]);

  const handleExecuteQuery = useCallback(() => {
    const request = buildChartQueryRequest();
    if (request) {
      executeChartQuery(request);
    }
  }, [buildChartQueryRequest, executeChartQuery]);

  const handlePageChange = useCallback((page: number, pageSize: number) => {
    setTablePagination({ ...tablePagination, page, pageSize });
    const request = buildChartQueryRequest();
    if (request) {
      request.pagination = { page, page_size: pageSize };
      executeChartQuery(request);
    }
  }, [buildChartQueryRequest, executeChartQuery, setTablePagination, tablePagination]);

  useEffect(() => {
    fetchDatasets();
  }, [fetchDatasets]);

  useEffect(() => {
    const editId = searchParams.get('edit');
    const datasetIdParam = searchParams.get('datasetId');

    if (editId && datasetIdParam) {
      const chartId = parseInt(editId, 10);
      const dsId = parseInt(datasetIdParam, 10);

      if (!isNaN(chartId) && !isNaN(dsId)) {
        setSelectedDatasetId(dsId);
        setEditingChartId(chartId);
      }
    }
  }, [searchParams]);

  useEffect(() => {
    if (selectedDatasetId) {
      fetchDatasetFields(selectedDatasetId);
    } else {
      resetChartBuilder();
    }
  }, [selectedDatasetId, fetchDatasetFields, resetChartBuilder]);

  useEffect(() => {
    if (selectedDatasetId && autoQuery) {
      const request = buildChartQueryRequest();
      if (request) {
        executeChartQuery(request);
      }
    }
  }, [chartBuilderConfig.chartType, selectedDatasetId, autoQuery, queryConfig, queryConfig.dimensionGroups, queryConfig.metricGroups, queryConfig.filters]);

  useEffect(() => {
    if (selectedDatasetId && queryConfig) {
      const hasDimensions = queryConfig.dimensionGroups.some(g => g.fields.length > 0);
      const hasMetrics = queryConfig.metricGroups.some(g => g.fields.length > 0);
      
      if (hasDimensions || hasMetrics) {
        if (!autoQuery) {
          const request = buildChartQueryRequest();
          if (request) {
            executeChartQuery(request);
          }
        }
      }
    }
  }, [selectedDatasetId, queryConfig, autoQuery]);

  useEffect(() => {
    const loadChartConfig = async () => {
      if (editingChartId && selectedDatasetId) {
        try {
          const { chartsApi } = await import('../api');
          const response = await chartsApi.getById(editingChartId);
          const chart = response.data;

          try {
            const config = JSON.parse(chart.config);
            setChartBuilderConfig({
              chartType: config.chartType || 'table',
              title: config.title || chart.name,
            });
            
            if (config.queryConfig) {
              setQueryConfig(config.queryConfig);
            }
          } catch (e) {
            setChartBuilderConfig({
              chartType: chart.chart_type as 'table' | 'line' | 'bar' | 'pie' | 'area' | 'scatter' | 'pivot',
              title: chart.name,
            });
          }
        } catch (error) {
          console.error('Failed to load chart config:', error);
        }
      }
    };

    loadChartConfig();
  }, [editingChartId, selectedDatasetId, setChartBuilderConfig, setQueryConfig]);

  const handleSave = async () => {
    if (!selectedDatasetId) {
      message.error('请先选择数据集');
      return;
    }

    try {
      const configJson = JSON.stringify({
        ...chartBuilderConfig,
        queryConfig,
      });

      if (editingChartId) {
        await updateChart(editingChartId, {
          name: chartBuilderConfig.title,
          dataset_id: selectedDatasetId,
          chart_type: chartBuilderConfig.chartType,
          config: configJson,
        });
        message.success('图表更新成功');
      } else {
        const newChart = await addChart({
          name: chartBuilderConfig.title,
          dataset_id: selectedDatasetId,
          chart_type: chartBuilderConfig.chartType,
          config: configJson,
        });
        message.success('图表保存成功');
        setEditingChartId(newChart.id);
      }
    } catch (error: any) {
      message.error(error.response?.data?.message || '保存失败');
    }
  };

  const handleReset = () => {
    resetChartBuilder();
    setSelectedDatasetId(null);
    setEditingChartId(null);
    message.info('已重置');
  };

  const handleDatasetChange = (value: number | null) => {
    setSelectedDatasetId(value);
    setEditingChartId(null);
  };

  const renderPreview = () => {
    if (chartBuilderConfig.chartType === 'table' || chartBuilderConfig.chartType === 'pivot') {
      return (
        <TableChart 
          data={chartData} 
          loading={chartDataLoading} 
          queryConfig={queryConfig}
          columns={tableColumns}
          pagination={chartBuilderConfig.chartType === 'table' ? tablePagination : undefined}
          onPageChange={chartBuilderConfig.chartType === 'table' ? handlePageChange : undefined}
        />
      );
    }
    return <ChartCanvas config={chartBuilderConfig} data={chartData} loading={chartDataLoading} />;
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <Layout style={{ minHeight: 'calc(100vh - 120px)' }}>
        <Header
          style={{
            background: '#fff',
            padding: '0 16px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Space>
            <Text strong style={{ fontSize: 16 }}>数据集:</Text>
            <Select
              style={{ width: 240 }}
              placeholder="选择数据集"
              value={selectedDatasetId}
              onChange={handleDatasetChange}
              allowClear
              options={datasets.map((ds) => ({
                value: ds.id,
                label: ds.name,
              }))}
            />
          </Space>
          <Space>
            <Space>
              <Text type="secondary">自动查询</Text>
              <Switch checked={autoQuery} onChange={toggleAutoQuery} size="small" />
            </Space>
            {!autoQuery && (
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleExecuteQuery}
                disabled={!selectedDatasetId}
              >
                执行查询
              </Button>
            )}
            {chartData.length > 0 && (
              <Button
                icon={<CodeOutlined />}
                onClick={() => setSqlModalVisible(true)}
              >
                查看 SQL
              </Button>
            )}
            <Button
              type="primary"
              icon={<SaveOutlined />}
              onClick={handleSave}
              disabled={!selectedDatasetId}
            >
              {editingChartId ? '更新' : '保存'}
            </Button>
            <Button icon={<ReloadOutlined />} onClick={handleReset}>
              重置
            </Button>
          </Space>
        </Header>

        <Layout>
          <Sider
            width={180}
            style={{ background: '#fff', padding: '12px', borderRight: '1px solid #f0f0f0' }}
          >
            <Card title="可用字段" size="small">
              <FieldListPanel fields={chartBuilderFields} loading={chartBuilderFieldsLoading} />
            </Card>
          </Sider>

          <Content style={{ padding: '12px', background: '#fafafa', display: 'flex', flexDirection: 'column', gap: 12 }}>
            <Card title="查询配置" size="small" style={{ flex: '0 0 auto' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <QueryConfigRow
                  rowType="dimension"
                  fields={getDimensionFields()}
                  availableFields={chartBuilderFields}
                  onRemoveField={removeDimensionField}
                  onAggregationChange={setMetricAggregation}
                  onAddField={(field) => addDimensionField(field)}
                />

                <QueryConfigRow
                  rowType="metric"
                  fields={getMetricFields()}
                  availableFields={chartBuilderFields}
                  aggregations={metricAggregations}
                  aliases={metricAliases}
                  onRemoveField={removeMetricField}
                  onAggregationChange={setMetricAggregation}
                  onAddField={(field) => addMetricField(field)}
                  onOpenSettings={(field) => {
                    const alias = prompt('输入字段别名:', field.name);
                    if (alias !== null) {
                      setMetricAlias(field.id, alias);
                    }
                  }}
                />

                <FilterBuilder
                  fields={chartBuilderFields}
                  filters={queryConfig.filters}
                  onAdd={() => addFilter()}
                  onRemove={removeFilter}
                  onUpdate={updateFilter}
                />
              </div>
            </Card>

            <Card title="预览" size="small" style={{ flex: 1, minHeight: 400 }}>
              <div style={{ height: 'calc(100vh - 480px)', minHeight: 300 }}>
                {renderPreview()}
              </div>
            </Card>
          </Content>

          <Sider
            width={260}
            style={{ background: '#fff', padding: '12px', borderLeft: '1px solid #f0f0f0' }}
          >
            <ConfigPanel
              config={chartBuilderConfig}
              onConfigChange={setChartBuilderConfig}
            />
          </Sider>
        </Layout>
      </Layout>
      <Modal
        title="生成的 SQL"
        open={sqlModalVisible}
        onCancel={() => setSqlModalVisible(false)}
        footer={null}
        width={800}
      >
        {chartQueryResponse && (
          <div>
            <Text strong>数据查询:</Text>
            <pre style={{ 
              background: '#f5f5f5', 
              padding: 12, 
              borderRadius: 4, 
              overflow: 'auto',
              maxHeight: 300,
              fontSize: 12
            }}>
              {chartQueryResponse.select_sql || '无'}
            </pre>
            {chartQueryResponse.count_sql && (
              <>
                <Text strong style={{ marginTop: 16, display: 'block' }}>计数查询:</Text>
                <pre style={{ 
                  background: '#f5f5f5', 
                  padding: 12, 
                  borderRadius: 4, 
                  overflow: 'auto',
                  maxHeight: 200,
                  fontSize: 12
                }}>
                  {chartQueryResponse.count_sql}
                </pre>
              </>
            )}
          </div>
        )}
      </Modal>
    </DndContext>
  );
};

export default ChartBuilder;
