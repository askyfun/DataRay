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

  it('addDimensionField 支持追加到指定维度组', () => {
    const { addDimensionGroup, addDimensionField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addDimensionGroup({ id: 'dim-group-main', fields: [] });
    addDimensionGroup({ id: 'dim-group-secondary', fields: [] });

    addDimensionField(fields[0], 0);
    addDimensionField(fields[1], 1);

    expect(useStore.getState().queryConfig.dimensionGroups[0].fields).toEqual(['field-0']);
    expect(useStore.getState().queryConfig.dimensionGroups[1].fields).toEqual(['field-1']);
  });

  it('addMetricField 支持追加到指定指标组', () => {
    useStore.setState({
      chartBuilderFields: [
        { id: 'metric-0', name: 'revenue', type: 'metric', dataType: 'float' },
        { id: 'metric-1', name: 'profit', type: 'metric', dataType: 'float' },
      ],
    });

    const { addMetricGroup, addMetricField } = useStore.getState();
    const fields = useStore.getState().chartBuilderFields;

    addMetricGroup({ id: 'metric-group-main', fields: [] });
    addMetricGroup({ id: 'metric-group-secondary', fields: [] });

    addMetricField(fields[0], 0);
    addMetricField(fields[1], 1);

    expect(useStore.getState().queryConfig.metricGroups[0].fields).toEqual(['metric-0']);
    expect(useStore.getState().queryConfig.metricGroups[1].fields).toEqual(['metric-1']);
  });

  it('addMetricField 写入第 2 个指标组时会自动补齐前置空组，避免稀疏数组', () => {
    useStore.setState({
      chartBuilderFields: [{ id: 'metric-0', name: 'revenue', type: 'metric', dataType: 'float' }],
    });

    const { addMetricField } = useStore.getState();
    const field = useStore.getState().chartBuilderFields[0];

    addMetricField(field, 1);

    expect(useStore.getState().queryConfig.metricGroups).toEqual([
      { id: 'metric-group-1', fields: [] },
      { id: 'metric-group-2', fields: ['metric-0'] },
    ]);
  });
});
