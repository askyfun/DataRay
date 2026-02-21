export type DatasetMode = 'direct' | 'accelerated';
export type QueryType = 'table' | 'sql';
export type ColumnRole = 'dimension' | 'metric';

export interface CreateDatasetRequest {
  name: string;
  datasource_id: number;
  table_name?: string;
  query_sql?: string;
  query_type: QueryType;
  mode?: DatasetMode;
  description?: string;
  tags?: string[];
}

export type UpdateDatasetRequest = CreateDatasetRequest;

export interface DatasetColumn {
  name: string;
  expr: string;
  type: string;
  type_config?: Record<string, unknown>;
  comment: string;
  role: ColumnRole;
}

export interface DatasetResponse {
  id: number;
  name: string;
  datasource_id: number;
  table_name?: string;
  query_sql?: string;
  query_type: string;
  mode: string;
  accelerate_config?: Record<string, unknown>;
  description?: string;
  tags: string[];
  refresh_strategy?: Record<string, unknown>;
  preview_data?: Record<string, unknown>;
  quality_rules?: Record<string, unknown>;
  columns: string;
  created_at?: string;
  updated_at?: string;
}

export interface DatasetListResponse {
  items: DatasetResponse[];
  total: number;
}

export interface UpdateColumnsRequest {
  columns: DatasetColumn[];
}

export interface DatasetPreviewResponse {
  columns: string[];
  data: Record<string, unknown>[];
}
