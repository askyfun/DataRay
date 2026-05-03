package handler

import (
	"strconv"

	"dataray/internal/domain/entity"
	"dataray/internal/response"
	"dataray/internal/service/datasource"

	"github.com/gin-gonic/gin"
)

// DatasourceHandler handles datasource HTTP requests
type DatasourceHandler struct {
	svc datasource.Service
}

// NewDatasourceHandler creates a new DatasourceHandler
func NewDatasourceHandler(svc datasource.Service) *DatasourceHandler {
	return &DatasourceHandler{svc: svc}
}

// List handles GET /api/datasources
func (h *DatasourceHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	datasources, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if datasources == nil {
		datasources = []entity.Datasource{}
	}
	response.Success(c, datasources)
}

// Get handles GET /api/datasources/:id
func (h *DatasourceHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	ds, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, ds)
}

// Create handles POST /api/datasources
func (h *DatasourceHandler) Create(c *gin.Context) {
	var req struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		DatabaseName string `json:"database_name"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "postgresql"
	}

	ds := &entity.Datasource{
		Name:         req.Name,
		Type:         req.Type,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	result, err := h.svc.Create(c.Request.Context(), ds)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Update handles PUT /api/datasources/:id
func (h *DatasourceHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		DatabaseName string `json:"database_name"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ds := &entity.Datasource{
		ID:           id,
		Name:         req.Name,
		Type:         req.Type,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	result, err := h.svc.Update(c.Request.Context(), ds)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete handles DELETE /api/datasources/:id
func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// TestConnection handles POST /api/datasources/test
func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	var req struct {
		Type         string `json:"type"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		DatabaseName string `json:"database_name"`
		Username     string `json:"username"`
		Password     string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "postgresql"
	}

	config := entity.DatasourceConnectionConfig{
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     req.Password,
	}

	if err := h.svc.TestConnection(c.Request.Context(), config, req.Type); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"status": "ok"})
}

// GetTables handles GET /api/datasources/:id/tables
func (h *DatasourceHandler) GetTables(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	tables, err := h.svc.GetTables(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, tables)
}

// GetColumns handles GET /api/datasources/:id/tables/:table/columns
func (h *DatasourceHandler) GetColumns(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	tableName := c.Param("table")
	if tableName == "" {
		response.BadRequest(c, "table name is required")
		return
	}

	columns, err := h.svc.GetColumns(c.Request.Context(), id, tableName)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := make([]map[string]interface{}, len(columns))
	for i, col := range columns {
		result[i] = map[string]interface{}{
			"name":      col.Name,
			"data_type": col.DataType,
			"comment":   col.Comment,
		}
	}
	response.Success(c, result)
}

// Preview handles POST /api/datasources/:id/preview
func (h *DatasourceHandler) Preview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		TableName string `json:"table_name"`
		QuerySQL  string `json:"query_sql"`
		QueryType string `json:"query_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Preview(c.Request.Context(), id, req.TableName, req.QuerySQL, req.QueryType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetFieldDistribution handles POST /api/datasources/:id/field-distribution
func (h *DatasourceHandler) GetFieldDistribution(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		TableName string `json:"table_name"`
		QuerySQL  string `json:"query_sql"`
		QueryType string `json:"query_type"`
		FieldName string `json:"field_name"`
		Limit     int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.FieldName == "" {
		response.BadRequest(c, "field_name is required")
		return
	}

	result, err := h.svc.GetFieldDistribution(c.Request.Context(), id, req.TableName, req.QuerySQL, req.QueryType, req.FieldName, req.Limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}
