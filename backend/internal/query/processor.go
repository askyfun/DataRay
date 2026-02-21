package query

import (
	"fmt"
	"math"
)

// Processor 接口定义
type Processor interface {
	Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error)
}

// TableProcessor Table 图表处理器
type TableProcessor struct {
	Pagination *Pagination
}

// Process 处理 Table 数据
func (p *TableProcessor) Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error) {
	if len(rows) == 0 {
		return &TableResponse{
			Columns:    []string{},
			Data:       []map[string]any{},
			Pagination: TablePagination{},
		}, nil
	}

	// 提取列名
	columns := make([]string, 0)
	for key := range rows[0] {
		columns = append(columns, key)
	}

	// 处理分页
	page := 1
	pageSize := 10
	total := len(rows)

	if p.Pagination != nil {
		page = p.Pagination.Page
		pageSize = p.Pagination.PageSize
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &TableResponse{
		Columns: columns,
		Data:    rows,
		Pagination: TablePagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// PieProcessor Pie 图表处理器
type PieProcessor struct {
	Threshold int // 长尾合并阈值，默认 20
}

// NewPieProcessor 创建 Pie 处理器
func NewPieProcessor() *PieProcessor {
	return &PieProcessor{
		Threshold: 20,
	}
}

// Process 处理 Pie 数据
func (p *PieProcessor) Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error) {
	if len(rows) == 0 || len(dims) == 0 || len(metrics) == 0 {
		return &PieResponse{
			Data: []PieDataItem{},
		}, nil
	}

	dimField := dims[0]
	metricField := metrics[0].ResolveAlias()

	// 计算总值
	var total float64
	dataItems := make([]PieDataItem, 0, len(rows))

	for _, row := range rows {
		dimValue := row[dimField]
		metricValue := row[metricField]

		var value float64
		switch v := metricValue.(type) {
		case float64:
			value = v
		case int64:
			value = float64(v)
		case int:
			value = float64(v)
		default:
			value = 0
		}

		total += value

		name := ""
		if dimValue != nil {
			name = toString(dimValue)
		}

		dataItems = append(dataItems, PieDataItem{
			Name:  name,
			Value: value,
		})
	}

	// 计算百分比
	for i := range dataItems {
		if total > 0 {
			dataItems[i].Percentage = math.Round(dataItems[i].Value*10000/total) / 100
		}
	}

	// 长尾合并
	var otherData []PieDataItem
	threshold := p.Threshold
	if len(dataItems) > threshold {
		// 按值排序
		for i := 0; i < len(dataItems)-1; i++ {
			for j := i + 1; j < len(dataItems); j++ {
				if dataItems[i].Value < dataItems[j].Value {
					dataItems[i], dataItems[j] = dataItems[j], dataItems[i]
				}
			}
		}

		// 保留前 N 个，其余合并
		mainData := dataItems[:threshold]
		otherData = dataItems[threshold:]

		// 更新主数据
		dataItems = mainData
	}

	// 合并 Other
	if len(otherData) > 0 {
		var otherValue float64
		var otherPercentage float64
		for _, item := range otherData {
			otherValue += item.Value
			otherPercentage += item.Percentage
		}

		dataItems = append(dataItems, PieDataItem{
			Name:       "其他",
			Value:      otherValue,
			Percentage: math.Round(otherPercentage*100) / 100,
		})
	}

	return &PieResponse{
		Data: dataItems,
	}, nil
}

// AxisProcessor 坐标轴图表处理器 (Bar, Line, Area)
type AxisProcessor struct{}

// Process 处理坐标轴图表数据
func (p *AxisProcessor) Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error) {
	if len(rows) == 0 || len(dims) == 0 || len(metrics) == 0 {
		return &AxisResponse{
			XAxis:  []string{},
			Series: []AxisSeries{},
		}, nil
	}

	dimField := dims[0]

	// X 轴数据
	xAxis := make([]string, len(rows))
	for i, row := range rows {
		val := row[dimField]
		if val != nil {
			xAxis[i] = toString(val)
		}
	}

	// Series 数据
	series := make([]AxisSeries, len(metrics))
	for j, metric := range metrics {
		metricAlias := metric.ResolveAlias()
		data := make([]any, len(rows))

		for i, row := range rows {
			val := row[metricAlias]
			data[i] = val
		}

		series[j] = AxisSeries{
			Name: metricAlias,
			Data: data,
		}
	}

	return &AxisResponse{
		XAxis:  xAxis,
		Series: series,
	}, nil
}

// ScatterProcessor Scatter 图表处理器
type ScatterProcessor struct{}

// Process 处理 Scatter 数据
func (p *ScatterProcessor) Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error) {
	if len(rows) == 0 || len(metrics) < 2 {
		return &ScatterResponse{
			Data: [][]float64{},
		}, nil
	}

	xField := metrics[0].ResolveAlias()
	yField := metrics[1].ResolveAlias()

	data := make([][]float64, 0, len(rows))

	for _, row := range rows {
		xVal := row[xField]
		yVal := row[yField]

		x, ok1 := toFloat64(xVal)
		y, ok2 := toFloat64(yVal)

		if ok1 && ok2 {
			data = append(data, []float64{x, y})
		}
	}

	return &ScatterResponse{
		Data: data,
	}, nil
}

// PivotProcessor 透视表处理器
type PivotProcessor struct{}

// Process 处理 Pivot 数据
func (p *PivotProcessor) Process(rows []map[string]any, dims []string, metrics []MetricConfig) (ChartQueryResponse, error) {
	if len(rows) == 0 {
		return &PivotResponse{
			Columns: []string{},
			Data:    []map[string]any{},
		}, nil
	}

	// 提取列名
	columns := make([]string, 0)
	for key := range rows[0] {
		columns = append(columns, key)
	}

	return &PivotResponse{
		Columns: columns,
		Data:    rows,
	}, nil
}

// GetProcessor 获取对应的处理器
func GetProcessor(chartType ChartType) Processor {
	switch chartType {
	case ChartTypeTable:
		return &TableProcessor{}
	case ChartTypePie:
		return NewPieProcessor()
	case ChartTypeBar, ChartTypeLine, ChartTypeArea:
		return &AxisProcessor{}
	case ChartTypeScatter:
		return &ScatterProcessor{}
	case ChartTypePivot:
		return &PivotProcessor{}
	default:
		return &AxisProcessor{}
	}
}

// toString 转换为字符串
func toString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toFloat64 转换为 float64
func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}
