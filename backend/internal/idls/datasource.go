package idls

type DatasourceType string

const (
	DatasourceTypePostgresql DatasourceType = "postgresql"
	DatasourceTypeMysql      DatasourceType = "mysql"
	DatasourceTypeClickhouse DatasourceType = "clickhouse"
	DatasourceTypeStarrocks  DatasourceType = "starrocks"
)

type CreateDatasourceRequest struct {
	Name         string         `json:"name" binding:"required"`
	Type         DatasourceType `json:"type" binding:"required"`
	Host         string         `json:"host" binding:"required"`
	Port         int            `json:"port" binding:"required"`
	DatabaseName string         `json:"database_name" binding:"required"`
	Username     string         `json:"username" binding:"required"`
	Password     string         `json:"password" binding:"required"`
}

type UpdateDatasourceRequest = CreateDatasourceRequest

type DatasourceResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type DatasourceListResponse struct {
	Items []DatasourceResponse `json:"items"`
	Total int64                `json:"total"`
}

type TestConnectionRequest struct {
	Type         DatasourceType `json:"type" binding:"required"`
	Host         string         `json:"host" binding:"required"`
	Port         int            `json:"port" binding:"required"`
	DatabaseName string         `json:"database_name" binding:"required"`
	Username     string         `json:"username" binding:"required"`
	Password     string         `json:"password" binding:"required"`
}

type TestConnectionResponse struct {
	Status string `json:"status"`
}

type TableInfo struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	Comment    string `json:"comment"`
	Role       string `json:"role"`
	IsVirtual  bool   `json:"is_virtual"`
	Expression string `json:"expression"`
}

type DatasetPreview struct {
	Columns []string         `json:"columns"`
	Data    []map[string]any `json:"data"`
}

type FieldDistributionValue struct {
	Value      any     `json:"value"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type FieldDistribution struct {
	FieldName    string                   `json:"field_name"`
	TotalCount   int                      `json:"total_count"`
	UniqueCount  int                      `json:"unique_count"`
	Distribution []FieldDistributionValue `json:"distribution"`
}
