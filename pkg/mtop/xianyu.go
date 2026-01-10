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
	// 基础信息
	ItemID       string `json:"itemId"`       // 商品ID
	Title        string `json:"title"`        // 商品标题
	ImageURL     string `json:"picUrl"`       // 图片链接
	CategoryID   int `json:"categoryId"`   // 叶子分类ID
	Location     string `json:"location"`     // 所在城市

	// 价格与行情
	Price         string `json:"price"`         // 当前售价
	PriceOriginal string `json:"priceOriginal"` // 原价
	UnitPrice     string `json:"unitPrice"`     // 单位价格

	// 热度与流量
	WantCount int    `json:"wantCount"` // 想要人数
	ViewCount int    `json:"viewCount"` // 浏览人数
	Status    string `json:"status"`    // 商品状态
	ShopLevel string `json:"shopLevel"` // 店铺级别

	// 卖家与服务
	SellerNick   string `json:"sellerNick"`   // 卖家名字
	SellerCredit string `json:"sellerCredit"` // 卖家信用
	FreeShipping bool   `json:"freeShipping"` // 是否包邮

	// 时间与活跃度
	PublishTime    string `json:"publishTime"`    // 发布时间
	PublishTimeTS  int64  `json:"publishTimeTs"`  // 发布时间戳（毫秒）
	ModifiedTime   string `json:"modifiedTime"`   // 下架/修改时间
	ModifiedTimeTS int64  `json:"modifiedTimeTs"` // 下架时间戳（毫秒）
	ProPolishTime  string `json:"proPolishTime"`  // 最近一次擦亮时间
	ProPolishTimeTS int64 `json:"proPolishTimeTs"` // 擦亮时间戳（毫秒）

	// 其他字段
	IsIdle        bool     `json:"isIdle"`        // 是否闲置
	VideoCoverURL string   `json:"videoCoverUrl"` // 视频封面
	VideoURL      string   `json:"videoUrl"`      // 视频URL
	Condition     string   `json:"condition"`     // 成色
	SoldOut       bool     `json:"soldOut"`       // 是否已售出
	Like          bool     `json:"like"`          // 是否收藏
	Tags          []string `json:"tags"`          // 商品标签
	IsVideo       bool     `json:"isVideo"`       // 是否视频
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
	MaxPages        int   // 最大爬取页数
	StartPage       int   // 起始页
	MinWantCount    int   // 最低想要人数（0表示不限制）
	DaysWithin      int   // 发布时间范围（天数，0表示不限制，默认7天）
}

