import React from 'react';
import FieldDropZone, { DropZoneType } from './FieldDropZone';
import type { ChartField } from '@/store';

export interface QueryConfigRowProps {
  /** 行类型 */
  rowType: DropZoneType;
  /** 当前字段列表 */
  fields: ChartField[];
  /** 所有可用字段列表 */
  availableFields?: ChartField[];
  /** 指标聚合方式映射 */
  aggregations?: Record<string, string>;
  /** 字段别名映射 */
  aliases?: Record<string, string>;
  /** 删除字段回调 */
  onRemoveField?: (fieldId: string) => void;
  /** 聚合方式变更回调 */
  onAggregationChange?: (fieldId: string, aggregation: string) => void;
  /** 打开设置回调 */
  onOpenSettings?: (field: ChartField) => void;
  /** 添加字段回调 */
  onAddField?: (field: ChartField) => void;
}

const ROW_CONFIG: Record<DropZoneType, { label: string; color: string }> = {
  dimension: { label: '维度', color: '#1890ff' },
  metric: { label: '指标', color: '#52c41a' },
  filter: { label: '筛选', color: '#fa8c16' },
};

const QueryConfigRow: React.FC<QueryConfigRowProps> = ({
  rowType,
  fields,
  availableFields,
  aggregations,
  aliases,
  onRemoveField,
  onAggregationChange,
  onOpenSettings,
  onAddField,
}) => {
  const config = ROW_CONFIG[rowType];

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-start',
        marginBottom: '8px',
      }}
    >
      <div
        style={{
          width: '48px',
          minWidth: '48px',
          fontWeight: 500,
          color: config.color,
          paddingTop: '10px',
          fontSize: '13px',
        }}
      >
        {config.label}
      </div>
      <div style={{ flex: 1 }}>
        <FieldDropZone
          zoneType={rowType}
          label={config.label}
          fields={fields}
          availableFields={availableFields}
          aggregations={aggregations}
          aliases={aliases}
          onRemoveField={onRemoveField}
          onAggregationChange={onAggregationChange}
          onOpenSettings={onOpenSettings}
          onAddField={onAddField}
        />
      </div>
    </div>
  );
};

export default QueryConfigRow;
