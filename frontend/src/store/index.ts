import { create } from 'zustand';
import {
  Datasource,
  DatasourceFormData,
  Dataset,
  DatasetFormData,
  Chart,
  ChartFormData,
  ShareFormData,
  DatasetColumn,
  datasourcesApi,
  datasetsApi,
  chartsApi,
  sharesApi,
  ChartQueryRequest,
  ChartQueryResponse,
  TableResponse,
  PieResponse,
  AxisResponse,
} from '../api';

// Field types for chart builder
export type FieldType = 'dimension' | 'metric';

// Field interface for chart builder
export interface ChartField {
  id: string;
  name: string;
  type: FieldType;
  dataType: string;
  comment?: string;
}

// 字段组 - 支持多维度/多指标
export interface FieldGroup {
  id: string;
  fields: string[];
  alias?: string;
}

// 过滤条件操作符
export type FilterOperator = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'in' | 'between' | 'isNull' | 'isNotNull';

// 过滤条件
export interface FilterCondition {
  id: string;
  field: string;
  operator: FilterOperator;
  value: any;
  valueEnd?: any;
  logic: 'and' | 'or';
}

// 图表配置接口
export interface ChartConfig {
  chartType: 'table' | 'line' | 'bar' | 'pie' | 'area' | 'scatter' | 'pivot';
  xAxisField: string | null;
  yAxisFields: string[];
  title: string;
}

// 查询配置 - 支持多维度组和多指标组
export interface QueryConfig {
  dimensionGroups: FieldGroup[];
  metricGroups: FieldGroup[];
  filters: FilterCondition[];
  sort?: { field: string; order: 'asc' | 'desc' };
  limit?: number;
}

// Store state interface
export interface AppState {
  // Datasources
  datasources: Datasource[];
  datasourcesLoading: boolean;
  datasourcesError: string | null;

  // Datasets
  datasets: Dataset[];
  datasetsLoading: boolean;
  datasetsError: string | null;

  // Charts
  charts: Chart[];
  chartsLoading: boolean;
  chartsError: string | null;

  // Chart Builder State
  chartBuilderFields: ChartField[];
  chartBuilderFieldsLoading: boolean;
  chartBuilderConfig: ChartConfig;
  chartData: any[];
  chartDataLoading: boolean;
  queryConfig: QueryConfig;
  autoQuery: boolean;
  metricAggregations: Record<string, string>;
  metricAliases: Record<string, string>;
  chartQueryResponse: ChartQueryResponse | null;
  tablePagination: { page: number; pageSize: number; total: number };
  tableColumns: string[];

  // Selected items
  selectedDatasourceId: number | null;
  selectedDatasetId: number | null;
  selectedChartId: number | null;

  // Actions - Datasources
  fetchDatasources: () => Promise<void>;
  addDatasource: (data: DatasourceFormData) => Promise<Datasource>;
  updateDatasource: (id: number, data: DatasourceFormData) => Promise<Datasource>;
  deleteDatasource: (id: number) => Promise<void>;
  setSelectedDatasource: (id: number | null) => void;

  // Actions - Datasets
  fetchDatasets: () => Promise<void>;
  addDataset: (data: DatasetFormData) => Promise<Dataset>;
  updateDataset: (id: number, data: DatasetFormData) => Promise<Dataset>;
  deleteDataset: (id: number) => Promise<void>;
  setSelectedDataset: (id: number | null) => void;

  // Actions - Charts
  fetchCharts: () => Promise<void>;
  addChart: (data: ChartFormData) => Promise<Chart>;
  updateChart: (id: number, data: Partial<ChartFormData>) => Promise<Chart>;
  deleteChart: (id: number) => Promise<void>;
  setSelectedChart: (id: number | null) => void;

  // Actions - Chart Builder
  fetchDatasetFields: (datasetId: number) => Promise<void>;
  setChartBuilderConfig: (config: Partial<ChartConfig>) => void;
  fetchChartData: (chartId: number) => Promise<void>;
  fetchQueryData: (datasetId: number, config: QueryConfig) => Promise<void>;
  resetChartBuilder: () => void;
  setQueryConfig: (config: Partial<QueryConfig>) => void;
  addDimensionGroup: (group?: FieldGroup) => void;
  removeDimensionGroup: (id: string) => void;
  addMetricGroup: (group?: FieldGroup) => void;
  removeMetricGroup: (id: string) => void;
  addFilter: (filter?: FilterCondition) => void;
  removeFilter: (id: string) => void;
  updateFilter: (id: string, filter: Partial<FilterCondition>) => void;
  addDimensionField: (field: ChartField) => void;
  removeDimensionField: (fieldId: string) => void;
  addMetricField: (field: ChartField) => void;
  removeMetricField: (fieldId: string) => void;
  setMetricAggregation: (fieldId: string, aggregation: string) => void;
  setMetricAlias: (fieldId: string, alias: string) => void;
  toggleAutoQuery: () => void;
  executeChartQuery: (request: ChartQueryRequest) => Promise<void>;
  setTablePagination: (pagination: { page: number; pageSize: number; total: number }) => void;