// GuessYouLike 获取猜你喜欢商品列表
// machId: 推荐码/机器ID（可选，用于个性化推荐）
// totalPages: 爬取页数
// opts: 可选参数（过滤条件等）
func (c *Client) GuessYouLike(machID string, totalPages int, opts ...GuessYouLikeOptions) ([]FeedItem, error) {
	options := GuessYouLikeOptions{
		MaxPages:     totalPages,
		StartPage:    1,
		MinWantCount: 0,
		DaysWithin:   14, // 默认近7天
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
		for idx, cardBytes := range feedData.CardList {
			fmt.Printf("[调试] 解析第 %d 个卡片\n", idx)
			var card struct {
				CardData struct {
					CategoryID  int `json:"categoryId"`
					Status      string `json:"status"`
					ViewCount   int    `json:"viewCount"`
					DetailParams struct {
						ItemID       string `json:"itemId"`
						PicUrl       string `json:"picUrl"`
						Title        string `json:"title"`
						UserNick     string `json:"userNick"`
						UserAvatarUrl string `json:"userAvatarUrl"`
						SoldPrice    string `json:"soldPrice"`
						IsVideo      string `json:"isVideo"`
					} `json:"detailParams"`
					User struct {
						UserNick string `json:"userNick"`
					} `json:"user"`
					PriceInfo struct {
						Price    string `json:"price"`
						OriPrice string `json:"oriPrice"`
					} `json:"priceInfo"`
					UnitPriceInfo struct {
						Price string `json:"price"`
					} `json:"unitPriceInfo"`
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
								LabelId string `json:"labelId"`
								Type    string `json:"type"`
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
				fmt.Printf("[调试] 解析卡片失败: %v\n", err)
				fmt.Printf("[调试] 卡片原始数据: %s\n", string(cardBytes))
				continue // 跳过无法解析的卡片
			}

			// 转换为 FeedItem
			item := FeedItem{
				ItemID:       card.CardData.DetailParams.ItemID,
				Title:        card.CardData.DetailParams.Title,
				Price:        card.CardData.PriceInfo.Price,
				PriceOriginal: card.CardData.PriceInfo.OriPrice,
				UnitPrice:    card.CardData.UnitPriceInfo.Price,
				ImageURL:     card.CardData.DetailParams.PicUrl,
				CategoryID:   card.CardData.CategoryID,
				Location:     card.CardData.City,
				SellerNick:   card.CardData.User.UserNick,
				WantCount:    0,
				ViewCount:    card.CardData.ViewCount,
				Status:       card.CardData.Status,
				IsVideo:      card.CardData.DetailParams.IsVideo == "1",
				Tags:         []string{},
				ShopLevel:    "",
				SellerCredit: "",
				FreeShipping: false,
			}

			// 解析想要人数、商品标签、店铺级别、卖家信用等（优先从 fishTags 解析）
			tagSet := make(map[string]bool)
			for _, region := range card.CardData.FishTags {
				for _, tag := range region.TagList {
					content := tag.Data.Content
					if content == "" {
						continue
					}

					// 检查是否为店铺级别（从 utParams.data.content 或 content 中检查）
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

					// 检查是否为卖家信用（包含"信用"关键词）
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

			// 备用：从 hotPoint 解析想要人数（如果 fishTags 中没有）
			if item.WantCount == 0 && card.CardData.HotPoint.Text != "" {
				fmt.Sscanf(card.CardData.HotPoint.Text, "%d人想要", &item.WantCount)
			}

			// 解析发布时间（从 attributeMap.gmtShelf 获取毫秒时间戳）
			if gmtShelf, ok := card.CardData.AttributeMap["gmtShelf"]; ok {
				if ms, err := strconv.ParseInt(gmtShelf, 10, 64); err == nil && ms > 0 {
					item.PublishTimeTS = ms
					// 转换为本地时间字符串
					item.PublishTime = time.Unix(ms/1000, 0).Format("2006-01-02 15:04:05")
				}
			}

			// 解析下架/修改时间（从 attributeMap.gmtModified 获取毫秒时间戳）
			if gmtModified, ok := card.CardData.AttributeMap["gmtModified"]; ok {
				if ms, err := strconv.ParseInt(gmtModified, 10, 64); err == nil && ms > 0 {
					item.ModifiedTimeTS = ms
					// 转换为本地时间字符串
					item.ModifiedTime = time.Unix(ms/1000, 0).Format("2006-01-02 15:04:05")
				}
			}

			// 解析是否包邮（从 attributeMap.freeShipping）
			if freeShipping, ok := card.CardData.AttributeMap["freeShipping"]; ok && freeShipping == "1" {
				item.FreeShipping = true
			}

			// 解析擦亮时间（从 attributeMap.proPolishTime 获取毫秒时间戳）
			if proPolishTime, ok := card.CardData.AttributeMap["proPolishTime"]; ok {
				if ms, err := strconv.ParseInt(proPolishTime, 10, 64); err == nil && ms > 0 {
					item.ProPolishTimeTS = ms
					// 转换为本地时间字符串
					item.ProPolishTime = time.Unix(ms/1000, 0).Format("2006-01-02 15:04:05")
				}
			}

			fmt.Printf("[调试] 解析成功: %s - %s, 店铺: %s, 信用: %s, 包邮: %v\n",
				item.Title, item.Price, item.ShopLevel, item.SellerCredit, item.FreeShipping)

			// 应用过滤条件
			if !options.MatchFilter(item) {
				fmt.Printf("[过滤] 跳过商品: %s (想要:%d, 发布:%s)\n", item.Title, item.WantCount, item.PublishTime)
				continue
			}

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

		// 价格信息
		fmt.Printf("    价格: %s", item.Price)
		if item.PriceOriginal != "" && item.PriceOriginal != item.Price {
			fmt.Printf(" (原价: %s)", item.PriceOriginal)
		}
		if item.UnitPrice != "" {
			fmt.Printf(", 单价: %s", item.UnitPrice)
		}
		fmt.Println()

		// 店铺和卖家信息
		if item.ShopLevel != "" {
			fmt.Printf("    店铺级别: %s", item.ShopLevel)
		}
		if item.SellerCredit != "" {
			fmt.Printf(" | 信用: %s", item.SellerCredit)
		}
		if item.SellerNick != "" {
			fmt.Printf(" | 卖家: %s", item.SellerNick)
		}
		if item.ShopLevel != "" || item.SellerCredit != "" || item.SellerNick != "" {
			fmt.Println()
		}

		// 服务信息
		serviceInfo := []string{}
		if item.FreeShipping {
			serviceInfo = append(serviceInfo, "包邮")
		}
		if item.IsVideo {
			serviceInfo = append(serviceInfo, "视频")
		}
		if len(serviceInfo) > 0 {
			fmt.Printf("    服务: %s\n", strings.Join(serviceInfo, ", "))
		}

		// 热度信息
		hotInfo := []string{}
		if item.WantCount > 0 {
			hotInfo = append(hotInfo, fmt.Sprintf("%d人想要", item.WantCount))
		}
		if item.ViewCount > 0 {
			hotInfo = append(hotInfo, fmt.Sprintf("%d人浏览", item.ViewCount))
		}
		if len(hotInfo) > 0 {
			fmt.Printf("    热度: %s\n", strings.Join(hotInfo, ", "))
		}

		// 位置信息
		if item.Location != "" {
			fmt.Printf("    地区: %s\n", item.Location)
		}

		// 分类信息
		if item.CategoryID != 0 {
			fmt.Printf("    分类ID: %s\n", item.CategoryID)
		}

		// 时间信息
		timeInfo := []string{}
		if item.PublishTime != "" {
			timeInfo = append(timeInfo, fmt.Sprintf("发布:%s", item.PublishTime))
		}
		if item.ModifiedTime != "" {
			timeInfo = append(timeInfo, fmt.Sprintf("修改:%s", item.ModifiedTime))
		}
		if len(timeInfo) > 0 {
			fmt.Printf("    时间: %s\n", strings.Join(timeInfo, ", "))
		}

		// 状态信息
		if item.Status != "" {
			fmt.Printf("    状态: %s\n", item.Status)
		}

		// 标签
		if len(item.Tags) > 0 {
			fmt.Printf("    标签: %s\n", strings.Join(item.Tags, ", "))
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

// MatchFilter 检查商品是否匹配过滤条件
func (o GuessYouLikeOptions) MatchFilter(item FeedItem) bool {
	// 检查最低想要人数
	if o.MinWantCount > 0 && item.WantCount < o.MinWantCount {
		return false
	}

	// 检查发布时间范围
	if o.DaysWithin > 0 && item.PublishTimeTS > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -o.DaysWithin).UnixMilli()
		if item.PublishTimeTS < cutoffTime {
			return false
		}
	}

	return true
}
