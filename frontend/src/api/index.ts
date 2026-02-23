import axios, { AxiosInstance, AxiosResponse } from 'axios';
import * as Sentry from '@sentry/react';
import type { QueryConfig } from '../store';

// Types
export type DatasourceType = 'postgresql' | 'clickhouse' | 'mysql' | 'starrocks';

export interface Datasource {
  id: number;
  name: string;
  type: DatasourceType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
  created_at?: string;
  updated_at?: string;
}

export interface DatasourceFormData {
  name: string;
  type: DatasourceType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

// 数据集模式
export type DatasetMode = 'direct' | 'accelerated';

// 数据类型
export type DataType = 'int' | 'float' | 'decimal' | 'string' | 'date' | 'datetime' | 'array' | 'dict' | 'boolean';

import type { StandardDataType, TypeConfig } from './datatypes';

// 列角色
export type ColumnRole = 'dimension' | 'metric';

export interface DatasetColumn {
  name: string;
  expr: string;
  type: StandardDataType;
  typeConfig?: TypeConfig;
  comment: string;
  role: ColumnRole;
}

export interface Dataset {
  id: number;
  name: string;
  datasource_id: number;
  table_name: string | null;
  query_sql: string | null;
  query_type: string;
  mode: DatasetMode;
  accelerate_config?: string;
  description?: string;
  tags?: string;
  refresh_strategy?: string;
  preview_data?: string;
  quality_rules?: string;
  columns: string;
  created_at?: string;
  updated_at?: string;
}

export interface DatasetFormData {
  name: string;
  datasource_id: number;
  table_name?: string;
  query_sql?: string;
  query_type: string;
  mode?: DatasetMode;
  description?: string;
  tags?: string[];
}

export interface Chart {
  id: number;
  name: string;
  dataset_id: number;
  chart_type: string;
  config: string;
  created_at?: string;
  updated_at?: string;
}

export interface ChartFormData {
  name: string;
  dataset_id: number;
  chart_type: string;
  config: string;
}

export interface Share {
  id: number;
  token: string;
  chart_id: number;
  password?: string;
  expires_at?: string;
  created_at?: string;
}

export interface ShareFormData {
  chart_id: number;
  password?: string;
  expires_at?: string;
}

export interface TestConnectionRequest {
  type: DatasourceType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

export interface TestConnectionResponse {
  status: string;
}

export interface TableInfo {
  Name: string;
  Comment: string;
}

export interface ColumnInfo {
  name: string;
  dataType: string;
  comment: string;
  role: ColumnRole;
  isVirtual: boolean;
  expression: string;
}

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: `http://${window.location.hostname || 'localhost'}:8080`,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status;
    const message = error.response?.data?.message || error.message || 'An error occurred';
    console.error('API Error:', message);
    
    // 上报非 2xx 响应到 Sentry
    if (status && status >= 400) {
      Sentry.captureMessage(`API Error: ${error.response?.config?.method?.toUpperCase()} ${error.response?.config?.url} returned ${status}: ${message}`);
    }
    
    return Promise.reject(error);
  }
);

// Datasources API
export const datasourcesApi = {
  // Get all datasources
  getAll: (): Promise<AxiosResponse<Datasource[]>> => {
    return apiClient.get<Datasource[]>('/api/datasources');
  },

  // Get single datasource
  getById: (id: number): Promise<AxiosResponse<Datasource>> => {
    return apiClient.get<Datasource>(`/api/datasources/${id}`);
  },

  // Create datasource
  create: (data: DatasourceFormData): Promise<AxiosResponse<Datasource>> => {
    return apiClient.post<Datasource>('/api/datasources', data);
  },

  // Update datasource
  update: (id: number, data: DatasourceFormData): Promise<AxiosResponse<Datasource>> => {
    return apiClient.put<Datasource>(`/api/datasources/${id}`, data);
  },

  // Delete datasource
  delete: (id: number): Promise<AxiosResponse<void>> => {
    return apiClient.delete<void>(`/api/datasources/${id}`);
  },

  // Test connection
  testConnection: (data: TestConnectionRequest): Promise<AxiosResponse<TestConnectionResponse>> => {
    return apiClient.post<TestConnectionResponse>('/api/datasources/test', data);
  },

  getTables: (id: number): Promise<AxiosResponse<TableInfo[]>> => {
    return apiClient.get<TableInfo[]>(`/api/datasources/${id}/tables`);
  },

  getTableColumns: (id: number, tableName: string): Promise<AxiosResponse<ColumnInfo[]>> => {
    return apiClient.get<ColumnInfo[]>(`/api/datasources/${id}/tables/${encodeURIComponent(tableName)}/columns`);
  },

  // Get data preview from datasource (before creating dataset)
  getPreview: (id: number, tableName: string, querySQL: string, queryType: string): Promise<AxiosResponse<DatasetPreview>> => {
    return apiClient.post<DatasetPreview>(`/api/datasources/${id}/preview`, {
      table_name: tableName,
      query_sql: querySQL,
      query_type: queryType,
    });
  },

  // Get field distribution
  getFieldDistribution: (id: number, tableName: string, querySQL: string, queryType: string, fieldName: string, limit?: number): Promise<AxiosResponse<FieldDistribution>> => {
    return apiClient.post<FieldDistribution>(`/api/datasources/${id}/field-distribution`, {
      table_name: tableName,
      query_sql: querySQL,
      query_type: queryType,
      field_name: fieldName,
      limit: limit || 20,
    });
  },
};

// 数据预览
export interface DatasetPreview {
  columns: string[];
  data: Record<string, unknown>[];
}

