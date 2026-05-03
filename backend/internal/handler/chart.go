package handler

import (
	"strconv"

	"dataray/internal/domain/entity"
	"dataray/internal/response"
	"dataray/internal/service/chart"

	"github.com/gin-gonic/gin"
)

// ChartHandler handles chart HTTP requests
type ChartHandler struct {
	svc chart.Service
}

// NewChartHandler creates a new ChartHandler
func NewChartHandler(svc chart.Service) *ChartHandler {
	return &ChartHandler{svc: svc}
}

// List handles GET /api/charts
func (h *ChartHandler) List(c *gin.Context) {
	limit, offset := getPaginationParams(c)

	charts, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if charts == nil {
		charts = []entity.Chart{}
	}
	response.Success(c, charts)
}

// Get handles GET /api/charts/:id
func (h *ChartHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	chart, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, chart)
}

// Create handles POST /api/charts
func (h *ChartHandler) Create(c *gin.Context) {
	var chart entity.Chart
	if err := c.ShouldBindJSON(&chart); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if chart.Config == "" {
		chart.Config = "{}"
	}

	result, err := h.svc.Create(c.Request.Context(), &chart)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Update handles PUT /api/charts/:id
func (h *ChartHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		Name      string `json:"name"`
		DatasetID int    `json:"dataset_id"`
		ChartType string `json:"chart_type"`
		Config    string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	chart := &entity.Chart{
		ID:        id,
		Name:      req.Name,
		DatasetID: req.DatasetID,
		ChartType: req.ChartType,
		Config:    req.Config,
	}

	result, err := h.svc.Update(c.Request.Context(), chart)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// Delete handles DELETE /api/charts/:id
func (h *ChartHandler) Delete(c *gin.Context) {
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

// GetData handles GET /api/charts/:id/data
func (h *ChartHandler) GetData(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	result, err := h.svc.GetData(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result.Data)
}

// Query handles POST /api/charts/query
func (h *ChartHandler) Query(c *gin.Context) {
	var req entity.ChartQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.svc.Query(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result.Data)
}
