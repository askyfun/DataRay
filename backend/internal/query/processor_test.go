package query

import (
	"testing"
)

// TestAxisProcessor_EmptyData 验证空行数据返回空响应
func TestAxisProcessor_EmptyData(t *testing.T) {
	p := &AxisProcessor{}
	resp, err := p.Process([]map[string]any{}, []string{"category"}, []MetricConfig{{Field: "value", Agg: AggSum}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.XAxis) != 0 {
		t.Errorf("expected empty XAxis, got %v", axisResp.XAxis)
	}
	if len(axisResp.Series) != 0 {
		t.Errorf("expected empty Series, got %v", axisResp.Series)
	}
}

// TestAxisProcessor_EmptyDims 验证空维度返回空响应
func TestAxisProcessor_EmptyDims(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{{"category": "A", "value": 100}}
	resp, err := p.Process(rows, []string{}, []MetricConfig{{Field: "value", Agg: AggSum}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.XAxis) != 0 {
		t.Errorf("expected empty XAxis, got %v", axisResp.XAxis)
	}
	if len(axisResp.Series) != 0 {
		t.Errorf("expected empty Series, got %v", axisResp.Series)
	}
}

// TestAxisProcessor_EmptyMetrics 验证空指标返回空响应
func TestAxisProcessor_EmptyMetrics(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{{"category": "A", "value": 100}}
	resp, err := p.Process(rows, []string{"category"}, []MetricConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.XAxis) != 0 {
		t.Errorf("expected empty XAxis, got %v", axisResp.XAxis)
	}
	if len(axisResp.Series) != 0 {
		t.Errorf("expected empty Series, got %v", axisResp.Series)
	}
}

// TestAxisProcessor_SingleDim_SingleMetric 验证单维度单指标的基本输出
func TestAxisProcessor_SingleDim_SingleMetric(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{
		{"category": "A", "value": 100},
		{"category": "B", "value": 200},
	}
	metrics := []MetricConfig{{Field: "value", Agg: AggSum}}
	resp, err := p.Process(rows, []string{"category"}, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.XAxis) != 2 {
		t.Fatalf("expected XAxis length 2, got %d", len(axisResp.XAxis))
	}
	if axisResp.XAxis[0] != "A" {
		t.Errorf("expected XAxis[0]='A', got %q", axisResp.XAxis[0])
	}
	if axisResp.XAxis[1] != "B" {
		t.Errorf("expected XAxis[1]='B', got %q", axisResp.XAxis[1])
	}

	if len(axisResp.Series) != 1 {
		t.Fatalf("expected Series length 1, got %d", len(axisResp.Series))
	}
	// 无 alias，ResolveAlias 返回 Field
	if axisResp.Series[0].Name != "value" {
		t.Errorf("expected series name 'value', got %q", axisResp.Series[0].Name)
	}
	if len(axisResp.Series[0].Data) != 2 {
		t.Fatalf("expected series data length 2, got %d", len(axisResp.Series[0].Data))
	}
}

// TestAxisProcessor_SingleDim_MultipleMetrics 验证多指标输出
func TestAxisProcessor_SingleDim_MultipleMetrics(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{
		{"category": "A", "revenue": 100, "cost": 50},
		{"category": "B", "revenue": 200, "cost": 80},
	}
	metrics := []MetricConfig{
		{Field: "revenue", Agg: AggSum, Alias: "total_revenue"},
		{Field: "cost", Agg: AggSum, Alias: "total_cost"},
	}
	resp, err := p.Process(rows, []string{"category"}, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.Series) != 2 {
		t.Fatalf("expected Series length 2, got %d", len(axisResp.Series))
	}
	if axisResp.Series[0].Name != "total_revenue" {
		t.Errorf("expected series[0] name 'total_revenue', got %q", axisResp.Series[0].Name)
	}
	if axisResp.Series[1].Name != "total_cost" {
		t.Errorf("expected series[1] name 'total_cost', got %q", axisResp.Series[1].Name)
	}
}

// TestAxisProcessor_TypeConversion 验证不同类型指标值保持原始类型
func TestAxisProcessor_TypeConversion(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{
		{"category": "A", "value": float64(100.5)},
		{"category": "B", "value": int64(200)},
		{"category": "C", "value": int(300)},
	}
	metrics := []MetricConfig{{Field: "value", Agg: AggSum}}
	resp, err := p.Process(rows, []string{"category"}, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	data := axisResp.Series[0].Data
	if len(data) != 3 {
		t.Fatalf("expected data length 3, got %d", len(data))
	}

	// 值保持原始类型，不做转换
	if _, ok := data[0].(float64); !ok {
		t.Errorf("expected data[0] type float64, got %T", data[0])
	}
	if _, ok := data[1].(int64); !ok {
		t.Errorf("expected data[1] type int64, got %T", data[1])
	}
	if _, ok := data[2].(int); !ok {
		t.Errorf("expected data[2] type int, got %T", data[2])
	}
}

// TestAxisProcessor_WithAlias 验证指标别名作为 series name
func TestAxisProcessor_WithAlias(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{
		{"category": "A", "total_revenue": 500},
	}
	metrics := []MetricConfig{{Field: "revenue", Agg: AggSum, Alias: "total_revenue"}}
	resp, err := p.Process(rows, []string{"category"}, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.Series) != 1 {
		t.Fatalf("expected Series length 1, got %d", len(axisResp.Series))
	}
	if axisResp.Series[0].Name != "total_revenue" {
		t.Errorf("expected series name 'total_revenue', got %q", axisResp.Series[0].Name)
	}
}

// TestAxisProcessor_NullDimension 验证维度值为 nil 时转为空字符串
func TestAxisProcessor_NullDimension(t *testing.T) {
	p := &AxisProcessor{}
	rows := []map[string]any{
		{"category": nil, "value": 100},
		{"category": "B", "value": 200},
	}
	metrics := []MetricConfig{{Field: "value", Agg: AggSum}}
	resp, err := p.Process(rows, []string{"category"}, metrics)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	axisResp, ok := resp.(*AxisResponse)
	if !ok {
		t.Fatalf("expected *AxisResponse, got %T", resp)
	}

	if len(axisResp.XAxis) != 2 {
		t.Fatalf("expected XAxis length 2, got %d", len(axisResp.XAxis))
	}
	if axisResp.XAxis[0] != "" {
		t.Errorf("expected XAxis[0]='' for nil dim, got %q", axisResp.XAxis[0])
	}
	if axisResp.XAxis[1] != "B" {
		t.Errorf("expected XAxis[1]='B', got %q", axisResp.XAxis[1])
	}
}

// TestGetProcessor_BarLineArea 验证 GetProcessor 返回正确的处理器类型
func TestGetProcessor_BarLineArea(t *testing.T) {
	cases := []struct {
		name      string
		chartType ChartType
	}{
		{"bar", ChartTypeBar},
		{"line", ChartTypeLine},
		{"area", ChartTypeArea},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			processor := GetProcessor(tc.chartType)
			if processor == nil {
				t.Fatal("expected non-nil processor")
			}
			if _, ok := processor.(*AxisProcessor); !ok {
				t.Fatalf("expected *AxisProcessor, got %T", processor)
			}
		})
	}
}
