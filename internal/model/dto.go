package model

import (
	"xianyu_aner/pkg/feishu"
)

// FeedRequest 猜你喜欢请求参数
type FeedRequest struct {
	Pages       int `form:"pages" binding:"omitempty,min=1,max=10"`
	MachID      string `form:"machId"`
	MinWantCount int `form:"minWantCount" binding:"omitempty,min=0"` // 最低想要人数
	DaysWithin   int `form:"daysWithin" binding:"omitempty,min=0"`    // 发布时间范围（天）
}

// FeedResponse 猜你喜欢响应
type FeedResponse struct {
	Success bool `json:"success"`
	Data    FeedData `json:"data"`
	Message string `json:"message,omitempty"`
}

// FeedData 商品数据
type FeedData struct {
	Total  int         `json:"total"`
	Pages  int         `json:"pages"`
	MachID string      `json:"machId"`
	Items  interface{} `json:"items"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// FeishuPushRequest 飞书推送请求
type FeishuPushRequest struct {
	Date        string          `json:"date" binding:"required"`     // 日期
	Products    []feishu.Product `json:"products" binding:"required"` // 商品列表
	AppToken    string          `json:"appToken,omitempty"`          // 可选：覆盖默认配置
	TableToken  string          `json:"tableToken,omitempty"`        // 可选：覆盖默认配置
}

// FeishuPushResponse 飞书推送响应
type FeishuPushResponse struct {
	Success bool             `json:"success"`
	Message string            `json:"message,omitempty"`
	Data    FeishuPushData    `json:"data,omitempty"`
}

// FeishuPushData 飞书推送数据
type FeishuPushData struct {
	RecordsCreated int    `json:"recordsCreated"`
	RecordsUpdated int    `json:"recordsUpdated"`
	TableToken     string `json:"tableToken"`
}
