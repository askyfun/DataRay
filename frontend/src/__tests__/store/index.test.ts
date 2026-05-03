import { describe, expect, it } from 'vitest';
import { useStore } from '@/store';

describe('Store', () => {
  it('should have initial state', () => {
    const store = useStore.getState();
    expect(store.datasources).toEqual([]);
    expect(store.datasets).toEqual([]);
    expect(store.charts).toEqual([]);
    expect(store.datasourcesLoading).toBe(false);
    expect(store.datasetsLoading).toBe(false);
    expect(store.chartsLoading).toBe(false);
  });

  it('should have action functions', () => {
    const store = useStore.getState();
    expect(typeof store.fetchDatasources).toBe('function');
    expect(typeof store.fetchDatasets).toBe('function');
    expect(typeof store.fetchCharts).toBe('function');
  });
});
