package feishu

import "time"

// FieldType 飞书字段类型
type FieldType int

const (
	FieldTypeText    FieldType = 1  // 文本
	FieldTypeNumber  FieldType = 2  // 数字
	FieldTypeSingleSelect FieldType = 3  // 单选
	FieldTypeMultiSelect FieldType = 4  // 多选
	FieldTypeDateTime   FieldType = 5  // 日期时间
	FieldTypeCheckbox   FieldType = 7  // 复选框
	FieldTypeURL        FieldType = 15 // URL
	FieldTypePhone      FieldType = 11 // 电话
	FieldTypeEmail      FieldType = 13 // 邮箱
)

// FieldSchema 字段定义
type FieldSchema struct {
	Type     FieldType `json:"type"`
	Label    string    `json:"label"`
	FieldName string   `json:"fieldName,omitempty"`
}

// ProductSchema 商品字段定义（对应用户提供的 PRODUCT_SCHEMA）
var ProductFields = []struct {
	Key      string
	Schema   FieldSchema
	CSVOrder int // CSV 导出顺序，0 表示不导出
}{
	{"itemId", FieldSchema{Type: FieldTypeText, Label: "商品ID"}, 1},
	{"title", FieldSchema{Type: FieldTypeText, Label: "商品标题"}, 2},
	{"price", FieldSchema{Type: FieldTypeText, Label: "价格"}, 3},
	{"priceNumber", FieldSchema{Type: FieldTypeNumber, Label: "价格数值"}, 0},
	{"originalPrice", FieldSchema{Type: FieldTypeText, Label: "原价"}, 4},
	{"originalPriceNumber", FieldSchema{Type: FieldTypeNumber, Label: "原价数值"}, 0},
	{"wantCnt", FieldSchema{Type: FieldTypeNumber, Label: "想要人数"}, 5},
	{"publishTime", FieldSchema{Type: FieldTypeText, Label: "发布时间"}, 6},
	{"publishTimeMs", FieldSchema{Type: FieldTypeDateTime, Label: "发布时间戳"}, 0},
	{"captureTime", FieldSchema{Type: FieldTypeText, Label: "采集时间"}, 0},
	{"captureTimeMs", FieldSchema{Type: FieldTypeDateTime, Label: "采集时间戳"}, 0},
	{"sellerNick", FieldSchema{Type: FieldTypeText, Label: "卖家昵称"}, 7},
	{"sellerCity", FieldSchema{Type: FieldTypeText, Label: "地区"}, 8},
	{"freeShip", FieldSchema{Type: FieldTypeText, Label: "包邮"}, 9},
	{"tags", FieldSchema{Type: FieldTypeText, Label: "商品标签"}, 10},
	{"coverUrl", FieldSchema{Type: FieldTypeURL, Label: "封面URL"}, 11},
	{"detailUrl", FieldSchema{Type: FieldTypeURL, Label: "商品详情URL"}, 12},
	{"exposureHeat", FieldSchema{Type: FieldTypeNumber, Label: "曝光热度"}, 0},
	{"proPolishTime", FieldSchema{Type: FieldTypeText, Label: "最近擦亮时间"}, 13},
	{"proPolishTimeMs", FieldSchema{Type: FieldTypeDateTime, Label: "擦亮时间戳"}, 0},
}

// Product 商品信息（对应 FeedItem，增加采集时间等字段）
type Product struct {
	ItemID            string    `json:"itemId"`
	Title             string    `json:"title"`
	Price             string    `json:"price"`
	PriceNumber       float64   `json:"priceNumber,omitempty"`
	OriginalPrice     string    `json:"originalPrice"`
	OriginalPriceNumber float64 `json:"originalPriceNumber,omitempty"`
	WantCnt           int       `json:"wantCnt"`
	PublishTime       string    `json:"publishTime"`
	PublishTimeMs     int64     `json:"publishTimeMs,omitempty"`
	CaptureTime       string    `json:"captureTime"`
	CaptureTimeMs     int64     `json:"captureTimeMs,omitempty"`
	SellerNick        string    `json:"sellerNick"`
	SellerCity        string    `json:"sellerCity"`
	FreeShip          string    `json:"freeShip"`
	Tags              string    `json:"tags"`
	CoverURL          string    `json:"coverUrl"`
	DetailURL         string    `json:"detailUrl"`
	ExposureHeat      int       `json:"exposureHeat,omitempty"`
	ProPolishTime     string    `json:"proPolishTime,omitempty"`
	ProPolishTimeMs   int64     `json:"proPolishTimeMs,omitempty"`
}

// PushToBitableRequest 推送到飞书多维表格请求
type PushToBitableRequest struct {
	Date     string     `json:"date"`      // 日期，用于数据表命名或筛选
	Products []Product  `json:"products"`  // 商品列表
}

// PushToBitableResponse 推送响应
type PushToBitableResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		RecordsCreated int `json:"recordsCreated"`
		RecordsUpdated int `json:"recordsUpdated"`
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
		PublishTime:   publishTime,
		CaptureTime:   now.Format("2006-01-02 15:04:05"),
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
		CaptureTime:   now.Format("2006-01-02 15:04:05"),
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
		if field.Schema.FieldName != "" {
			mapping[field.Key] = field.Schema.FieldName
		} else {
			mapping[field.Key] = field.Schema.Label
		}
	}
	return mapping
}
