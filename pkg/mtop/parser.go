package mtop

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseFeedResponse 解析 Feed API 响应
func ParseFeedResponse(resp *Response) ([]FeedItem, bool, error) {
	var feedData struct {
		CardList   []json.RawMessage `json:"cardList"`
		FeedsCount int               `json:"feedsCount"`
		NextPage   bool              `json:"nextPage"`
		ServerTime string            `json:"serverTime"`
	}
	if err := json.Unmarshal(resp.Data, &feedData); err != nil {
		return nil, false, fmt.Errorf("解析数据失败: %w", err)
	}

	items := make([]FeedItem, 0, len(feedData.CardList))
	for _, cardBytes := range feedData.CardList {
		item, err := ParseCardToFeedItem(cardBytes)
		if err != nil {
			// 跳过无法解析的卡片
			continue
		}
		if item.ItemID != "" { // 过滤空 ID
			items = append(items, item)
		}
	}

	return items, feedData.NextPage, nil
}

// ParseCardToFeedItem 解析卡片数据为 FeedItem
func ParseCardToFeedItem(cardBytes json.RawMessage) (FeedItem, error) {
	var card struct {
		CardData struct {
			CategoryID   int    `json:"categoryId"`
			Status       string `json:"status"`
			ViewCount    int    `json:"viewCount"`
			DetailParams struct {
				ItemID   string `json:"itemId"`
				PicUrl   string `json:"picUrl"`
				Title    string `json:"title"`
				UserNick string `json:"userNick"`
				IsVideo  string `json:"isVideo"`
			} `json:"detailParams"`
			User struct {
				UserNick string `json:"userNick"`
			} `json:"user"`
			PriceInfo struct {
				Price string `json:"price"`
			} `json:"priceInfo"`
			HotPoint struct {
				Text string `json:"text"`
			} `json:"hotPoint"`
			City         string            `json:"city"`
			AttributeMap map[string]string `json:"attributeMap"`
			FishTags     map[string]struct {
				TagList []struct {
					Data struct {
						Content string `json:"content"`
					} `json:"data"`
					UtParams *struct {
						Data *struct {
							Content string `json:"content"`
						} `json:"args"`
					} `json:"utParams"`
				} `json:"tagList"`
			} `json:"fishTags"`
		} `json:"cardData"`
	}
	if err := json.Unmarshal(cardBytes, &card); err != nil {
		return FeedItem{}, err
	}

	item := FeedItem{
		ItemID:     card.CardData.DetailParams.ItemID,
		Title:      card.CardData.DetailParams.Title,
		Price:      card.CardData.PriceInfo.Price,
		ImageURL:   card.CardData.DetailParams.PicUrl,
		CategoryID: card.CardData.CategoryID,
		Location:   card.CardData.City,
		SellerNick: card.CardData.User.UserNick,
		WantCount:  0,
		ViewCount:  card.CardData.ViewCount,
		Status:     card.CardData.Status,
		IsVideo:    card.CardData.DetailParams.IsVideo == "1",
		Tags:       []string{},
	}

	// 提取标签信息
	tagSet := make(map[string]bool)
	for _, region := range card.CardData.FishTags {
		for _, tag := range region.TagList {
			content := tag.Data.Content
			if content == "" {
				continue
			}

			// 检查是否为店铺级别
			shopLevel := ""
			if tag.UtParams != nil && tag.UtParams.Data != nil {
				if strings.Contains(tag.UtParams.Data.Content, "level") {
					shopLevel = tag.UtParams.Data.Content
				}
			}
			if shopLevel == "" && strings.Contains(content, "level") {
				shopLevel = content
			}
			if shopLevel != "" {
				item.ShopLevel = shopLevel
				tagSet[shopLevel] = true
				continue
			}

			// 检查是否为卖家信用
			if strings.Contains(content, "信用") {
				item.SellerCredit = content
				tagSet[content] = true
				continue
			}

			// 解析想要人数
			if strings.HasSuffix(content, "人想要") {
				numStr := strings.TrimSuffix(content, "人想要")
				if num, err := strconv.Atoi(numStr); err == nil {
					item.WantCount = num
				}
				continue
			}

			// 处理商品标签
			tagContent := content
			if strings.Contains(content, "freeShippingIcon") {
				tagContent = "包邮"
			}
			if tagContent != "" {
				tagSet[tagContent] = true
			}
		}
	}

	// 转换为 Tags 切片
	for tag := range tagSet {
		item.Tags = append(item.Tags, tag)
	}

	// 备用：从 hotPoint 解析想要人数
	if item.WantCount == 0 && card.CardData.HotPoint.Text != "" {
		fmt.Sscanf(card.CardData.HotPoint.Text, "%d人想要", &item.WantCount)
	}

	// 提取时间戳信息
	if gmtShelf, ok := card.CardData.AttributeMap["gmtShelf"]; ok {
		if ms, err := strconv.ParseInt(gmtShelf, 10, 64); err == nil && ms > 0 {
			item.PublishTimeTS = ms
		}
	}

	if gmtModified, ok := card.CardData.AttributeMap["gmtModified"]; ok {
		if ms, err := strconv.ParseInt(gmtModified, 10, 64); err == nil && ms > 0 {
			item.ModifiedTimeTS = ms
		}
	}

	if freeShipping, ok := card.CardData.AttributeMap["freeShipping"]; ok && freeShipping == "1" {
		item.FreeShipping = true
	}

	if proPolishTime, ok := card.CardData.AttributeMap["proPolishTime"]; ok {
		if ms, err := strconv.ParseInt(proPolishTime, 10, 64); err == nil && ms > 0 {
			item.ProPolishTimeTS = ms
		}
	}

	return item, nil
}
