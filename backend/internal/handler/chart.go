package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"dataray/internal/model"
	"dataray/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ChartHandler struct {
	db *bun.DB
}

func NewChartHandler(db *bun.DB) *ChartHandler {
	return &ChartHandler{db: db}
}

func (h *ChartHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	var charts []model.Chart
	query := h.db.NewSelect().Model(&charts)
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
	c.JSON(http.StatusOK, charts)
}

func (h *ChartHandler) Create(c *gin.Context) {
	var chart model.Chart
	if err := c.ShouldBindJSON(&chart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if chart.Config == "" {
		chart.Config = "{}"
	}

	chartService := service.NewChartService(h.db)
	created, err := chartService.Create(c.Request.Context(), &chart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *ChartHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	chart := &model.Chart{ID: id}
	if err := h.db.NewSelect().Model(chart).WherePK().Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chart)
}

type UpdateChartRequest struct {
	Name      string `json:"name"`
	ChartType string `json:"chart_type"`
	Config    string `json:"config"`
}

func (h *ChartHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req UpdateChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chartService := service.NewChartService(h.db)
	updated, err := chartService.Update(c.Request.Context(), id, req.Name, req.ChartType, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *ChartHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	chartService := service.NewChartService(h.db)
	if err := chartService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ChartHandler) GetData(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	chartService := service.NewChartService(h.db)
	result, err := chartService.GetData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.Rows)
}

func (h *ChartHandler) Query(c *gin.Context) {
	var req struct {
		DatasetID int `json:"dataset_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = req
	c.JSON(http.StatusOK, gin.H{"error": "not implemented, use POST /api/charts/query with query.ChartQueryRequest"})
}

func (h *ChartHandler) ExecuteQuery(c *gin.Context) {
	_ = c
	_ = fmt.Sprintf("not implemented")
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
