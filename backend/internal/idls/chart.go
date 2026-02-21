package idls

type ChartType string

const (
	ChartTypeLine    ChartType = "line"
	ChartTypeBar     ChartType = "bar"
	ChartTypePie     ChartType = "pie"
	ChartTypeScatter ChartType = "scatter"
	ChartTypeTable   ChartType = "table"
)

type CreateChartRequest struct {
	Name      string    `json:"name" binding:"required"`
	DatasetID int       `json:"dataset_id" binding:"required"`
	ChartType ChartType `json:"chart_type" binding:"required"`
	Config    string    `json:"config" binding:"required"`
}

type UpdateChartRequest struct {
	Name      *string    `json:"name,omitempty"`
	DatasetID *int       `json:"dataset_id,omitempty"`
	ChartType *ChartType `json:"chart_type,omitempty"`
	Config    *string    `json:"config,omitempty"`
}

type ChartResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	DatasetID int    `json:"dataset_id"`
	ChartType string `json:"chart_type"`
	Config    string `json:"config"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type ChartListResponse struct {
	Items []ChartResponse `json:"items"`
	Total int64           `json:"total"`
}

type ChartQueryAggregation string

const (
	ChartQueryAggSum   ChartQueryAggregation = "sum"
	ChartQueryAggAvg   ChartQueryAggregation = "avg"
	ChartQueryAggCount ChartQueryAggregation = "count"
	ChartQueryAggMax   ChartQueryAggregation = "max"
	ChartQueryAggMin   ChartQueryAggregation = "min"
)

type ChartQueryMetric struct {
	Field string                `json:"field" binding:"required"`
	Agg   ChartQueryAggregation `json:"agg" binding:"required"`
	Alias *string               `json:"alias,omitempty"`
}

type ChartQueryFilterOp string

const (
	FilterOpEq        ChartQueryFilterOp = "eq"
	FilterOpNeq       ChartQueryFilterOp = "neq"
	FilterOpGt        ChartQueryFilterOp = "gt"
	FilterOpGte       ChartQueryFilterOp = "gte"
	FilterOpLt        ChartQueryFilterOp = "lt"
	FilterOpLte       ChartQueryFilterOp = "lte"
	FilterOpLike      ChartQueryFilterOp = "like"
	FilterOpIn        ChartQueryFilterOp = "in"
	FilterOpBetween   ChartQueryFilterOp = "between"
	FilterOpIsNull    ChartQueryFilterOp = "isNull"
	FilterOpIsNotNull ChartQueryFilterOp = "isNotNull"
)

type ChartQueryFilter struct {
	Field    string             `json:"field" binding:"required"`
	Op       ChartQueryFilterOp `json:"op" binding:"required"`
	Value    any                `json:"value"`
	ValueEnd *any               `json:"value_end,omitempty"`
	Logic    string             `json:"logic"`
}

type ChartQueryPagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ChartQuerySort struct {
	Field string `json:"field" binding:"required"`
	Order string `json:"order" binding:"required"`
}

type ChartQueryRequest struct {
	DatasetID  int                   `json:"dataset_id" binding:"required"`
	ChartType  string                `json:"chart_type" binding:"required"`
	Dims       []string              `json:"dims"`
	Metrics    []ChartQueryMetric    `json:"metrics"`
	Filters    []ChartQueryFilter    `json:"filters"`
	Pagination *ChartQueryPagination `json:"pagination,omitempty"`
	Sort       *ChartQuerySort       `json:"sort,omitempty"`
}

type TableResponse struct {
	Columns    []string         `json:"columns"`
	Data       []map[string]any `json:"data"`
	Pagination *TablePagination `json:"pagination"`
}

type TablePagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type PieResponseItem struct {
	Name       string  `json:"name"`
	Value      float64 `json:"value"`
	Percentage float64 `json:"percentage"`
}

type PieResponse struct {
	Data []PieResponseItem `json:"data"`
}

type AxisSeries struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

type AxisResponse struct {
	XAxis  []string     `json:"x_axis"`
	Series []AxisSeries `json:"series"`
}

type ScatterResponse struct {
	Data [][2]float64 `json:"data"`
}

type ChartDataResponse = any
