package service

import (
	"fmt"
	"time"

	"xianyu_aner/pkg/feishu"
	"xianyu_aner/pkg/mtop"
	"xianyu_aner/pkg/util"
)

// Converter 数据转换服务（通用）
type Converter struct{}

// NewConverter 创建转换服务
func NewConverter() *Converter {
	return &Converter{}
}

// FeedItemToBasicProduct 将 FeedItem 转换为基础产品（用于去重）
func (c *Converter) FeedItemToBasicProduct(item mtop.FeedItem) feishu.Product {
	now := time.Now()
	return feishu.Product{
		ItemID:        item.ItemID,
		Price:         item.Price,
		WantCnt:       item.WantCount,
		PublishTimeMs: item.PublishTimeTS,
		CaptureTimeMs: now.UnixMilli(),
		Title:         item.Title,
		CoverURL:      item.ImageURL,
		DetailURL:     BuildDetailURL(item.ItemID),
	}
}

// FeedItemsToBasicProducts 将多个 FeedItem 转换为基础产品（用于去重）
func (c *Converter) FeedItemsToBasicProducts(items []mtop.FeedItem) []feishu.Product {
	products := make([]feishu.Product, 0, len(items))
	for _, item := range items {
		products = append(products, c.FeedItemToBasicProduct(item))
	}
	return products
}

// ItemDetailToProduct 将 ItemDetail 转换为完整产品
func (c *Converter) ItemDetailToProduct(detail *mtop.ItemDetail) feishu.Product {
	now := time.Now()
	return feishu.Product{
		ItemID:         detail.ItemID,
		Title:          detail.Title,
		Price:          detail.Price,
		WantCnt:        detail.WantCount,
		PublishTimeMs:  detail.PublishTimeTS,
		CaptureTimeMs:  now.UnixMilli(),
		ViewCount:      detail.ViewCount,
		CollectCount:   detail.CollectCount,
		SellerNick:     detail.SellerNick,
		SellerCity:     detail.SellerCity,
		SellerCredit:   detail.SellerCredit,
		FreeShip:       util.BoolToYesNo(detail.FreeShipping),
		Tags:           util.StringsJoin(detail.Tags, ", "),
		Condition:      detail.Condition,
		Description:    detail.Description,
		VideoURL:       detail.VideoURL,
		CoverURL:       detail.ImageURL,
		DetailURL:      BuildDetailURL(detail.ItemID),
	}
}

// MergeDetailToProduct 将详情合并到基础产品
func (c *Converter) MergeDetailToProduct(basic feishu.Product, detail *mtop.ItemDetail) feishu.Product {
	result := basic
	result.ViewCount = detail.ViewCount
	result.CollectCount = detail.CollectCount
	result.Condition = detail.Condition
	result.SellerNick = detail.SellerNick
	result.SellerCity = detail.SellerCity
	result.SellerCredit = detail.SellerCredit
	result.FreeShip = util.BoolToYesNo(detail.FreeShipping)
	result.Tags = util.StringsJoin(detail.Tags, ", ")
	result.Description = detail.Description
	result.VideoURL = detail.VideoURL
	if detail.ImageURL != "" {
		result.CoverURL = detail.ImageURL
	}
	return result
}

// BuildDetailURL 构建商品详情URL
func BuildDetailURL(itemID string) string {
	return fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", itemID)
}

// Result 执行结果
type Result struct {
	TotalItems int
	Duration   time.Duration
}
