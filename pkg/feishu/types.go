package feishu

import "time"

// FieldType 飞书字段类型
type FieldType int

const (
	FieldTypeText         FieldType = 1  // 文本
	FieldTypeNumber       FieldType = 2  // 数字
	FieldTypeSingleSelect FieldType = 3  // 单选
	FieldTypeMultiSelect  FieldType = 4  // 多选
	FieldTypeDateTime     FieldType = 5  // 日期时间
	FieldTypeCheckbox     FieldType = 7  // 复选框
	FieldTypeURL          FieldType = 15 // URL
	FieldTypePhone        FieldType = 11 // 电话
	FieldTypeEmail        FieldType = 13 // 邮箱
)

// FieldSchema 字段定义
type FieldSchema struct {
	Type      FieldType `json:"type"`
	Label     string    `json:"label"`
	FieldName string    `json:"fieldName,omitempty"`
}

// ProductFields 商品字段定义（优化后保留24个核心字段，移除重复字段）
var ProductFields = []struct {
	Key      string
	Schema   FieldSchema
	CSVOrder int // CSV 导出顺序，0 表示不导出
}{
	// ==================== 基本信息 ====================
	{"itemId", FieldSchema{Type: FieldTypeText, Label: "商品ID"}, 1},
	{"title", FieldSchema{Type: FieldTypeText, Label: "商品标题"}, 2},
	{"subTitle", FieldSchema{Type: FieldTypeText, Label: "副标题"}, 3},
	{"price", FieldSchema{Type: FieldTypeText, Label: "价格"}, 4},
	{"originalPrice", FieldSchema{Type: FieldTypeText, Label: "原价"}, 5},
	{"condition", FieldSchema{Type: FieldTypeText, Label: "成色"}, 6},

	// ==================== 热度指标 ====================
	{"wantCnt", FieldSchema{Type: FieldTypeNumber, Label: "想要人数"}, 7},
	{"viewCount", FieldSchema{Type: FieldTypeNumber, Label: "浏览次数"}, 8},
	{"collectCount", FieldSchema{Type: FieldTypeNumber, Label: "收藏次数"}, 9},
	{"exposureHeat", FieldSchema{Type: FieldTypeNumber, Label: "曝光热度"}, 10},

	// ==================== 卖家信息 ====================
	{"sellerNick", FieldSchema{Type: FieldTypeText, Label: "卖家昵称"}, 11},
	{"sellerCity", FieldSchema{Type: FieldTypeText, Label: "卖家地区"}, 12},
	{"sellerCredit", FieldSchema{Type: FieldTypeText, Label: "卖家信用"}, 13},
	{"sellerItemCount", FieldSchema{Type: FieldTypeNumber, Label: "在售商品数"}, 14},
	{"sellerSoldCount", FieldSchema{Type: FieldTypeNumber, Label: "已售数量"}, 15},
	{"freeShip", FieldSchema{Type: FieldTypeText, Label: "包邮"}, 16},

	// ==================== 时间信息 ====================
	{"publishTimeMs", FieldSchema{Type: FieldTypeDateTime, Label: "发布时间"}, 17},
	{"captureTimeMs", FieldSchema{Type: FieldTypeDateTime, Label: "采集时间"}, 18},

	// ==================== 链接资源 ====================
	{"coverUrl", FieldSchema{Type: FieldTypeURL, Label: "封面图"}, 19},
	{"detailUrl", FieldSchema{Type: FieldTypeURL, Label: "商品详情"}, 20},
	{"videoUrl", FieldSchema{Type: FieldTypeURL, Label: "视频链接"}, 21},

	// ==================== 其他 ====================
	{"tags", FieldSchema{Type: FieldTypeText, Label: "商品标签"}, 22},
	{"itemStatusStr", FieldSchema{Type: FieldTypeText, Label: "商品状态"}, 23},
	{"description", FieldSchema{Type: FieldTypeText, Label: "详细描述"}, 24},
}

