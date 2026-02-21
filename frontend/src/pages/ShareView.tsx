import { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import {
  Card,
  Typography,
  Spin,
  Input,
  Button,
  Space,
  Result,
  Tag,
} from 'antd';
import {
  LockOutlined,
  ShareAltOutlined,
  BarChartOutlined,
  LineChartOutlined,
  PieChartOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { sharesApi, chartsApi, Chart } from '../api';

const { Title, Text } = Typography;

interface ShareInfo {
  id: number;
  token: string;
  chart_id: number;
  password?: string;
  expires_at?: string;
  created_at?: string;
}

interface ChartConfig {
  chartType: 'line' | 'bar' | 'pie';
  xAxisField: string | null;
  yAxisFields: string[];
  title: string;
}

const ShareView: React.FC = () => {
  const { token } = useParams<{ token: string }>();
  const [loading, setLoading] = useState(true);
  const [authLoading, setAuthLoading] = useState(false);
  const [shareInfo, setShareInfo] = useState<ShareInfo | null>(null);
  const [chart, setChart] = useState<Chart | null>(null);
  const [chartData, setChartData] = useState<any[]>([]);
  const [chartDataLoading, setChartDataLoading] = useState(false);
  const [needsPassword, setNeedsPassword] = useState(false);
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [password, setPassword] = useState('');

  // Fetch share info on mount
  useEffect(() => {
    if (token) {
      fetchShareInfo();
    }
  }, [token]);

  // Fetch share info
  const fetchShareInfo = async () => {
    setLoading(true);
    try {
      const response = await sharesApi.getByToken(token!);
      const share = response.data;

      // Check if share has expired
      if (share.expires_at && new Date(share.expires_at) < new Date()) {
        setShareInfo(share);
        setNeedsPassword(false);
        return;
      }

      setShareInfo(share);

      // Check if password is required
      if (share.password) {
        setNeedsPassword(true);
        setLoading(false);
        return;
      }

      // No password required, fetch chart directly
      await fetchChart(share.chart_id);
    } catch (error: any) {
      if (error.response?.status === 401 || error.response?.status === 403) {
        // Password required
        setNeedsPassword(true);
        setShareInfo(error.response?.data?.share || { id: 0, token: token!, chart_id: 0 });
      } else {
        // Other error
        console.error('Failed to fetch share:', error);
      }
    } finally {
      setLoading(false);
    }
  };

  // Verify password and fetch chart
  const handleVerifyPassword = async () => {
    if (!password) {
      setPasswordError('Please enter the password');
      return;
    }

    setAuthLoading(true);
    setPasswordError(null);

    try {
      const response = await sharesApi.verifyPassword(token!, password);
      setNeedsPassword(false);
      await fetchChart(response.data.chart_id);
    } catch (error: any) {
      if (error.response?.status === 401 || error.response?.status === 403) {
        setPasswordError('Invalid password');
      } else {
        setPasswordError(error.response?.data?.message || 'Verification failed');
      }
    } finally {
      setAuthLoading(false);
    }
  };

  // Fetch chart data
  const fetchChart = async (chartId: number) => {
    setChartDataLoading(true);
    try {
      const [chartResponse, dataResponse] = await Promise.all([
        chartsApi.getById(chartId),
        chartsApi.getChartData(chartId),
      ]);

      setChart(chartResponse.data);
      setChartData(dataResponse.data);
    } catch (error: any) {
      console.error('Failed to fetch chart:', error);
    } finally {
      setChartDataLoading(false);
    }
  };

  // Generate chart option
  const getChartOption = useCallback(() => {
    if (!chart || chartData.length === 0) {
      return null;
    }

    let config: ChartConfig;
    try {
      config = JSON.parse(chart.config);
    } catch {
      config = {
        chartType: chart.chart_type as 'line' | 'bar' | 'pie',
        xAxisField: null,
        yAxisFields: [],
        title: chart.name,
      };
    }

    if (!config.xAxisField || config.yAxisFields.length === 0) {
      return null;
    }

    const xAxisData = chartData.map((item) => item[config.xAxisField!]);

    const commonOptions = {
      title: {
        text: config.title || chart.name,
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
    };

    switch (config.chartType) {
      case 'line':
        return {
          ...commonOptions,
          xAxis: {
            type: 'category',
            data: xAxisData,
          },
          yAxis: {
            type: 'value',
          },
          series: config.yAxisFields.map((yField) => ({
            name: yField,
            type: 'line',
            data: chartData.map((item) => item[yField]),
          })),
        };

      case 'bar':
        return {
          ...commonOptions,
          xAxis: {
            type: 'category',
            data: xAxisData,
          },
          yAxis: {
            type: 'value',
          },
          series: config.yAxisFields.map((yField) => ({
            name: yField,
            type: 'bar',
            data: chartData.map((item) => item[yField]),
          })),
        };

      case 'pie':
        return {
          ...commonOptions,
          series: [
            {
              name: config.yAxisFields[0] || 'Value',
              type: 'pie',
              radius: '50%',
              data: chartData.map((item) => ({
                name: item[config.xAxisField!],
                value: item[config.yAxisFields[0] || ''],
              })),
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)',
                },
              },
            },
          ],
        };

      default:
        return null;
    }
  }, [chart, chartData]);

  // Get chart type icon
  const getChartTypeIcon = (type: string) => {
    switch (type) {
      case 'line':
        return <LineChartOutlined />;
      case 'pie':
        return <PieChartOutlined />;
      case 'bar':
      default:
        return <BarChartOutlined />;
    }
  };

  // Loading state
  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
        <Spin size="large" />
      </div>
    );
  }

  // Expired share
  if (shareInfo?.expires_at && new Date(shareInfo.expires_at) < new Date()) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5', padding: 24 }}>
        <Result
          status="warning"
          title="Share Expired"
          subTitle="This share link has expired and is no longer accessible."
        />
      </div>
    );
  }

  // Password input
  if (needsPassword) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5', padding: 24 }}>
        <Card style={{ width: 400, textAlign: 'center' }}>
          <div style={{ marginBottom: 24 }}>
            <LockOutlined style={{ fontSize: 48, color: '#faad14' }} />
          </div>
          <Title level={4}>Password Required</Title>
          <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
            This chart is password protected. Please enter the password to view.
          </Text>
          <Space direction="vertical" style={{ width: '100%' }}>
            <Input.Password
              placeholder="Enter password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setPasswordError(null);
              }}
              onPressEnter={handleVerifyPassword}
              status={passwordError ? 'error' : undefined}
            />
            {passwordError && (
              <Text type="danger" style={{ fontSize: 12 }}>
                {passwordError}
              </Text>
            )}
            <Button
              type="primary"
              loading={authLoading}
              onClick={handleVerifyPassword}
              block
            >
              View Chart
            </Button>
          </Space>
        </Card>
      </div>
    );
  }

  // Chart display
  const chartOption = getChartOption();

  return (
    <div style={{ minHeight: '100vh', background: '#f0f2f5', padding: 24 }}>
      <Card>
        {/* Header */}
        <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            {chart && getChartTypeIcon(chart.chart_type)}
            <Title level={4} style={{ margin: 0 }}>
              {chart?.name || 'Shared Chart'}
            </Title>
          </Space>
          <Space>
            <Tag icon={<ShareAltOutlined />} color="blue">
              Shared View
            </Tag>
          </Space>
        </div>

        {/* Chart */}
        {chartDataLoading ? (
          <div style={{ textAlign: 'center', padding: '100px 0' }}>
            <Spin size="large" />
            <div style={{ marginTop: 16 }}>
              <Text type="secondary">Loading chart data...</Text>
            </div>
          </div>
        ) : chartOption ? (
          <div style={{ height: 'calc(100vh - 250px)', minHeight: 400 }}>
            <ReactECharts
              option={chartOption}
              style={{ height: '100%', width: '100%' }}
              opts={{ renderer: 'canvas' }}
            />
          </div>
        ) : (
          <Result
            status="warning"
            title="Unable to display chart"
            subTitle="The chart configuration may be invalid or no data is available."
          />
        )}
      </Card>
    </div>
  );
};

export default ShareView;
