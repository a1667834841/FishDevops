package mtop

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FeedItem 猜你喜欢商品项
type FeedItem struct {
	ItemID        string   `json:"itemId"`
	Title         string   `json:"title"`
	Price         string   `json:"price"`
	PriceOriginal string   `json:"priceOriginal"`
	ImageURL      string   `json:"picUrl"`
	ShopName      string   `json:" "`
	Location      string   `json:"location"`
	WantCount     int      `json:"wantCount"`
	IsIdle        bool     `json:"isIdle"`
	VideoCoverURL string   `json:"videoCoverUrl"`
	VideoURL      string   `json:"videoUrl"`
	Condition     string   `json:"condition"`
	PublishTime   string   `json:"publishTime"`
	SoldOut       bool     `json:"soldOut"`
	Like          bool     `json:"like"`
	Tags          []string `json:"tags"` // 商品标签
}

// GuessYouLikeRequest 猜你喜欢请求参数
type GuessYouLikeRequest struct {
	ItemID     string `json:"itemId"`
	MachID     string `json:"machId"`
	PageNumber int    `json:"pageNumber"`
	PageSize   int    `json:"pageSize"`
}

// GuessYouLikeOptions 获取猜你喜欢的选项
type GuessYouLikeOptions struct {
	MaxPages  int  // 最大爬取页数
	StartPage int  // 起始页
}

