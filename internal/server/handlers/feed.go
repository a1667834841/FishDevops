package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"xianyu_aner/internal/model"
	"xianyu_aner/pkg/mtop"
)

// FeedHandler Feed处理器
type FeedHandler struct {
	mtopClient *mtop.Client
}

// NewFeedHandler 创建Feed处理器
func NewFeedHandler(mtopClient *mtop.Client) *FeedHandler {
	return &FeedHandler{mtopClient: mtopClient}
}

// HandleFeed 处理猜你喜欢请求
func (h *FeedHandler) HandleFeed(c *gin.Context) {
	var req model.FeedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "参数错误: pages 必须是 1-10 之间的整数",
		})
		return
	}

	req = h.applyDefaults(req)
	h.logRequest(req)

	items, err := h.fetchFeedItems(req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.logSuccess(items)
	c.JSON(http.StatusOK, model.FeedResponse{
		Success: true,
		Data: model.FeedData{
			Total:  len(items),
			Pages:  req.Pages,
			MachID: req.MachID,
			Items:  items,
		},
	})
}

func (h *FeedHandler) applyDefaults(req model.FeedRequest) model.FeedRequest {
	if req.Pages == 0 {
		req.Pages = 1
	}
	if req.DaysWithin == 0 {
		req.DaysWithin = 7
	}
	return req
}

func (h *FeedHandler) fetchFeedItems(req model.FeedRequest) ([]mtop.FeedItem, error) {
	return h.mtopClient.GuessYouLike(req.MachID, req.Pages, mtop.GuessYouLikeOptions{
		MinWantCount: req.MinWantCount,
		DaysWithin:   req.DaysWithin,
	})
}

func (h *FeedHandler) logRequest(req model.FeedRequest) {
	log.Printf("收到请求: pages=%d, machId=%s, minWantCount=%d, daysWithin=%d",
		req.Pages, req.MachID, req.MinWantCount, req.DaysWithin)
}

func (h *FeedHandler) logSuccess(items []mtop.FeedItem) {
	log.Printf("成功获取 %d 条商品（已过滤）", len(items))
}

func (h *FeedHandler) handleError(c *gin.Context, err error) {
	log.Printf("获取数据失败: %v", err)
	c.JSON(http.StatusInternalServerError, model.ErrorResponse{
		Success: false,
		Error:   fmt.Sprintf("获取数据失败: %v", err),
	})
}
