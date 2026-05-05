package query

// ChartSpec 图表语义规格，表达一个图表实例绑定了什么字段和配置。
// 属于 Chart 语义层，不直接生成 SQL。
type ChartSpec struct {
	ChartType       ChartType        `json:"chart_type"`
	DimensionGroups []DimensionGroup `json:"dimension_groups"`
	MetricGroups    []MetricGroup    `json:"metric_groups"`
	Style           map[string]any   `json:"style,omitempty"`
	QueryOptions    map[string]any   `json:"query_options,omitempty"`
}

// DimensionGroup 维度组，例如 "x_axis"、"rows"、"columns"、"category"
type DimensionGroup struct {
	Name   string           `json:"name"`
	Label  string           `json:"label"`
	Fields []DimensionField `json:"fields"`
}

// DimensionField 维度字段绑定
type DimensionField struct {
	Field      string `json:"field"`
	Label      string `json:"label,omitempty"`
	Granularity string `json:"granularity,omitempty"` // 时间粒度: day, week, month, etc.
}

// MetricGroup 指标组，例如 "values"、"primary_values"、"secondary_values"
type MetricGroup struct {
	Name   string        `json:"name"`
	Label  string        `json:"label"`
	Fields []MetricField `json:"fields"`
}

// MetricField 指标字段绑定
type MetricField struct {
	Field  string          `json:"field"`
	Label  string          `json:"label,omitempty"`
	Agg    AggregationType `json:"agg"`
	Alias  string          `json:"alias,omitempty"`
	Unit   string          `json:"unit,omitempty"`
	Format string          `json:"format,omitempty"`
}

// QuerySpec 查询语义规格，表达"查什么"，不包含视觉语义。
// 属于 Query 语义层，由 ChartSpec 转换而来。
type QuerySpec struct {
	Dimensions []DimensionExpr `json:"dimensions"`
	Metrics    []MetricExpr2   `json:"metrics"`
	Filters    []FilterConfig  `json:"filters"`
	Sort       *SortConfig     `json:"sort,omitempty"`
	Pagination *Pagination     `json:"pagination,omitempty"`
	Limit      int             `json:"limit,omitempty"`
}

// DimensionExpr 结构化维度表达式
type DimensionExpr struct {
	Field       string `json:"field"`
	Label       string `json:"label,omitempty"`
	Granularity string `json:"granularity,omitempty"`
}

// MetricExpr2 结构化指标表达式（命名避免与 ast.go 中的 MetricExpr 冲突）
type MetricExpr2 struct {
	Field  string          `json:"field"`
	Agg    AggregationType `json:"agg"`
	Alias  string          `json:"alias,omitempty"`
	Unit   string          `json:"unit,omitempty"`
	Format string          `json:"format,omitempty"`
}

// ChartSpecFromRequest 将旧的 ChartQueryRequest 转换为 ChartSpec（兼容 adapter）。
// 旧请求的 dims/metrics 映射为默认组。
func ChartSpecFromRequest(req *ChartQueryRequest) *ChartSpec {
	spec := &ChartSpec{
		ChartType: req.ChartType,
		Style:     make(map[string]any),
		QueryOptions: make(map[string]any),
	}

	// 维度 → 默认维度组
	if len(req.Dims) > 0 {
		fields := make([]DimensionField, len(req.Dims))
		for i, d := range req.Dims {
			fields[i] = DimensionField{Field: d}
		}
		spec.DimensionGroups = []DimensionGroup{
			{Name: defaultDimGroupName(req.ChartType), Label: "维度", Fields: fields},
		}
	}

	// 指标 → 默认指标组
	if len(req.Metrics) > 0 {
		fields := make([]MetricField, len(req.Metrics))
		for i, m := range req.Metrics {
			fields[i] = MetricField{
				Field: m.Field,
				Agg:   m.Agg,
				Alias: m.Alias,
			}
		}
		spec.MetricGroups = []MetricGroup{
			{Name: defaultMetricGroupName(req.ChartType), Label: "指标", Fields: fields},
		}
	}

	return spec
}

// QuerySpecFromRequest 将旧的 ChartQueryRequest 转换为 QuerySpec（兼容 adapter）。
func QuerySpecFromRequest(req *ChartQueryRequest) *QuerySpec {
	spec := &QuerySpec{
		Filters:    req.Filters,
		Sort:       req.Sort,
		Pagination: req.Pagination,
	}

	for _, d := range req.Dims {
		spec.Dimensions = append(spec.Dimensions, DimensionExpr{Field: d})
	}

	for _, m := range req.Metrics {
		spec.Metrics = append(spec.Metrics, MetricExpr2{
			Field: m.Field,
			Agg:   m.Agg,
			Alias: m.Alias,
		})
	}

	return spec
}

// QuerySpecToBuildArgs 将 QuerySpec 拆解为 BunQueryBuilder.Build() 所需的平铺参数。
// 用于兼容模式：新类型 → 旧 builder 路径。
func QuerySpecToBuildArgs(spec *QuerySpec) (dims []string, metrics []MetricConfig, filters []FilterConfig) {
	dims = make([]string, 0, len(spec.Dimensions))
	for _, d := range spec.Dimensions {
		dims = append(dims, d.Field)
	}

	metrics = make([]MetricConfig, 0, len(spec.Metrics))
	for _, m := range spec.Metrics {
		metrics = append(metrics, MetricConfig{
			Field: m.Field,
			Agg:   m.Agg,
			Alias: m.Alias,
		})
	}

	filters = spec.Filters
	if filters == nil {
		filters = []FilterConfig{}
	}

	return
}

// defaultDimGroupName 根据图表类型返回默认维度组名
func defaultDimGroupName(chartType ChartType) string {
	switch chartType {
	case ChartTypePivot:
		return "rows"
	case ChartTypePie:
		return "category"
	case ChartTypeScatter:
		return "dims"
	default:
		return "x_axis"
	}
}

// defaultMetricGroupName 根据图表类型返回默认指标组名
func defaultMetricGroupName(chartType ChartType) string {
	switch chartType {
	case ChartTypeScatter:
		return "values"
	default:
		return "values"
	}
}
