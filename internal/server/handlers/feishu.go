package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"xianyu_aner/internal/model"
	"xianyu_aner/pkg/feishu"
)

// FeishuHandler 飞书处理器
type FeishuHandler struct {
	feishuClient *feishu.Client
	feishuConfig *feishu.BitableConfig
}

// NewFeishuHandler 创建飞书处理器
func NewFeishuHandler(feishuClient *feishu.Client, feishuConfig *feishu.BitableConfig) *FeishuHandler {
	return &FeishuHandler{
		feishuClient: feishuClient,
		feishuConfig: feishuConfig,
	}
}

// HandleFeishuPush 处理飞书推送请求
func (h *FeishuHandler) HandleFeishuPush(c *gin.Context) {
	// 检查是否配置了飞书客户端
	if h.feishuClient == nil || h.feishuConfig == nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Success: false,
			Error:   "飞书服务未配置，请设置 FeishuAppID 和 FeishuAppSecret",
		})
		return
	}

	var req model.FeishuPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("参数错误: %v", err),
		})
		return
	}

	log.Printf("收到飞书推送请求: date=%s, products=%d", req.Date, len(req.Products))

	// 使用请求中的token或默认配置
	appToken := h.feishuConfig.AppToken
	tableToken := h.feishuConfig.TableToken

	if req.AppToken != "" {
		appToken = req.AppToken
	}
	if req.TableToken != "" {
		tableToken = req.TableToken
	}

	// 调用飞书客户端推送数据
	result, err := h.feishuClient.PushToBitable(appToken, tableToken, req.Products)
	if err != nil {
		log.Printf("推送失败: %v", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("推送失败: %v", err),
		})
		return
	}

	log.Printf("推送成功: created=%d", result.Data.RecordsCreated)

	// 构建响应
	c.JSON(http.StatusOK, model.FeishuPushResponse{
		Success: true,
		Message: fmt.Sprintf("成功推送 %d 条记录到飞书表格", result.Data.RecordsCreated),
		Data: model.FeishuPushData{
			RecordsCreated: result.Data.RecordsCreated,
			RecordsUpdated: result.Data.RecordsUpdated,
			TableToken:     result.Data.TableToken,
		},
	})
}
