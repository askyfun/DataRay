import { Select, Input, Button, Space } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import type { FilterCondition, FilterOperator, ChartField } from '../../store';

interface FilterBuilderProps {
  fields: ChartField[];
  filters: FilterCondition[];
  onAdd: () => void;
  onRemove: (id: string) => void;
  onUpdate: (id: string, filter: Partial<FilterCondition>) => void;
}

const operatorOptions: { value: FilterOperator; label: string }[] = [
  { value: 'eq', label: '=' },
  { value: 'neq', label: '!=' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '>=' },
  { value: 'lt', label: '<' },
  { value: 'lte', label: '<=' },
  { value: 'like', label: 'LIKE' },
  { value: 'in', label: 'IN' },
  { value: 'between', label: 'BETWEEN' },
  { value: 'isNull', label: 'IS NULL' },
  { value: 'isNotNull', label: 'IS NOT NULL' },
];

const needsValue = (operator: FilterOperator): boolean => {
  return !['isNull', 'isNotNull'].includes(operator);
};

const needsTwoValues = (operator: FilterOperator): boolean => {
  return operator === 'between';
};

const needsMultiValues = (operator: FilterOperator): boolean => {
  return operator === 'in';
};

const FilterBuilder: React.FC<FilterBuilderProps> = ({
  fields,
  filters,
  onAdd,
  onRemove,
  onUpdate,
}) => {
  const fieldOptions = fields.map((f) => ({
    value: f.name,
    label: f.name,
  }));

  const handleFieldChange = (id: string, field: string) => {
    onUpdate(id, { field });
  };

  const handleOperatorChange = (id: string, operator: FilterOperator) => {
    onUpdate(id, { operator, value: '', valueEnd: undefined });
  };

  const handleValueChange = (id: string, value: any) => {
    onUpdate(id, { value });
  };

  const handleValueEndChange = (id: string, valueEnd: any) => {
    onUpdate(id, { valueEnd });
  };

  const handleLogicChange = (id: string, logic: 'and' | 'or') => {
    onUpdate(id, { logic });
  };

  return (
    <div>
      {filters.map((filter, index) => (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            marginBottom: 12,
            gap: 8,
          }}
        >
          {index > 0 ? (
            <Select
              value={filter.logic}
              onChange={(value) => handleLogicChange(filter.id, value)}
              style={{ width: 70 }}
              options={[
                { value: 'and', label: 'AND' },
                { value: 'or', label: 'OR' },
              ]}
            />
          ) : (
            <div style={{ width: 70 }} />
          )}

          <Select
            style={{ flex: 1, minWidth: 120 }}
            placeholder="Select field"
            value={filter.field || undefined}
            onChange={(value) => handleFieldChange(filter.id, value)}
            options={fieldOptions}
            allowClear
          />

          <Select
            style={{ width: 100 }}
            value={filter.operator}
            onChange={(value) => handleOperatorChange(filter.id, value)}
            options={operatorOptions}
          />

          {needsValue(filter.operator) && (
            <>
              {needsTwoValues(filter.operator) ? (
                <Space>
                  <Input
                    style={{ width: 100 }}
                    placeholder="Min"
                    value={filter.value ?? ''}
                    onChange={(e) => handleValueChange(filter.id, e.target.value)}
                  />
                  <span>-</span>
                  <Input
                    style={{ width: 100 }}
                    placeholder="Max"
                    value={filter.valueEnd ?? ''}
                    onChange={(e) => handleValueEndChange(filter.id, e.target.value)}
                  />
                </Space>
              ) : needsMultiValues(filter.operator) ? (
                <Input
                  style={{ flex: 1, minWidth: 150 }}
                  placeholder="value1, value2, value3"
                  value={filter.value ?? ''}
                  onChange={(e) => handleValueChange(filter.id, e.target.value)}
                />
              ) : (
                <Input
                  style={{ flex: 1, minWidth: 120 }}
                  placeholder="Value"
                  value={filter.value ?? ''}
                  onChange={(e) => handleValueChange(filter.id, e.target.value)}
                />
              )}
            </>
          )}

          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => onRemove(filter.id)}
          />
        </div>
      ))}
      <Button
        type="dashed"
        icon={<PlusOutlined />}
        onClick={onAdd}
        style={{ width: '100%' }}
      >
        Add Filter
      </Button>
    </div>
  );
};

export default FilterBuilder;
