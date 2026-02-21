import React, { useState } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { PlusOutlined } from '@ant-design/icons';
import { Dropdown, Button } from 'antd';
import FieldPill from './FieldPill';
import type { ChartField } from '@/store';

export type DropZoneType = 'dimension' | 'metric' | 'filter';

export interface FieldDropZoneProps {
  zoneType: DropZoneType;
  label: string;
  fields: ChartField[];
  /** 所有可用字段列表，用于点击选择 */
  availableFields?: ChartField[];
  aggregations?: Record<string, string>;
  aliases?: Record<string, string>;
  onRemoveField?: (fieldId: string) => void;
  onAggregationChange?: (fieldId: string, aggregation: string) => void;
  onOpenSettings?: (field: ChartField) => void;
  onAddField?: (field: ChartField) => void;
  emptyText?: string;
}

const FieldDropZone: React.FC<FieldDropZoneProps> = ({
  zoneType,
  fields,
  availableFields = [],
  aggregations = {},
  aliases = {},
  onRemoveField,
  onAggregationChange,
  onOpenSettings,
  onAddField,
  emptyText,
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id: `dropzone-${zoneType}`,
    data: { type: zoneType },
  });

  const [dropdownOpen, setDropdownOpen] = useState(false);

  const defaultEmptyText = {
    dimension: '拖拽维度字段到此，或点击+添加',
    metric: '拖拽指标字段到此，或点击+添加',
    filter: '拖拽字段添加筛选',
  };

  const filteredFields = availableFields.filter(f => {
    if (zoneType === 'dimension') return f.type === 'dimension';
    if (zoneType === 'metric') return f.type === 'metric';
    return true;
  }).filter(f => !fields.some(added => added.id === f.id));

  const dropdownItems = filteredFields.map(field => ({
    key: field.id,
    label: (
      <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <span style={{
          width: 6,
          height: 6,
          borderRadius: '50%',
          backgroundColor: field.type === 'dimension' ? '#1890ff' : '#52c41a'
        }} />
        {field.name}
      </span>
    ),
    onClick: () => {
      onAddField?.(field);
      setDropdownOpen(false);
    },
  }));

  const getZoneColor = () => {
    switch (zoneType) {
      case 'dimension':
        return isOver ? '#e6f4ff' : '#fafafa';
      case 'metric':
        return isOver ? '#f6ffed' : '#fafafa';
      case 'filter':
        return isOver ? '#fff7e6' : '#fafafa';
      default:
        return '#fafafa';
    }
  };

  const getBorderColor = () => {
    if (!isOver) return '#d9d9d9';
    switch (zoneType) {
      case 'dimension':
        return '#1890ff';
      case 'metric':
        return '#52c41a';
      case 'filter':
        return '#fa8c16';
      default:
        return '#d9d9d9';
    }
  };

  return (
    <div
      ref={setNodeRef}
      style={{
        minHeight: '40px',
        padding: '6px 10px',
        backgroundColor: getZoneColor(),
        border: `1px dashed ${getBorderColor()}`,
        borderRadius: '6px',
        transition: 'all 0.2s ease',
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        flexWrap: 'wrap',
      }}
    >
      {fields.length === 0 ? (
        filteredFields.length > 0 ? (
          <Dropdown
            menu={{ items: dropdownItems }}
            trigger={['click']}
            open={dropdownOpen}
            onOpenChange={setDropdownOpen}
            placement="bottomLeft"
          >
            <Button
              type="dashed"
              size="small"
              icon={<PlusOutlined />}
              onClick={(e) => e.preventDefault()}
              style={{ border: 'none', padding: '4px 8px', height: 'auto' }}
            />
          </Dropdown>
        ) : (
          <span style={{ color: '#999', fontSize: '13px' }}>
            {emptyText || defaultEmptyText[zoneType]}
          </span>
        )
      ) : (
        <>
          {fields.map((field) => (
            <FieldPill
              key={field.id}
              field={field}
              fieldType={zoneType === 'filter' ? 'dimension' : zoneType}
              aggregation={(aggregations[field.id] || 'sum') as 'sum' | 'avg' | 'count' | 'max' | 'min' | 'none'}
              alias={aliases[field.id]}
              onRemove={() => onRemoveField?.(field.id)}
              onAggregationChange={(agg) => onAggregationChange?.(field.id, agg)}
              onSettings={() => onOpenSettings?.(field)}
            />
          ))}
          {filteredFields.length > 0 && (
            <Dropdown
              menu={{ items: dropdownItems }}
              trigger={['click']}
              placement="bottomLeft"
            >
              <Button
                type="dashed"
                size="small"
                icon={<PlusOutlined />}
                onClick={(e) => e.preventDefault()}
                style={{ border: 'none', padding: '4px 8px', height: 'auto', minWidth: 24 }}
              />
            </Dropdown>
          )}
        </>
      )}
    </div>
  );
};

export default FieldDropZone;
