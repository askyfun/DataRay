package handler

import (
	"net/http"

	"dataray/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

type ShareHandler struct {
	db *bun.DB
}

func NewShareHandler(db *bun.DB) *ShareHandler {
	return &ShareHandler{db: db}
}

type CreateShareRequest struct {
	ChartID int `json:"chart_id"`
}

func (h *ShareHandler) Create(c *gin.Context) {
	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shareService := service.NewShareService(h.db)
	share, err := shareService.Create(c.Request.Context(), req.ChartID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, share)
}

func (h *ShareHandler) Get(c *gin.Context) {
	token := c.Param("token")

	shareService := service.NewShareService(h.db)
	share, err := shareService.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}

	c.JSON(http.StatusOK, share)
}

func (h *ShareHandler) View(c *gin.Context) {
	token := c.Param("token")

	shareService := service.NewShareService(h.db)
	share, err := shareService.GetByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}

	if shareService.IsExpired(share) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	c.Redirect(http.StatusFound, "/#/share/"+token)
}
