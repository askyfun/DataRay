import React from 'react';
import { useDraggable } from '@dnd-kit/core';
import { Tag } from 'antd';
import type { ChartField } from '@/store';

export interface DraggableFieldProps {
  field: ChartField;
}

const DraggableField: React.FC<DraggableFieldProps> = ({ field }) => {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `field-${field.id}`,
    data: {
      type: 'field',
      field,
      fieldType: field.type,
    },
  });

  const getTagColor = () => {
    if (field.type === 'dimension') {
      if (field.dataType === 'date' || field.dataType === 'timestamp') {
        return 'purple';
      }
      return 'blue';
    }
    return 'green';
  };

  return (
    <Tag
      ref={setNodeRef}
      color={getTagColor()}
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