// 字段分布
export interface FieldDistribution {
  field_name: string;
  total_count: number;
  unique_count: number;
  distribution: Array<{
    value: unknown;
    count: number;
    percentage: number;
  }>;
}

// Datasets API
export const datasetsApi = {
  // Get all datasets
  getAll: (): Promise<AxiosResponse<Dataset[]>> => {
    return apiClient.get<Dataset[]>('/api/datasets');
  },

  // Get single dataset
  getById: (id: number): Promise<AxiosResponse<Dataset>> => {
    return apiClient.get<Dataset>(`/api/datasets/${id}`);
  },

  // Create dataset
  create: (data: DatasetFormData): Promise<AxiosResponse<Dataset>> => {
    return apiClient.post<Dataset>('/api/datasets', data);
  },

  // Update dataset
  update: (id: number, data: DatasetFormData): Promise<AxiosResponse<Dataset>> => {
    return apiClient.put<Dataset>(`/api/datasets/${id}`, data);
  },

  // Delete dataset
  delete: (id: number): Promise<AxiosResponse<void>> => {
    return apiClient.delete<void>(`/api/datasets/${id}`);
  },

  // Get columns from dataset
  getColumns: (id: number): Promise<AxiosResponse<DatasetColumn[]>> => {
    return apiClient.get<DatasetColumn[]>(`/api/datasets/${id}/columns`);
  },

  // Update columns
  updateColumns: (id: number, columns: DatasetColumn[]): Promise<AxiosResponse<Dataset>> => {
    return apiClient.post<Dataset>(`/api/datasets/${id}/columns`, columns);
  },

  // Get data preview
  getPreview: (id: number): Promise<AxiosResponse<DatasetPreview>> => {
    return apiClient.get<DatasetPreview>(`/api/datasets/${id}/preview`);
  },
};

// Charts API
export const chartsApi = {
  // Get all charts
  getAll: (): Promise<AxiosResponse<Chart[]>> => {
    return apiClient.get<Chart[]>('/api/charts');
  },

  // Get single chart
  getById: (id: number): Promise<AxiosResponse<Chart>> => {
    return apiClient.get<Chart>(`/api/charts/${id}`);
  },

  // Create chart
  create: (data: ChartFormData): Promise<AxiosResponse<Chart>> => {
    return apiClient.post<Chart>('/api/charts', data);
  },

  // Update chart
  update: (id: number, data: Partial<ChartFormData>): Promise<AxiosResponse<Chart>> => {
    return apiClient.put<Chart>(`/api/charts/${id}`, data);
  },

  // Delete chart
  delete: (id: number): Promise<AxiosResponse<void>> => {
    return apiClient.delete<void>(`/api/charts/${id}`);
  },

  // Get chart data
  getChartData: (id: number): Promise<AxiosResponse<any[]>> => {
    return apiClient.get<any[]>(`/api/charts/${id}/data`);
  },

  // Execute query with config
  executeQuery: (datasetId: number, config: QueryConfig): Promise<AxiosResponse<any[]>> => {
    return apiClient.post<any[]>(`/api/datasets/${datasetId}/query`, config);
  },

    executeChartQuery: (request: ChartQueryRequest): Promise<AxiosResponse<ChartQueryResponse>> => {
    return apiClient.post<ChartQueryResponse>('/api/charts/query', request);
  },
};

export type ChartQueryAggregation = 'sum' | 'avg' | 'count' | 'max' | 'min';

export interface ChartQueryMetric {
  field: string;
  agg: ChartQueryAggregation;
  alias?: string;
}

export interface ChartQueryFilter {
  field: string;
  op: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'in' | 'between' | 'isNull' | 'isNotNull';
  value: any;
  value_end?: any;
  logic: 'and' | 'or';
}

export interface ChartQueryPagination {
  page: number;
  page_size: number;
}

export interface ChartQuerySort {
  field: string;
  order: 'asc' | 'desc';
}

export interface ChartQueryRequest {
  dataset_id: number;
  chart_type: string;
  dims: string[];
  metrics: ChartQueryMetric[];
  filters: ChartQueryFilter[];
  pagination?: ChartQueryPagination;
  sort?: ChartQuerySort;
}

export interface TableResponse {
  columns: string[];
  data: Record<string, any>[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
}

export interface PieResponse {
  data: Array<{
    name: string;
    value: number;
    percentage: number;
  }>;
}

export interface AxisResponse {
  x_axis: string[];
  series: Array<{
    name: string;
    data: any[];
  }>;
}

export interface ScatterResponse {
  data: Array<[number, number]>;
}

export interface GeneratedSQL {
  select_sql?: string;
  count_sql?: string;
}

export type ChartQueryResponse = (TableResponse | PieResponse | AxisResponse | ScatterResponse | any[]) & GeneratedSQL;

// Shares API
export const sharesApi = {
  // Get all shares
  getAll: (): Promise<AxiosResponse<Share[]>> => {
    return apiClient.get<Share[]>('/api/shares');
  },

  // Create share
  create: (data: ShareFormData): Promise<AxiosResponse<Share>> => {
    return apiClient.post<Share>('/api/shares', data);
  },

  // Get share by token
  getByToken: (token: string): Promise<AxiosResponse<Share>> => {
    return apiClient.get<Share>(`/api/shares/${token}`);
  },

  // Verify share password
  verifyPassword: (token: string, password: string): Promise<AxiosResponse<Share>> => {
    return apiClient.post<Share>(`/api/shares/${token}/verify`, { password });
  },
};

export default apiClient;
