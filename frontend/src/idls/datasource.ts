export type DatasourceType = 'postgresql' | 'mysql' | 'clickhouse' | 'starrocks';

export interface CreateDatasourceRequest {
  name: string;
  type: DatasourceType;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
}

export type UpdateDatasourceRequest = CreateDatasourceRequest;

export interface DatasourceResponse {
  id: number;
  name: string;
  type: string;
  host: string;
  port: number;
  database_name: string;
  username: string;
  password: string;
  created_at?: string;
  updated_at?: string;
}

export interface DatasourceListResponse {
  items: DatasourceResponse[];
  total: number;
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
  name: string;
  comment: string;
}

export interface ColumnInfo {
  name: string;
  data_type: string;
  comment: string;
  role: string;
  is_virtual: boolean;
  expression: string;
}

export interface DatasetPreview {
  columns: string[];
  data: Record<string, unknown>[];
}

export interface FieldDistributionValue {
  value: unknown;
  count: number;
  percentage: number;
}

export interface FieldDistribution {
  field_name: string;
  total_count: number;
  unique_count: number;
  distribution: FieldDistributionValue[];
}