  // Actions - Shares
  createShare: (data: ShareFormData) => Promise<string>;

  // Clear errors
  clearError: () => void;
}

// Create store
export const useStore = create<AppState>((set) => ({
  // Initial state - Datasources
  datasources: [],
  datasourcesLoading: false,
  datasourcesError: null,

  // Initial state - Datasets
  datasets: [],
  datasetsLoading: false,
  datasetsError: null,

  // Initial state - Charts
  charts: [],
  chartsLoading: false,
  chartsError: null,

  // Initial state - Chart Builder
  chartBuilderFields: [],
  chartBuilderFieldsLoading: false,
  chartBuilderConfig: {
    chartType: 'table',
    xAxisField: null,
    yAxisFields: [],
    title: 'New Chart',
  },
  chartData: [],
  chartDataLoading: false,
  queryConfig: {
    dimensionGroups: [],
    metricGroups: [],
    filters: [],
    limit: 1000,
  },
  autoQuery: true,
  metricAggregations: {},
  metricAliases: {},
  chartQueryResponse: null,
  tablePagination: { page: 1, pageSize: 10, total: 0 },
  tableColumns: [],

  // Selected items
  selectedDatasourceId: null,
  selectedDatasetId: null,
  selectedChartId: null,

  // Datasources actions
  fetchDatasources: async () => {
    set({ datasourcesLoading: true, datasourcesError: null });
    try {
      const response = await datasourcesApi.getAll();
      set({ datasources: response.data, datasourcesLoading: false });
    } catch (error: any) {
      set({
        datasourcesError: error.response?.data?.message || error.message,
        datasourcesLoading: false,
      });
    }
  },

  addDatasource: async (data: DatasourceFormData) => {
    const response = await datasourcesApi.create(data);
    set((state) => ({
      datasources: [...state.datasources, response.data],
    }));
    return response.data;
  },

  updateDatasource: async (id: number, data: DatasourceFormData) => {
    const response = await datasourcesApi.update(id, data);
    set((state) => ({
      datasources: state.datasources.map((ds) => (ds.id === id ? response.data : ds)),
    }));
    return response.data;
  },

  deleteDatasource: async (id: number) => {
    await datasourcesApi.delete(id);
    set((state) => ({
      datasources: state.datasources.filter((ds) => ds.id !== id),
      selectedDatasourceId:
        state.selectedDatasourceId === id ? null : state.selectedDatasourceId,
    }));
  },

  setSelectedDatasource: (id: number | null) => {
    set({ selectedDatasourceId: id });
  },

  // Datasets actions
  fetchDatasets: async () => {
    set({ datasetsLoading: true, datasetsError: null });
    try {
      const response = await datasetsApi.getAll();
      set({ datasets: response.data, datasetsLoading: false });
    } catch (error: any) {
      set({
        datasetsError: error.response?.data?.message || error.message,
        datasetsLoading: false,
      });
    }
  },

  addDataset: async (data: DatasetFormData) => {
    const response = await datasetsApi.create(data);
    set((state) => ({
      datasets: [...state.datasets, response.data],
    }));
    return response.data;
  },

  updateDataset: async (id: number, data: DatasetFormData) => {
    const response = await datasetsApi.update(id, data);
    set((state) => ({
      datasets: state.datasets.map((ds) => (ds.id === id ? response.data : ds)),
    }));
    return response.data;
  },

  deleteDataset: async (id: number) => {
    await datasetsApi.delete(id);
    set((state) => ({
      datasets: state.datasets.filter((ds) => ds.id !== id),
      selectedDatasetId:
        state.selectedDatasetId === id ? null : state.selectedDatasetId,
    }));
  },

  setSelectedDataset: (id: number | null) => {
    set({ selectedDatasetId: id });
  },

  // Charts actions
  fetchCharts: async () => {
    set({ chartsLoading: true, chartsError: null });
    try {
      const response = await chartsApi.getAll();
      set({ charts: response.data, chartsLoading: false });
    } catch (error: any) {
      set({
        chartsError: error.response?.data?.message || error.message,
        chartsLoading: false,
      });
    }
  },

  addChart: async (data: ChartFormData) => {
    const response = await chartsApi.create(data);
    set((state) => ({
      charts: [...state.charts, response.data],
    }));
    return response.data;
  },

  updateChart: async (id: number, data: Partial<ChartFormData>) => {
    const response = await chartsApi.update(id, data);
    set((state) => ({
      charts: state.charts.map((c) => (c.id === id ? response.data : c)),
    }));
    return response.data;
  },

  deleteChart: async (id: number) => {
    await chartsApi.delete(id);
    set((state) => ({
      charts: state.charts.filter((c) => c.id !== id),
      selectedChartId:
        state.selectedChartId === id ? null : state.selectedChartId,
    }));
  },

  setSelectedChart: (id: number | null) => {
    set({ selectedChartId: id });
  },

  // Shares actions
  createShare: async (data: ShareFormData) => {
    const response = await sharesApi.create(data);
    return response.data.token;
  },

  // Chart Builder actions
  fetchDatasetFields: async (datasetId: number) => {
    set({ chartBuilderFieldsLoading: true });
    try {
      const response = await datasetsApi.getColumns(datasetId);
      const columns = response.data;

      // 使用后端返回的 role，如果没有则自动推断
      const fields: ChartField[] = columns.map((col: DatasetColumn, index: number) => ({
        id: `field-${index}`,
        name: col.name,
        type: col.role || (['int', 'float', 'decimal', 'numeric', 'double', 'real'].includes(col.type?.toLowerCase() ?? '')
          ? 'metric'
          : 'dimension'),
        dataType: col.type,
        comment: col.comment,
      }));

      set({ chartBuilderFields: fields, chartBuilderFieldsLoading: false });
    } catch (error: any) {
      set({ chartBuilderFields: [], chartBuilderFieldsLoading: false });
    }
  },

  setChartBuilderConfig: (config: Partial<ChartConfig>) => {
    set((state) => ({
      chartBuilderConfig: { ...state.chartBuilderConfig, ...config },
    }));
  },

  fetchChartData: async (chartId: number) => {
    set({ chartDataLoading: true });
    try {
      const response = await chartsApi.getChartData(chartId);
      set({ chartData: response.data, chartDataLoading: false });
    } catch (error: any) {
      set({ chartData: [], chartDataLoading: false });
    }
  },

  resetChartBuilder: () => {
    set({
      chartBuilderFields: [],
      chartBuilderConfig: {
        chartType: 'table',
        xAxisField: null,
        yAxisFields: [],
        title: 'New Chart',
      },
      chartData: [],
      queryConfig: {
        dimensionGroups: [],
        metricGroups: [],
        filters: [],
        limit: 1000,
      },
    });
  },

  setQueryConfig: (config: Partial<QueryConfig>) => {
    set((state) => ({
      queryConfig: { ...state.queryConfig, ...config },
    }));
  },

  addDimensionGroup: (group?: FieldGroup) => {
    set((state) => {
      const newGroup = group || {
        id: `dim-group-${Date.now()}`,
        fields: [],
      };
      return {
        queryConfig: {
          ...state.queryConfig,
          dimensionGroups: [...state.queryConfig.dimensionGroups, newGroup],
        },
      };
    });
  },

  removeDimensionGroup: (id: string) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        dimensionGroups: state.queryConfig.dimensionGroups.filter((g) => g.id !== id),
      },
    }));
  },

  addMetricGroup: (group?: FieldGroup) => {
    set((state) => {
      const newGroup = group || {
        id: `metric-group-${Date.now()}`,
        fields: [],
      };
      return {
        queryConfig: {
          ...state.queryConfig,
          metricGroups: [...state.queryConfig.metricGroups, newGroup],
        },
      };
    });
  },

  removeMetricGroup: (id: string) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        metricGroups: state.queryConfig.metricGroups.filter((g) => g.id !== id),
      },
    }));
  },

  addFilter: (filter?: FilterCondition) => {
    set((state) => {
      const newFilter = filter || {
        id: `filter-${Date.now()}`,
        field: '',
        operator: 'eq',
        value: '',
        logic: 'and',
      };
      return {
        queryConfig: {
          ...state.queryConfig,
          filters: [...state.queryConfig.filters, newFilter],
        },
      };
    });
  },

  removeFilter: (id: string) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        filters: state.queryConfig.filters.filter((f) => f.id !== id),
      },
    }));
  },

  updateFilter: (id: string, filterUpdate: Partial<FilterCondition>) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        filters: state.queryConfig.filters.map((f) =>
          f.id === id ? { ...f, ...filterUpdate } : f
        ),
      },
    }));
  },

  addDimensionField: (field: ChartField) => {
    set((state) => {
      const existingFields = state.queryConfig.dimensionGroups[0]?.fields || [];
      if (existingFields.includes(field.id)) return state;
      
      const newGroup = {
        id: 'dim-group-main',
        fields: [...existingFields, field.id],
      };
      
      return {
        queryConfig: {
          ...state.queryConfig,
          dimensionGroups: state.queryConfig.dimensionGroups.length > 0
            ? [{ ...state.queryConfig.dimensionGroups[0], fields: newGroup.fields }]
            : [newGroup],
        },
      };
    });
  },

  removeDimensionField: (fieldId: string) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        dimensionGroups: state.queryConfig.dimensionGroups.map((g) => ({
          ...g,
          fields: g.fields.filter((f) => f !== fieldId),
        })),
      },
    }));
  },

  addMetricField: (field: ChartField) => {
    set((state) => {
      const existingFields = state.queryConfig.metricGroups[0]?.fields || [];
      if (existingFields.includes(field.id)) return state;
      
      const newGroup = {
        id: 'metric-group-main',
        fields: [...existingFields, field.id],
      };
      
      return {
        queryConfig: {
          ...state.queryConfig,
          metricGroups: state.queryConfig.metricGroups.length > 0
            ? [{ ...state.queryConfig.metricGroups[0], fields: newGroup.fields }]
            : [newGroup],
        },
      };
    });
  },

  removeMetricField: (fieldId: string) => {
    set((state) => ({
      queryConfig: {
        ...state.queryConfig,
        metricGroups: state.queryConfig.metricGroups.map((g) => ({
          ...g,
          fields: g.fields.filter((f) => f !== fieldId),
        })),
      },
    }));
  },

  setMetricAggregation: (fieldId: string, aggregation: string) => {
    set((state) => ({
      metricAggregations: { ...state.metricAggregations, [fieldId]: aggregation },
    }));
  },

  setMetricAlias: (fieldId: string, alias: string) => {
    set((state) => ({
      metricAliases: { ...state.metricAliases, [fieldId]: alias },
    }));
  },

  toggleAutoQuery: () => {
    set((state) => ({ autoQuery: !state.autoQuery }));
  },

  fetchQueryData: async (datasetId: number, config: QueryConfig) => {
    set({ chartDataLoading: true });
    try {
      const response = await chartsApi.executeQuery(datasetId, config);
      set({ chartData: response.data, chartDataLoading: false });
    } catch (error: any) {
      set({ chartData: [], chartDataLoading: false });
    }
  },

  executeChartQuery: async (request: ChartQueryRequest) => {
    set({ chartDataLoading: true });
    try {
      const response = await chartsApi.executeChartQuery(request);
      const fullResponse = response.data; // 包含 data + select_sql + count_sql
      const responseData = fullResponse.data;

      if (request.chart_type === 'table' && responseData && 'pagination' in responseData) {
        const tableData = responseData as TableResponse;
        set({
          chartData: tableData.data,
          chartQueryResponse: fullResponse as ChartQueryResponse,
          tablePagination: {
            page: tableData.pagination.page,
            pageSize: tableData.pagination.page_size,
            total: tableData.pagination.total,
          },
          tableColumns: tableData.columns || [],
          chartDataLoading: false,
        });
      } else if (request.chart_type === 'pie' && responseData && 'data' in responseData) {
        const pieData = responseData as PieResponse;
        const transformedData = pieData.data.map((item: any) => ({
          name: item.name,
          value: item.value,
        }));
        set({
          chartData: transformedData,
          chartQueryResponse: fullResponse as ChartQueryResponse,
          chartDataLoading: false,
        });
      } else if (request.chart_type === 'bar' || request.chart_type === 'line' || request.chart_type === 'area') {
        const axisData = responseData as AxisResponse;
        const transformedData = axisData.x_axis.map((xVal: string, idx: number) => {
          const row: Record<string, any> = { [request.dims[0] || 'x']: xVal };
          axisData.series.forEach((series: any) => {
            row[series.name] = series.data[idx];
          });
          return row;
        });
        set({
          chartData: transformedData,
          chartQueryResponse: fullResponse as ChartQueryResponse,
          chartDataLoading: false,
        });
      } else {
        set({
          chartData: Array.isArray(responseData) ? responseData : [],
          chartQueryResponse: responseData,
          chartDataLoading: false,
        });
      }
    } catch (error: any) {
      console.error('Chart query failed:', error);
      set({ chartData: [], chartQueryResponse: null, chartDataLoading: false });
    }
  },

  setTablePagination: (pagination: { page: number; pageSize: number; total: number }) => {
    set({ tablePagination: pagination });
  },

  // Clear errors
  clearError: () => {
    set({
      datasourcesError: null,
      datasetsError: null,
      chartsError: null,
    });
  },
}));

export default useStore;
