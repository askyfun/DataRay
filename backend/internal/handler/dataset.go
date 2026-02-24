package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"dataray/internal/model"
	"dataray/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type DatasetHandler struct {
	db *bun.DB
}

func NewDatasetHandler(db *bun.DB) *DatasetHandler {
	return &DatasetHandler{db: db}
}

func (h *DatasetHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	var datasets []model.Dataset
	query := h.db.NewSelect().Model(&datasets)
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
	c.JSON(http.StatusOK, datasets)
}

type CreateDatasetRequest struct {
	Name         string `json:"name"`
	DatasourceID int    `json:"datasource_id"`
	TableName    string `json:"table_name"`
	QuerySQL     string `json:"query_sql"`
	QueryType    string `json:"query_type"`
	Mode         string `json:"mode"`
	Description  string `json:"description"`
	Tags         string `json:"tags"`
	Columns      string `json:"columns"`
}

func (h *DatasetHandler) Create(c *gin.Context) {
	var req CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request body: %v", err)})
		return
	}

	ds := &model.Dataset{
		Name:         req.Name,
		DatasourceID: req.DatasourceID,
		QueryType:    req.QueryType,
		Mode:         req.Mode,
	}

	if req.TableName != "" {
		ds.TableName = sql.NullString{String: req.TableName, Valid: true}
	}
	if req.QuerySQL != "" {
		ds.QuerySQL = sql.NullString{String: req.QuerySQL, Valid: true}
	}
	if req.Description != "" {
		ds.Description = sql.NullString{String: req.Description, Valid: true}
	}

	dsService := service.NewDatasetService(h.db)
	created, err := dsService.Create(c.Request.Context(), ds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *DatasetHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ds := &model.Dataset{ID: id}
	if err := h.db.NewSelect().Model(ds).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ds)
}

func (h *DatasetHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	dsService := service.NewDatasetService(h.db)
	if err := dsService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *DatasetHandler) GetColumns(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	dsService := service.NewDatasetService(h.db)
	columns, err := dsService.GetColumns(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, columns)
}

func (h *DatasetHandler) UpdateColumns(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var columns []model.DatasetColumn
	if err := c.ShouldBindJSON(&columns); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	columnsJSON, err := json.Marshal(columns)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dsService := service.NewDatasetService(h.db)
	updated, err := dsService.UpdateColumns(c.Request.Context(), id, string(columnsJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	dsService := service.NewDatasetService(h.db)
	result, err := dsService.Preview(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"columns": result.Columns,
		"data":    result.Rows,
	})
}