// GuessYouLike 获取猜你喜欢商品列表
// machId: 推荐码/机器ID（可选，用于个性化推荐）
// totalPages: 爬取页数
func (c *Client) GuessYouLike(machID string, totalPages int, opts ...GuessYouLikeOptions) ([]FeedItem, error) {
	options := GuessYouLikeOptions{
		MaxPages:  totalPages,
		StartPage: 1,
	}
	if len(opts) > 0 {
		options = opts[0]
	}

	var allItems []FeedItem

	for page := options.StartPage; page <= options.MaxPages; page++ {
		reqData := GuessYouLikeRequest{
			ItemID:     "",
			MachID:     machID,
			PageNumber: page,
			PageSize:   30,
		}

		resp, err := c.Do(Request{
			API:    "mtop.taobao.idlehome.home.webpc.feed",
			Data:   reqData,
			Method: "POST",
		})
		if err != nil {
			return nil, fmt.Errorf("第 %d 页请求失败: %w", page, err)
		}

		// 检查返回状态
		success := false
		for _, r := range resp.Ret {
			if r == "SUCCESS::调用成功" || r == "SUCCESS" {
				success = true
				break
			}
		}
		if !success {
			return nil, fmt.Errorf("第 %d 页返回错误: ret=%v", page, resp.Ret)
		}

		// 解析数据 - resp.Data 已经是 data 字段的原始内容
		var feedData struct {
			CardList  []json.RawMessage `json:"cardList"`
			FeedsCount int              `json:"feedsCount"`
			NextPage   bool              `json:"nextPage"`
			ServerTime string            `json:"serverTime"`
		}
		if err := json.Unmarshal(resp.Data, &feedData); err != nil {
			return nil, fmt.Errorf("解析第 %d 页数据失败: %w", page, err)
		}

		fmt.Printf("[调试] cardList 数量: %d\n", len(feedData.CardList))

		// 解析每个 card
		for _, cardBytes := range feedData.CardList {
			// fmt.Printf("[调试] 解析第 %d 个卡片\n", idx)
			var card struct {
				CardData struct {
					DetailParams struct {
						ItemID       string `json:"itemId"`
						PicUrl       string `json:"picUrl"`
						Title        string `json:"title"`
						UserNick     string `json:"userNick"`
						UserAvatarUrl string `json:"userAvatarUrl"`
						SoldPrice    string `json:"soldPrice"`
						IsVideo      string `json:"isVideo"`
					} `json:"detailParams"`
					PriceInfo struct {
						Price    string `json:"price"`
						OriPrice string `json:"oriPrice"`
					} `json:"priceInfo"`
					HotPoint struct {
						Text string `json:"text"`
					} `json:"hotPoint"`
					Images []struct {
						Url string `json:"url"`
					} `json:"images"`
					RedirectUrl string `json:"redirectUrl"`
					City        string `json:"city"`
					ItemId      string `json:"itemId"`
					AttributeMap map[string]string `json:"attributeMap"`
					FishTags    map[string]struct {
						TagList []struct {
							Data struct {
								Content string `json:"content"`
							} `json:"data"`
						} `json:"tagList"`
					} `json:"fishTags"`
				} `json:"cardData"`
			}
			if err := json.Unmarshal(cardBytes, &card); err != nil {
				fmt.Printf("[调试] 解析卡片失败: %v\n", err)
				fmt.Printf("[调试] 卡片原始数据: %s\n", string(cardBytes))
				continue // 跳过无法解析的卡片
			}

			// 转换为 FeedItem
			item := FeedItem{
				ItemID:    card.CardData.DetailParams.ItemID,
				Title:     card.CardData.DetailParams.Title,
				Price:     card.CardData.PriceInfo.Price,
				PriceOriginal: card.CardData.PriceInfo.OriPrice,
				ImageURL:  card.CardData.DetailParams.PicUrl,
				Location:  card.CardData.City,
				WantCount: 0,
				Tags:      []string{},
			}

			// 解析想要人数和商品标签（优先从 fishTags 解析）
			tagSet := make(map[string]bool)
			for _, region := range card.CardData.FishTags {
				for _, tag := range region.TagList {
					content := tag.Data.Content
					if content == "" {
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

			// 备用：从 hotPoint 解析想要人数（如果 fishTags 中没有）
			if item.WantCount == 0 && card.CardData.HotPoint.Text != "" {
				fmt.Sscanf(card.CardData.HotPoint.Text, "%d人想要", &item.WantCount)
			}

			// 解析发布时间（从 attributeMap.gmtShelf 获取毫秒时间戳）
			if gmtShelf, ok := card.CardData.AttributeMap["gmtShelf"]; ok {
				if ms, err := strconv.ParseInt(gmtShelf, 10, 64); err == nil && ms > 0 {
					// 转换为本地时间字符串
					item.PublishTime = time.Unix(ms/1000, 0).Format("2006-01-02 15:04:05")
				}
			}

			// fmt.Printf("[调试] 解析成功: %s - %s\n", item.Title, item.Price)
			allItems = append(allItems, item)
		}

		// 如果没有下一页，提前结束
		if !feedData.NextPage {
			break
		}
	}

	return allItems, nil
}

// PrintGuessYouLike 打印猜你喜欢商品信息
func PrintGuessYouLike(items []FeedItem) {
	fmt.Printf("\n========== 猜你喜欢 (%d 条) ==========\n", len(items))
	for i, item := range items {
		fmt.Printf("\n[%d] %s\n", i+1, item.Title)
		fmt.Printf("    商品ID: %s\n", item.ItemID)
		fmt.Printf("    价格: %s", item.Price)
		if item.PriceOriginal != "" && item.PriceOriginal != item.Price {
			fmt.Printf(" (原价: %s)", item.PriceOriginal)
		}
		fmt.Println()
		if item.Location != "" {
			fmt.Printf("    地区: %s\n", item.Location)
		}
		if item.Condition != "" {
			fmt.Printf("    成色: %s\n", item.Condition)
		}
		if item.WantCount > 0 {
			fmt.Printf("    想要: %d人\n", item.WantCount)
		}
		if item.SoldOut {
			fmt.Printf("    状态: 已售出\n")
		}
	}
	fmt.Printf("\n===================================\n")
}

// SaveGuessYouLikeToFile 保存猜你喜欢到文件
func SaveGuessYouLikeToFile(items []FeedItem, filename string) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return saveToFile(filename, data)
}

// saveToFile 辅助函数：保存到文件
func saveToFile(filename string, data []byte) error {
	// 简单实现，可以使用 os.WriteFile
	return fmt.Errorf("not implemented")
}
