package idls

type DatasetMode string

const (
	DatasetModeDirect      DatasetMode = "direct"
	DatasetModeAccelerated DatasetMode = "accelerated"
)

type QueryType string

const (
	QueryTypeTable QueryType = "table"
	QueryTypeSQL   QueryType = "sql"
)

type ColumnRole string

const (
	ColumnRoleDimension ColumnRole = "dimension"
	ColumnRoleMetric    ColumnRole = "metric"
)

type CreateDatasetRequest struct {
	Name         string      `json:"name" binding:"required"`
	DatasourceID int         `json:"datasource_id" binding:"required"`
	TableName    *string     `json:"table_name"`
	QuerySQL     *string     `json:"query_sql"`
	QueryType    QueryType   `json:"query_type" binding:"required"`
	Mode         DatasetMode `json:"mode"`
	Description  *string     `json:"description"`
	Tags         []string    `json:"tags"`
}

type UpdateDatasetRequest = CreateDatasetRequest

type DatasetColumn struct {
	Name       string `json:"name"`
	Expr       string `json:"expr"`
	Type       string `json:"type"`
	TypeConfig any    `json:"type_config,omitempty"`
	Comment    string `json:"comment"`
	Role       string `json:"role"`
}

type DatasetResponse struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	DatasourceID     int      `json:"datasource_id"`
	TableName        *string  `json:"table_name,omitempty"`
	QuerySQL         *string  `json:"query_sql,omitempty"`
	QueryType        string   `json:"query_type"`
	Mode             string   `json:"mode"`
	AccelerateConfig any      `json:"accelerate_config,omitempty"`
	Description      *string  `json:"description,omitempty"`
	Tags             []string `json:"tags"`
	RefreshStrategy  any      `json:"refresh_strategy,omitempty"`
	PreviewData      any      `json:"preview_data,omitempty"`
	QualityRules     any      `json:"quality_rules,omitempty"`
	Columns          string   `json:"columns"`
	CreatedAt        string   `json:"created_at,omitempty"`
	UpdatedAt        string   `json:"updated_at,omitempty"`
}

type DatasetListResponse struct {
	Items []DatasetResponse `json:"items"`
	Total int64             `json:"total"`
}

type UpdateColumnsRequest struct {
	Columns []DatasetColumn `json:"columns" binding:"required"`
}

type DatasetPreviewResponse struct {
	Columns []string         `json:"columns"`
	Data    []map[string]any `json:"data"`
}
