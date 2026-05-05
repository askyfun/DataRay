package handler

import (
	"strconv"

	"dataray/internal/domain/entity"
	"dataray/internal/response"
	"dataray/internal/service/dataset"

	"github.com/gin-gonic/gin"
)

// DatasetHandler handles dataset HTTP requests
type DatasetHandler struct {
	svc dataset.Service
}

// NewDatasetHandler creates a new DatasetHandler
func NewDatasetHandler(svc dataset.Service) *DatasetHandler {
	return &DatasetHandler{svc: svc}
}

// List handles GET /api/datasets
func (h *DatasetHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	datasets, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if datasets == nil {
		datasets = []entity.Dataset{}
	}
	response.Success(c, datasets)
}

// Get handles GET /api/datasets/:id
func (h *DatasetHandler) Get(c *gin.Context) {
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

// Create handles POST /api/datasets
func (h *DatasetHandler) Create(c *gin.Context) {
	var req struct {
		Name         string `json:"name"`
		DatasourceID int    `json:"datasource_id"`
		TableName    string `json:"table_name"`
		QuerySQL     string `json:"query_sql"`
		QueryType    string `json:"query_type"`
		Mode         string `json:"mode"`
		Description  string `json:"description"`
		Tags         string `json:"tags"`
		Columns      string `json:"columns"`
		ShardEnabled bool   `json:"shard_enabled"`
		ShardKeys    string `json:"shard_keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ds := &entity.Dataset{
		Name:         req.Name,
		DatasourceID: req.DatasourceID,
		QueryType:    req.QueryType,
		Mode:         req.Mode,
		Tags:         req.Tags,
		Columns:      req.Columns,
		ShardEnabled: req.ShardEnabled,
		ShardKeys:    req.ShardKeys,
	}
	if req.TableName != "" {
		ds.TableName = &req.TableName
	}
	if req.QuerySQL != "" {
		ds.QuerySQL = &req.QuerySQL
	}
	if req.Description != "" {
		ds.Description = &req.Description
	}
	if ds.QueryType == "" {
		ds.QueryType = "table"
	}
	if ds.Mode == "" {
		ds.Mode = "direct"
	}
	if ds.Tags == "" {
		ds.Tags = "[]"
	}
	if ds.QualityRules == "" {
		ds.QualityRules = "[]"
	}
	if ds.Columns == "" {
		ds.Columns = "[]"
	}

	result, err := h.svc.Create(c.Request.Context(), ds)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete handles DELETE /api/datasets/:id
func (h *DatasetHandler) Delete(c *gin.Context) {
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

// GetColumns handles GET /api/datasets/:id/columns
func (h *DatasetHandler) GetColumns(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	columns, err := h.svc.GetColumns(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, columns)
}

// UpdateColumns handles POST /api/datasets/:id/columns
func (h *DatasetHandler) UpdateColumns(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var columns []entity.DatasetColumn
	if err := c.ShouldBindJSON(&columns); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.UpdateColumns(c.Request.Context(), id, columns)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Preview handles GET /api/datasets/:id/preview
func (h *DatasetHandler) Preview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	result, err := h.svc.Preview(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Query handles POST /api/datasets/:id/query
func (h *DatasetHandler) Query(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var config entity.QueryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Query(c.Request.Context(), id, config)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}
