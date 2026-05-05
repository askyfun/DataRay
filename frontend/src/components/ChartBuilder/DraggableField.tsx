import { useDraggable } from '@dnd-kit/core';
import { Tag } from 'antd';
import React from 'react';
import type { ChartField } from '@/store';

export interface DraggableFieldProps {
  field: ChartField;
}

export interface FieldDragPreviewProps {
  field: ChartField;
}

/**
 * 根据字段角色和数据类型返回统一标签颜色。
 * 调用场景：侧边栏字段标签与拖拽中的 overlay 预览需要保持一致的视觉语义。
 * 主要逻辑：日期维度使用紫色，普通维度使用蓝色，指标使用绿色。
 */
const getFieldTagColor = (field: ChartField): string => {
  if (field.type === 'dimension') {
    if (field.dataType === 'date' || field.dataType === 'timestamp') {
      return 'purple';
    }
    return 'blue';
  }
  return 'green';
};

/**
 * 渲染拖拽中的字段预览标签。
 * 调用场景：ChartBuilder 的 DragOverlay 需要一个跟随鼠标移动的轻量视觉副本。
 * 主要逻辑：复用字段颜色，并关闭指针事件以免遮挡 drop zone 命中。
 */
export const FieldDragPreview: React.FC<FieldDragPreviewProps> = ({ field }) => {
  return (
    <Tag
      data-testid="drag-overlay-field"
      color={getFieldTagColor(field)}
      style={{
        cursor: 'grabbing',
        marginBottom: '4px',
        pointerEvents: 'none',
        boxShadow: '0 8px 24px rgba(0, 0, 0, 0.18)',
        borderRadius: '6px',
      }}
    >
      {field.name}
    </Tag>
  );
};

/**
 * 渲染侧边栏可拖拽字段标签。
 * 调用场景：图表查询界面的字段列表。
 * 主要逻辑：通过 dnd-kit 注册 draggable 节点，并在拖拽时降低源标签透明度。
 */
const DraggableField: React.FC<DraggableFieldProps> = ({ field }) => {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `field-${field.id}`,
    data: {
      type: 'field',
      field,
      fieldType: field.type,
    },
  });

  return (
    <Tag
      ref={setNodeRef}
      color={getFieldTagColor(field)}
      style={{
        cursor: 'grab',
        opacity: isDragging ? 0.5 : 1,
        marginBottom: '4px',
      }}
      {...listeners}
      {...attributes}
    >
      {field.name}
    </Tag>
  );
};

export default DraggableField;
