package query

// PlannedQuery 代表 QueryPlanner 为旧执行链路产出的兼容结果。
// 调用场景：兼容期内由 QuerySpec 规划出旧 builder / executor 仍可消费的平铺参数。
type PlannedQuery struct {
	Dims       []string
	Metrics    []MetricConfig
	Filters    []FilterConfig
	Sort       *SortConfig
	Pagination *Pagination
	Limit      int
}

// QueryPlanner 负责将 QuerySpec 规划为当前执行链路可消费的查询参数。
// 当前仅支持基础维度、基础聚合、线性过滤、排序和分页能力。
type QueryPlanner struct{}

// NewQueryPlanner 创建最小 QueryPlanner。
func NewQueryPlanner() *QueryPlanner {
	return &QueryPlanner{}
}

// Plan 根据 QuerySpec 生成兼容旧执行链路的 PlannedQuery。
// 当前实现保持与 QuerySpecToBuildArgs 等价，为后续扩展更复杂规划能力预留入口。
func (p *QueryPlanner) Plan(spec *QuerySpec) *PlannedQuery {
	if spec == nil {
		return &PlannedQuery{
			Dims:    []string{},
			Metrics: []MetricConfig{},
			Filters: []FilterConfig{},
		}
	}

	dims, metrics, filters := QuerySpecToBuildArgs(spec)

	return &PlannedQuery{
		Dims:       dims,
		Metrics:    metrics,
		Filters:    filters,
		Sort:       spec.Sort,
		Pagination: spec.Pagination,
		Limit:      spec.Limit,
	}
}