// Product 商品信息（优化后保留24个核心字段）
type Product struct {
	// ==================== 基本信息 ====================
	ItemID        string `json:"itemId"`
	Title         string `json:"title"`
	SubTitle      string `json:"subTitle,omitempty"`
	Price         string `json:"price"`
	OriginalPrice string `json:"originalPrice"`
	Condition     string `json:"condition,omitempty"` // 成色

	// ==================== 热度指标 ====================
	WantCnt      int `json:"wantCnt"`
	ViewCount    int `json:"viewCount,omitempty"`    // 浏览次数
	CollectCount int `json:"collectCount,omitempty"` // 收藏次数
	ExposureHeat int `json:"exposureHeat,omitempty"` // 曝光热度

	// ==================== 卖家信息 ====================
	SellerNick      string `json:"sellerNick"`
	SellerCity      string `json:"sellerCity"`
	SellerCredit    string `json:"sellerCredit,omitempty"`    // 卖家信用
	SellerItemCount int    `json:"sellerItemCount,omitempty"` // 在售商品数
	SellerSoldCount int    `json:"sellerSoldCount,omitempty"` // 已售数量
	FreeShip        string `json:"freeShip"`

	// ==================== 时间信息 ====================
	PublishTimeMs int64 `json:"publishTimeMs,omitempty"` // 发布时间戳
	CaptureTimeMs int64 `json:"captureTimeMs,omitempty"` // 采集时间戳

	// ==================== 链接资源 ====================
	CoverURL  string `json:"coverUrl"`
	DetailURL string `json:"detailUrl"`
	VideoURL  string `json:"videoUrl,omitempty"` // 视频URL

	// ==================== 其他 ====================
	Tags          string `json:"tags"`
	ItemStatusStr string `json:"itemStatusStr,omitempty"` // 商品状态
	Description   string `json:"description,omitempty"`   // 详细描述
}

// PushToBitableRequest 推送到飞书多维表格请求
type PushToBitableRequest struct {
	Date     string    `json:"date"`     // 日期，用于数据表命名或筛选
	Products []Product `json:"products"` // 商品列表
}

// PushToBitableResponse 推送响应
type PushToBitableResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		RecordsCreated int    `json:"recordsCreated"`
		RecordsUpdated int    `json:"recordsUpdated"`
		TableToken     string `json:"tableToken"`
	} `json:"data,omitempty"`
}

// NewProduct 从 FeedItem 创建 Product（用于从 mtop.FeedItem 转换）
func NewProduct(itemID, title, price, originalPrice string, wantCnt int, publishTime, sellerNick, sellerCity, coverURL, detailURL string) Product {
	now := time.Now()
	return Product{
		ItemID:        itemID,
		Title:         title,
		Price:         price,
		OriginalPrice: originalPrice,
		WantCnt:       wantCnt,
		CaptureTimeMs: now.UnixMilli(),
		SellerNick:    sellerNick,
		SellerCity:    sellerCity,
		CoverURL:      coverURL,
		DetailURL:     detailURL,
	}
}

// NewProductWithTimestamps 创建带时间戳的 Product
func NewProductWithTimestamps(itemID, title, price, originalPrice string, wantCnt int, publishTime int64, sellerNick, sellerCity, coverURL, detailURL string) Product {
	now := time.Now()
	return Product{
		ItemID:        itemID,
		Title:         title,
		Price:         price,
		OriginalPrice: originalPrice,
		WantCnt:       wantCnt,
		PublishTimeMs: publishTime,
		CaptureTimeMs: now.UnixMilli(),
		SellerNick:    sellerNick,
		SellerCity:    sellerCity,
		CoverURL:      coverURL,
		DetailURL:     detailURL,
	}
}

// GetFieldMapping 获取字段映射（Key -> FieldSchema）
func GetFieldMapping() map[string]FieldSchema {
	mapping := make(map[string]FieldSchema)
	for _, field := range ProductFields {
		mapping[field.Key] = field.Schema
	}
	return mapping
}

// GetFieldNameMapping 获取字段名映射（Key -> 飞书显示字段名）
func GetFieldNameMapping() map[string]string {
	mapping := make(map[string]string)
	for _, field := range ProductFields {
		// 直接使用 Key 作为字段名（英文标识符）
		mapping[field.Key] = field.Key
	}
	return mapping
}
