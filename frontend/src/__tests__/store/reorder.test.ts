import { beforeEach, describe, expect, it } from 'vitest';
import { useStore } from '../../store';

describe('reorderDimensionField', () => {
  beforeEach(() => {
    const store = useStore.getState();
    store.resetChartBuilder();
    // 模拟已有字段
    useStore.setState({
      chartBuilderFields: [
        { id: 'field-0', name: 'name', type: 'dimension', dataType: 'string' },
        { id: 'field-1', name: 'city', type: 'dimension', dataType: 'string' },
        { id: 'field-2', name: 'date', type: 'dimension', dataType: 'date' },
      ],
    });
  });

  it('addDimensionField 按顺序追加', () => {
    const { addDimensionField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addDimensionField(fields[0]);
    addDimensionField(fields[1]);

    const dims = useStore.getState().queryConfig.dimensionGroups[0]?.fields;
    expect(dims).toEqual(['field-0', 'field-1']);
  });

  it('reorderDimensionField 交换两个字段', () => {
    const { addDimensionField, reorderDimensionField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addDimensionField(fields[0]);
    addDimensionField(fields[1]);

    // 初始顺序: [field-0, field-1]
    expect(useStore.getState().queryConfig.dimensionGroups[0].fields).toEqual([
      'field-0',
      'field-1',
    ]);

    // 交换: [field-1, field-0]
    reorderDimensionField(0, 1);
    expect(useStore.getState().queryConfig.dimensionGroups[0].fields).toEqual([
      'field-1',
      'field-0',
    ]);
  });

  it('reorderDimensionField 移动到末尾', () => {
    const { addDimensionField, reorderDimensionField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addDimensionField(fields[0]);
    addDimensionField(fields[1]);
    addDimensionField(fields[2]);

    // 初始: [field-0, field-1, field-2]
    reorderDimensionField(0, 2);
    // 期望: [field-1, field-2, field-0]
    expect(useStore.getState().queryConfig.dimensionGroups[0].fields).toEqual([
      'field-1',
      'field-2',
      'field-0',
    ]);
  });

  it('reorderDimensionField 移动到开头', () => {
    const { addDimensionField, reorderDimensionField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addDimensionField(fields[0]);
    addDimensionField(fields[1]);
    addDimensionField(fields[2]);

    // 初始: [field-0, field-1, field-2]
    reorderDimensionField(2, 0);
    // 期望: [field-2, field-0, field-1]
    expect(useStore.getState().queryConfig.dimensionGroups[0].fields).toEqual([
      'field-2',
      'field-0',
      'field-1',
    ]);
  });
});
