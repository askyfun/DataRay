package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"dataray/internal/datasource"
	"dataray/internal/model"
	"dataray/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type DatasourceHandler struct {
	db *bun.DB
}

func NewDatasourceHandler(db *bun.DB) *DatasourceHandler {
	return &DatasourceHandler{db: db}
}

func (h *DatasourceHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	var datasources []model.Datasource
	query := h.db.NewSelect().Model(&datasources)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, datasources)
}

type CreateDatasourceRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var req CreateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "postgresql"
	}

	ds := &model.Datasource{
		Name:         req.Name,
		Type:         req.Type,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	if _, err := h.db.NewInsert().Model(ds).Returning("*").Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

func (h *DatasourceHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ds := &model.Datasource{ID: id}
	if err := h.db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

type UpdateDatasourceRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req UpdateDatasourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ds := &model.Datasource{
		ID:           id,
		Name:         req.Name,
		Type:         req.Type,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	if _, err := h.db.NewUpdate().Model(ds).WherePK().Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated := &model.Datasource{ID: id}
	if err := h.db.NewSelect().Model(updated).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if _, err := h.db.NewDelete().Model(&model.Datasource{}).Where("id = ?", id).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type TestConnectionRequest struct {
	Type         string `json:"type"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = "postgresql"
	}

	driverType := datasource.DriverType(req.Type)
	driver, err := datasource.NewDriver(driverType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported driver type: %s", req.Type)})
		return
	}

	config := datasource.ConnectionConfig{
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	if err := driver.TestConnection(c.Request.Context(), config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("connection failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *DatasourceHandler) GetTables(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	ds := &model.Datasource{ID: id}
	if err := h.db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	dsService := service.NewDatasourceService(h.db)
	conn, err := dsService.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	tables, err := conn.GetTables(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get tables: %v", err)})
		return
	}

	c.JSON(http.StatusOK, tables)
}

func (h *DatasourceHandler) GetTableColumns(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tableName := c.Param("table")

	ds := &model.Datasource{ID: id}
	if err := h.db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	dsService := service.NewDatasourceService(h.db)
	conn, err := dsService.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	columns, err := conn.GetColumns(c.Request.Context(), tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get columns: %v", err)})
		return
	}

	result := make([]map[string]interface{}, len(columns))
	for i, col := range columns {
		result[i] = map[string]interface{}{
			"name":     col.Name,
			"dataType": col.Type,
			"comment":  col.Comment,
		}
	}
	c.JSON(http.StatusOK, result)
}

type PreviewRequest struct {
	TableName string `json:"table_name"`
	QuerySQL  string `json:"query_sql"`
	QueryType string `json:"query_type"`
}

func (h *DatasourceHandler) Preview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dsService := service.NewDatasourceService(h.db)
	conn, err := dsService.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	var query string
	if req.QueryType == "sql" {
		query = fmt.Sprintf("%s LIMIT 10", req.QuerySQL)
	} else {
		if !isValidIdentifier(req.TableName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table name"})
			return
		}
		query = fmt.Sprintf("SELECT * FROM %s LIMIT 10", req.TableName)
	}

	result, err := conn.Execute(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"columns": result.Columns,
		"data":    result.Rows,
	})
}

type FieldDistributionRequest struct {
	TableName string `json:"table_name"`
	QuerySQL  string `json:"query_sql"`
	QueryType string `json:"query_type"`
	FieldName string `json:"field_name"`
	Limit     int    `json:"limit"`
}

func (h *DatasourceHandler) GetFieldDistribution(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var req FieldDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.FieldName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field_name is required"})
		return
	}

	if req.Limit <= 0 || req.Limit > 50 {
		req.Limit = 20
	}

	dsService := service.NewDatasourceService(h.db)
	conn, err := dsService.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	var query string
	var baseTable string

	if req.QueryType == "sql" {
		query = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM (%s) as _subquery GROUP BY %s ORDER BY _count DESC LIMIT %d",
			req.FieldName, req.QuerySQL, req.FieldName, req.Limit)
	} else {
		if !isValidIdentifier(req.TableName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid table name"})
			return
		}
		baseTable = req.TableName
		query = fmt.Sprintf("SELECT %s, COUNT(*) as _count FROM %s GROUP BY %s ORDER BY _count DESC LIMIT %d",
			req.FieldName, baseTable, req.FieldName, req.Limit)
	}

	result, err := conn.Execute(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to query: %v", err)})
		return
	}

	var totalCount int64
	var totalQuery string
	if req.QueryType == "sql" {
		totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM (%s) as _subquery", req.QuerySQL)
	} else {
		totalQuery = fmt.Sprintf("SELECT COUNT(*) as _total FROM %s", baseTable)
	}

	totalResult, err := conn.Execute(c.Request.Context(), totalQuery)
	if err == nil && len(totalResult.Rows) > 0 {
		if val, ok := totalResult.Rows[0]["_total"]; ok {
			switch v := val.(type) {
			case int64:
				totalCount = v
			case float64:
				totalCount = int64(v)
			}
		}
	}

	distribution := make([]map[string]interface{}, 0, len(result.Rows))
	for _, row := range result.Rows {
		value := row[req.FieldName]
		count := row["_count"]

		var countInt int64
		switch v := count.(type) {
		case int64:
			countInt = v
		case float64:
			countInt = int64(v)
		}

		var percentage float64
		if totalCount > 0 {
			percentage = float64(countInt) * 100.0 / float64(totalCount)
		}

		distribution = append(distribution, map[string]interface{}{
			"value":      value,
			"count":      countInt,
			"percentage": percentage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"field_name":   req.FieldName,
		"total_count":  totalCount,
		"unique_count": len(result.Rows),
		"distribution": distribution,
	})
}

func parseID(c *gin.Context, paramName string) (int, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s parameter: must be a positive integer", paramName)
	}
	return id, nil
}

func getPaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return
}

func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	runes := []rune(name)
	first := runes[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return false
	}
	for _, c := range runes[1:] {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}
