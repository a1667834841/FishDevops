package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"xianyu_aner/internal/model"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HandleHealth 处理健康检查请求
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status: "ok",
		Time:   time.Now().Format(time.RFC3339),
	})
}
