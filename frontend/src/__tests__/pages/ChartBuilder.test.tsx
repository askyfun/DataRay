import { render, waitFor } from '@testing-library/react';
import type { AxiosResponse } from 'axios';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { chartsApi, datasetsApi } from '../../api';
import type { ApiResponse } from '../../lib/api/client';
import ChartBuilder from '../../pages/ChartBuilder';
import { useStore } from '../../store';

function mockAxiosResponse<T>(data: ApiResponse<T>): AxiosResponse<ApiResponse<T>> {
  return { data, status: 200, statusText: 'OK', headers: {}, config: {} as never };
}

vi.mock('echarts-for-react', () => ({
  default: () => <div data-testid="echarts" />,
}));

vi.mock('../../components/ChartBuilder/DraggableField', () => ({
  default: () => <div data-testid="draggable-field" />,
}));

vi.mock('../../components/ChartBuilder/FilterBuilder', () => ({
  default: () => <div data-testid="filter-builder" />,
}));

vi.mock('../../components/ChartBuilder/QueryConfigRow', () => ({
  default: () => <div data-testid="query-config-row" />,
}));

vi.mock('../../api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../api')>();
  return {
    ...actual,
    datasetsApi: {
      ...actual.datasetsApi,
      getAll: vi.fn(),
      getColumns: vi.fn(),
    },
    chartsApi: {
      ...actual.chartsApi,
      getById: vi.fn(),
      executeChartQuery: vi.fn(),
    },
  };
});

const mockGetDatasets = vi.mocked(datasetsApi.getAll);
const mockGetColumns = vi.mocked(datasetsApi.getColumns);
const mockGetChartById = vi.mocked(chartsApi.getById);
const mockExecuteChartQuery = vi.mocked(chartsApi.executeChartQuery);

const resetChartBuilderState = () => {
  useStore.setState({
    datasets: [],
    datasetsLoading: false,
    datasetsError: null,
    chartBuilderFields: [],
    chartBuilderFieldsLoading: false,
    chartBuilderConfig: {
      chartType: 'table',
      xAxisField: null,
      yAxisFields: [],
      title: 'New Chart',
    },
    chartData: [],
    chartDataLoading: false,
    queryConfig: {
      dimensionGroups: [{ id: 'dim-group-main', fields: ['field-0'] }],
      metricGroups: [],
      filters: [],
      limit: 1000,
    },
    autoQuery: true,
    metricAggregations: {},
    metricAliases: {},
    chartQueryResponse: null,
    tablePagination: { page: 1, pageSize: 10, total: 0 },
    tableColumns: [],
  });
};

const renderChartBuilder = () => {
  return render(
    <MemoryRouter initialEntries={['/chart-builder?edit=1&datasetId=1']}>
      <Routes>
        <Route path="/chart-builder" element={<ChartBuilder />} />
      </Routes>
    </MemoryRouter>
  );
};

describe('ChartBuilder', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetChartBuilderState();

    mockGetDatasets.mockResolvedValue(
      mockAxiosResponse({
        code: 20000,
        msg: 'ok',
        trace: '',
        data: [
          {
            id: 1,
            name: 'Sales',
            datasource_id: 1,
            table_name: 'sales',
            query_sql: null,
            query_type: 'table',
            mode: 'direct',
            columns: '[]',
          },
        ],
      })
    );

    mockGetColumns.mockResolvedValue(
      mockAxiosResponse({
        code: 20000,
        msg: 'ok',
        trace: '',
        data: [
          {
            name: 'region',
            expr: 'region',
            type: 'string',
            comment: '',
            role: 'dimension',
          },
        ],
      })
    );

    mockGetChartById.mockResolvedValue(
      mockAxiosResponse({
        code: 20000,
        msg: 'ok',
        trace: '',
        data: {
          id: 1,
          name: 'Sales Table',
          dataset_id: 1,
          chart_type: 'table',
          config: JSON.stringify({
            chartType: 'table',
            title: 'Sales Table',
            queryConfig: {
              dimensionGroups: [{ id: 'dim-group-main', fields: ['field-0'] }],
              metricGroups: [],
              filters: [],
              limit: 1000,
            },
          }),
        },
      })
    );

    mockExecuteChartQuery.mockResolvedValue(
      mockAxiosResponse({
        code: 20000,
        msg: 'ok',
        trace: '',
        data: {
          data: {
            columns: ['region'],
            data: [{ region: 'East' }],
            pagination: {
              page: 1,
              page_size: 10,
              total: 42,
              total_pages: 5,
            },
          },
          select_sql: 'select region from sales',
          count_sql: 'select count(*) from sales',
        },
      })
    );
  });

  it('does not repeat table auto query when the response only updates total pagination', async () => {
    renderChartBuilder();

    await waitFor(() => {
      expect(mockExecuteChartQuery).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(useStore.getState().tablePagination.total).toBe(42);
    });

    await new Promise((resolve) => setTimeout(resolve, 50));

    expect(mockExecuteChartQuery).toHaveBeenCalledTimes(1);
    expect(mockExecuteChartQuery).toHaveBeenCalledWith(
      expect.objectContaining({
        chart_type: 'table',
        pagination: { page: 1, page_size: 10 },
      })
    );
  });
});
