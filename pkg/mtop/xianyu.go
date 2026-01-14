package mtop

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ==================== 商品详情 API ====================

// ItemDetailRequest 商品详情请求参数
type ItemDetailRequest struct {
	ItemID string `json:"itemId"`
}

// ==================== SKU 相关结构体 ====================

// SKUProperty SKU属性
type SKUProperty struct {
	PropertyID      int64  `json:"propertyId"`      // 属性ID
	PropertyText    string `json:"propertyText"`    // 属性名，如 "重量"
	ValueID         int64  `json:"valueId"`         // 属性值ID
	ValueText       string `json:"valueText"`       // 属性值，如 "多段式伸缩带 透明"
	ActualValueText string `json:"actualValueText"` // 完整属性值文本
}

// SKU 商品规格
type SKU struct {
	SKUID        int64         `json:"skuId"`        // SKU ID
	InventoryID  int64         `json:"inventoryId"`  // 库存ID
	Price        int           `json:"price"`        // 价格（分为单位）
	PriceInCent  int           `json:"priceInCent"`  // 价格（分为单位）
	Quantity     int           `json:"quantity"`     // 库存数量
	PropertyList []SKUProperty `json:"propertyList"` // 属性列表
}

// CPVLabel 商品属性标签（如成色）
type CPVLabel struct {
	PropertyID   int64  `json:"propertyId"`   // 属性ID
	PropertyName string `json:"propertyName"` // 属性名，如 "成色"
	ValueID      int64  `json:"valueId"`      // 属性值ID
	ValueName    string `json:"valueName"`    // 属性值，如 "全新"
}

// ItemTag 商品标签
type ItemTag struct {
	ChannelCateID int64  `json:"channelCateId"` // 频道分类ID
	From          string `json:"from"`          // 来源
	Text          string `json:"text"`          // 标签文本
	Properties    string `json:"properties"`    // 属性
}

// ImageInfo 图片信息
type ImageInfo struct {
	URL        string `json:"url"`        // 图片URL
	Major      bool   `json:"major"`      // 是否主图
	WidthSize  int    `json:"widthSize"`  // 宽度
	HeightSize int    `json:"heightSize"` // 高度
}

// ItemDetail 商品详情数据结构
// 数据分析用途：该结构体包含丰富的商品维度信息，可用于价格分析、地域分析、热度分析等
type ItemDetail struct {
	// ==================== 基础信息 ====================
	// [数据分析价值: 高] 商品唯一标识，用于数据关联和去重
	ItemID string `json:"itemId"` // 商品ID
	// [数据分析价值: 中] 可进行文本分析提取关键词（品牌、型号等），或NLP分类
	Title string `json:"title"` // 商品标题
	// [数据分析价值: 低] 补充信息，非核心分析字段
	SubTitle string `json:"subTitle"` // 副标题
	// [数据分析价值: 低] 简短描述，非核心分析字段
	Desc string `json:"desc"` // 商品描述
	// [数据分析价值: 低] URL字段，一般不用于直接数据分析
	ImageURL string `json:"picUrl"` // 主图URL
	// [数据分析价值: 低] URL字段，一般不用于直接数据分析
	VideoURL string `json:"videoUrl"` // 视频URL
	// [数据分析价值: 高] 分类维度，可进行分类统计和趋势分析
	CategoryID int `json:"categoryId"` // 分类ID

	// ==================== 价格信息 ====================
	// [数据分析价值: 高] 核心数值字段，需要解析为float用于价格分布、区间分析
	Price string `json:"price"` // 当前价格（格式: "100.00"）

	// ==================== 卖家信息 ====================
	// [数据分析价值: 高] 卖家唯一标识，可分析卖家活跃度、商品数量分布
	SellerID string `json:"sellerId"` // 卖家ID
	// [数据分析价值: 中] 卖家昵称，可用于文本分析或去重标识
	SellerNick string `json:"sellerNick"` // 卖家昵称
	// [数据分析价值: 低] URL字段，一般不用于直接数据分析
	AvatarURL string `json:"avatarUrl"` // 卖家头像

	// ==================== 商品状态/热度指标 ====================
	// [数据分析价值: 中] 商品状态枚举（online/offline/sold等），可筛选有效数据
	Status string `json:"status"` // 商品状态
	// [数据分析价值: 高] 需求热度指标，可分析受欢迎程度、预测成交概率
	WantCount int `json:"wantCount"` // 想要人数
	// [数据分析价值: 高] 曝光度指标，可计算转化率 = WantCount/ViewCount
	ViewCount int `json:"viewCount"` // 浏览次数
	// [数据分析价值: 中] 收藏热度，辅助指标
	CollectCount int `json:"collectCount"` // 收藏次数

	// ==================== 地址信息 ====================
	// [数据分析价值: 高] 地理位置文本，可解析为省/市进行地域分布分析
	Location string `json:"location"` // 所在城市（格式: "广东深圳"）

	// ==================== 时间信息 ====================
	// [数据分析价值: 低] 字符串格式，不便于直接计算
	PublishTime string `json:"publishTime"` // 发布时间（字符串格式）
	// [数据分析价值: 高] Unix时间戳(毫秒)，核心时间字段，可进行时间序列分析、周期性分析
	PublishTimeTS int64 `json:"publishTimeTs"` // 发布时间戳

	// ==================== 商品属性 ====================
	// [数据分析价值: 中] 成色描述（如"99新"、"95新"），需要标准化处理
	Condition string `json:"condition"` // 成色
	// [数据分析价值: 中] 布尔值，可区分新旧商品类别进行对比分析
	IsNew bool `json:"isNew"` // 是否全新
	// [数据分析价值: 高] 布尔值，包邮是影响价格和转化率的重要因素
	FreeShipping bool `json:"freeShipping"` // 是否包邮
	// [数据分析价值: 中] 标签数组，可提取特征、进行聚类分析
	Tags []string `json:"tags"` // 标签（如: "包邮", "可小刀"）

	// ==================== 媒体资源 ====================
	// [数据分析价值: 低] 图片数量可作为辅助指标（图片数 vs 浏览量）
	ImageList []string `json:"imageList"` // 商品图片列表

	// ==================== 文本内容 ====================
	// [数据分析价值: 中] 长文本，可用于NLP分析提取关键词、情感分析
	Description string `json:"description"` // 详细描述内容

	// ==================== 其他 ====================
	// [数据分析价值: 中] 店铺级别，可作为卖家信誉分析维度
	ShopLevel string `json:"shopLevel"` // 店铺级别
	// [数据分析价值: 中] 卖家芝麻信用等级名称（如 "信用极好"）
	SellerCredit string `json:"sellerCredit"` // 卖家芝麻信用
	// [数据分析价值: 中] 卖家注册天数（需大于0）
	SellerRegDays int `json:"sellerRegDays"` // 卖家注册天数

	// ==================== 新增字段（API 实际返回） ====================
	// 价格相关
	SoldPrice   string `json:"soldPrice"`   // API 原始价格字符串
	PriceInCent int    `json:"priceInCent"` // 价格（分为单位）

	// 库存相关
	TotalStock int `json:"totalStock"` // 总库存

	// 状态相关
	ItemStatus    int    `json:"itemStatus"`    // 商品状态码
	ItemStatusStr string `json:"itemStatusStr"` // 商品状态文本

	// SKU 相关（完整解析）
	HasSKU    bool       `json:"hasSku"`    // 是否有规格
	SKUList   []SKU      `json:"skuList"`   // SKU列表
	CPVLabels []CPVLabel `json:"cpvLabels"` // 属性标签（成色等）
	ItemTags  []ItemTag  `json:"itemTags"`  // 商品标签

	// 卖家扩展信息
	SellerCity      string `json:"sellerCity"`      // 卖家城市
	SellerItemCount int    `json:"sellerItemCount"` // 卖家在售商品数
	SellerSoldCount int    `json:"sellerSoldCount"` // 卖家已售数量
	SellerSignature string `json:"sellerSignature"` // 卖家签名
}

// ItemDetailResponse API响应结构
type ItemDetailResponse struct {
	Item     *ItemDetail `json:"item"`
	Data     interface{} `json:"data"`
	Success  bool        `json:"success"`
	ErrorMsg string      `json:"errorMsg"`
}

// FeedItem 猜你喜欢商品项
type FeedItem struct {
	// 基础信息
	ItemID     string `json:"itemId"`     // 商品ID
	Title      string `json:"title"`      // 商品标题
	ImageURL   string `json:"picUrl"`     // 图片链接
	CategoryID int    `json:"categoryId"` // 叶子分类ID
	Location   string `json:"location"`   // 所在城市

	// 价格与行情
	Price string `json:"price"` // 当前售价

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
	PublishTime     string `json:"publishTime"`     // 发布时间
	PublishTimeTS   int64  `json:"publishTimeTs"`   // 发布时间戳（毫秒）
	ModifiedTime    string `json:"modifiedTime"`    // 下架/修改时间
	ModifiedTimeTS  int64  `json:"modifiedTimeTs"`  // 下架时间戳（毫秒）
	ProPolishTime   string `json:"proPolishTime"`   // 最近一次擦亮时间
	ProPolishTimeTS int64  `json:"proPolishTimeTs"` // 擦亮时间戳（毫秒）

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
	MaxPages     int // 最大爬取页数
	StartPage    int // 起始页
	MinWantCount int // 最低想要人数（0表示不限制）
	DaysWithin   int // 发布时间范围（天数，0表示不限制，默认7天）
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
			CardList   []json.RawMessage `json:"cardList"`
			FeedsCount int               `json:"feedsCount"`
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
					CategoryID   int    `json:"categoryId"`
					Status       string `json:"status"`
					ViewCount    int    `json:"viewCount"`
					DetailParams struct {
						ItemID        string `json:"itemId"`
						PicUrl        string `json:"picUrl"`
						Title         string `json:"title"`
						UserNick      string `json:"userNick"`
						UserAvatarUrl string `json:"userAvatarUrl"`
						SoldPrice     string `json:"soldPrice"`
						IsVideo       string `json:"isVideo"`
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
					RedirectUrl  string            `json:"redirectUrl"`
					City         string            `json:"city"`
					ItemId       string            `json:"itemId"`
					AttributeMap map[string]string `json:"attributeMap"`
					FishTags     map[string]struct {
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

			// fmt.Printf("[调试] 解析成功: %s - %s, 店铺: %s, 信用: %s, 包邮: %v\n",
			// 	item.Title, item.Price, item.ShopLevel, item.SellerCredit, item.FreeShipping)

			// 检查 itemId 是否为空，为空则跳过
			if item.ItemID == "" {
				fmt.Printf("[过滤] 跳过商品: %s (itemId为空)\n", item.Title)
				continue
			}

			// // 应用过滤条件
			// if !options.MatchFilter(item) {
			// 	fmt.Printf("[过滤] 跳过商品: %s (想要:%d, 发布:%s)\n", item.Title, item.WantCount, item.PublishTime)
			// 	continue
			// }

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
		fmt.Printf("    价格: %s\n", item.Price)

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
			fmt.Printf("    分类ID: %d\n", item.CategoryID)
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

// FetchItemDetailWithRetry 带重试机制的商品详情获取
// maxRetries: 最大重试次数
func (c *Client) FetchItemDetailWithRetry(itemID string, maxRetries int) (*ItemDetail, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避：等待一段时间后重试
			waitTime := time.Duration(attempt) * time.Second
			fmt.Printf("[重试 %d/%d] 等待 %v 后重试...\n", attempt, maxRetries, waitTime)
			time.Sleep(waitTime)
		}

		detail, err := c.FetchItemDetail(itemID)
		if err == nil {
			return detail, nil
		}

		lastErr = err

		// 检查是否是限流错误，如果是则重试
		if strings.Contains(err.Error(), "RGV587_ERROR") || strings.Contains(err.Error(), "被挤爆") {
			fmt.Printf("[限流 %d/%d] 遇到限流，将重试...\n", attempt+1, maxRetries)
			continue
		}

		// 如果不是限流错误，直接返回不重试
		return nil, err
	}

	return nil, fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
}

// FetchItemDetail 获取商品详情
// 根据 xianyu-api.js 中的详情 API 实现
// API: mtop.taobao.idle.pc.detail
func (c *Client) FetchItemDetail(itemID string) (*ItemDetail, error) {
	if itemID == "" {
		return nil, fmt.Errorf("itemID 不能为空")
	}

	reqData := ItemDetailRequest{
		ItemID: itemID,
	}

	resp, err := c.Do(Request{
		API:    "mtop.taobao.idle.pc.detail",
		Data:   reqData,
		Method: "POST",
	})
	if err != nil {
		return nil, fmt.Errorf("请求详情API失败: %w", err)
	}

	// 调试：打印响应信息
	fmt.Printf("[调试] 详情API响应 - ret: %v, data长度: %d\n", resp.Ret, len(resp.Data))

	// 检查返回状态
	success := false
	for _, r := range resp.Ret {
		if r == "SUCCESS::调用成功" || r == "SUCCESS" {
			success = true
			break
		}
	}
	if !success {
		// 打印原始响应用于调试
		// fmt.Printf("[调试] 详情API原始响应: %s\n", string(resp.Data))
		return nil, fmt.Errorf("详情API返回错误: ret=%v", resp.Ret)
	}

	// 解析响应数据 - 匹配实际 API 返回结构
	var detailData struct {
		ItemDO struct {
			// 基础信息
			ItemID     int64  `json:"itemId"`
			Title      string `json:"title"`
			Desc       string `json:"desc"`
			CategoryID int    `json:"categoryId"`

			// 价格信息
			SoldPrice string `json:"soldPrice"` // 售价字符串
			PriceUnit string `json:"priceUnit"`

			// 商品状态
			ItemStatus    int    `json:"itemStatus"`
			ItemStatusStr string `json:"itemStatusStr"`

			// 热度指标
			WantCnt    int `json:"wantCnt"`
			BrowseCnt  int `json:"browseCnt"`
			CollectCnt int `json:"collectCnt"`

			// 时间信息
			GMTCreate      int64  `json:"gmtCreate"` // 毫秒时间戳
			GMT_CREATEDATE string `json:"GMT_CREATE_DATE_KEY"`

			// 库存
			Quantity int `json:"quantity"` // 总库存

			// 图片列表
			ImageInfos []struct {
				URL        string `json:"url"`
				Major      bool   `json:"major"`
				WidthSize  int    `json:"widthSize"`
				HeightSize int    `json:"heightSize"`
			} `json:"imageInfos"`

			// SKU 列表
			SKUList []struct {
				SKUID        int64 `json:"skuId"`
				InventoryID  int64 `json:"inventoryId"`
				Price        int   `json:"price"` // 分为单位
				PriceInCent  int   `json:"priceInCent"`
				Quantity     int   `json:"quantity"`
				PropertyList []struct {
					PropertyID      int64  `json:"propertyId"`
					PropertyText    string `json:"propertyText"`
					ValueID         int64  `json:"valueId"`
					ValueText       string `json:"valueText"`
					ActualValueText string `json:"actualValueText"`
				} `json:"propertyList"`
			} `json:"skuList"`

			// 属性标签（成色等）
			CPVLabels []struct {
				PropertyID   int64  `json:"propertyId"`
				PropertyName string `json:"propertyName"`
				ValueID      int64  `json:"valueId"`
				ValueName    string `json:"valueName"`
			} `json:"cpvLabels"`

			// 商品标签
			ItemLabelExtList []struct {
				ChannelCateID int64  `json:"channelCateId"`
				From          string `json:"from"`
				Text          string `json:"text"`
				Properties    string `json:"properties"`
			} `json:"itemLabelExtList"`

			// 通用标签（如"包邮"）
			CommonTags []struct {
				Text string `json:"text"`
			} `json:"commonTags"`

			// 运费
			TransportFee string `json:"transportFee"`
		} `json:"itemDO"`

		SellerDO struct {
			SellerID          int64  `json:"sellerId"`
			Nick              string `json:"nick"`
			UniqueName        string `json:"uniqueName"`
			City              string `json:"city"`
			PortraitUrl       string `json:"portraitUrl"`
			Signature         string `json:"signature"`
			ItemCount         int    `json:"itemCount"`
			HasSoldNumInteger int    `json:"hasSoldNumInteger"`
			UserRegDay        int    `json:"userRegDay"`
			ZhumaAuth         bool   `json:"zhimaAuth"`
			ZhumaLevelInfo    struct {
				LevelCode string `json:"levelCode"`
				LevelName string `json:"levelName"`
			} `json:"zhumaLevelInfo"`
			IdleFishCreditTag struct {
				TrackParams struct {
					SellerLevel string `json:"sellerLevel"`
				} `json:"trackParams"`
			} `json:"idleFishCreditTag"`
		} `json:"sellerDO"`
	}

	if err := json.Unmarshal(resp.Data, &detailData); err != nil {
		return nil, fmt.Errorf("解析详情数据失败: %w", err)
	}

	// 转换为 ItemDetail - 保持向后兼容的字段映射
	item := &ItemDetail{
		// 基础信息
		ItemID:     fmt.Sprintf("%d", detailData.ItemDO.ItemID),
		Title:      detailData.ItemDO.Title,
		Desc:       detailData.ItemDO.Desc,
		CategoryID: detailData.ItemDO.CategoryID,

		// 价格信息（向后兼容：新字段映射到旧字段）
		SoldPrice: detailData.ItemDO.SoldPrice,
		Price:     detailData.ItemDO.SoldPrice, // 兼容旧字段

		// 热度指标（向后兼容：新字段映射到旧字段）
		WantCount:    detailData.ItemDO.WantCnt,
		ViewCount:    detailData.ItemDO.BrowseCnt,
		CollectCount: detailData.ItemDO.CollectCnt,

		// 时间信息（向后兼容）
		PublishTimeTS: detailData.ItemDO.GMTCreate,
		PublishTime:   time.Unix(detailData.ItemDO.GMTCreate/1000, 0).Format("2006-01-02 15:04:05"),

		// 库存
		TotalStock: detailData.ItemDO.Quantity,

		// 状态
		Status:        detailData.ItemDO.ItemStatusStr,
		ItemStatus:    detailData.ItemDO.ItemStatus,
		ItemStatusStr: detailData.ItemDO.ItemStatusStr,

		// 描述
		Description: detailData.ItemDO.Desc,

		// 新增字段
		PriceInCent: 0, // 将从 SKU 中获取
	}

	// 解析卖家信息（从 sellerDO）
	item.SellerID = fmt.Sprintf("%d", detailData.SellerDO.SellerID)
	item.SellerNick = detailData.SellerDO.Nick
	item.AvatarURL = detailData.SellerDO.PortraitUrl
	item.Location = detailData.SellerDO.City
	item.SellerCity = detailData.SellerDO.City
	item.SellerItemCount = detailData.SellerDO.ItemCount
	item.SellerSoldCount = detailData.SellerDO.HasSoldNumInteger
	item.SellerSignature = detailData.SellerDO.Signature
	// 卖家芝麻信用（取 levelName，如 "信用极好"）
	item.SellerCredit = detailData.SellerDO.ZhumaLevelInfo.LevelName
	// 卖家注册天数（需大于0）
	if detailData.SellerDO.UserRegDay > 0 {
		item.SellerRegDays = detailData.SellerDO.UserRegDay
	}
	// 店铺级别（从 idleFishCreditTag.trackParams.sellerLevel 获取）
	if detailData.SellerDO.IdleFishCreditTag.TrackParams.SellerLevel != "" {
		item.ShopLevel = detailData.SellerDO.IdleFishCreditTag.TrackParams.SellerLevel
	}

	// 处理图片列表
	for _, img := range detailData.ItemDO.ImageInfos {
		item.ImageList = append(item.ImageList, img.URL)
		if img.Major {
			item.ImageURL = img.URL
		}
	}

	// 解析 SKU 信息（完整解析）
	if len(detailData.ItemDO.SKUList) > 0 {
		item.HasSKU = true
		for _, apiSKU := range detailData.ItemDO.SKUList {
			sku := SKU{
				SKUID:       apiSKU.SKUID,
				InventoryID: apiSKU.InventoryID,
				Price:       apiSKU.PriceInCent,
				PriceInCent: apiSKU.PriceInCent,
				Quantity:    apiSKU.Quantity,
			}
			for _, prop := range apiSKU.PropertyList {
				sku.PropertyList = append(sku.PropertyList, SKUProperty{
					PropertyID:      prop.PropertyID,
					PropertyText:    prop.PropertyText,
					ValueID:         prop.ValueID,
					ValueText:       prop.ValueText,
					ActualValueText: prop.ActualValueText,
				})
			}
			item.SKUList = append(item.SKUList, sku)

			// 使用第一个 SKU 的价格
			if item.PriceInCent == 0 {
				item.PriceInCent = apiSKU.PriceInCent
			}
		}
	}

	// 解析属性标签（成色等）
	for _, label := range detailData.ItemDO.CPVLabels {
		item.CPVLabels = append(item.CPVLabels, CPVLabel{
			PropertyID:   label.PropertyID,
			PropertyName: label.PropertyName,
			ValueID:      label.ValueID,
			ValueName:    label.ValueName,
		})

		// 提取 Condition（向后兼容）
		if label.PropertyName == "成色" {
			item.Condition = label.ValueName
			if label.ValueName == "全新" {
				item.IsNew = true
			}
		}
	}

	// 解析商品标签
	for _, tag := range detailData.ItemDO.ItemLabelExtList {
		item.ItemTags = append(item.ItemTags, ItemTag{
			ChannelCateID: tag.ChannelCateID,
			From:          tag.From,
			Text:          tag.Text,
			Properties:    tag.Properties,
		})
	}

	// 解析通用标签（如"包邮"）
	for _, tag := range detailData.ItemDO.CommonTags {
		item.Tags = append(item.Tags, tag.Text)
		if tag.Text == "包邮" {
			item.FreeShipping = true
		}
	}

	return item, nil
}

// PrintItemDetail 打印商品详情
func PrintItemDetail(detail *ItemDetail) {
	fmt.Printf("\n========== 商品详情 ==========\n")
	fmt.Printf("商品ID: %s\n", detail.ItemID)
	fmt.Printf("标题: %s\n", detail.Title)
	if detail.SubTitle != "" {
		fmt.Printf("副标题: %s\n", detail.SubTitle)
	}
	if detail.Desc != "" {
		fmt.Printf("简述: %s\n", detail.Desc)
	}

	// 价格信息
	fmt.Printf("\n【价格】\n")
	fmt.Printf("  售价: %s\n", detail.Price)

	// 卖家信息
	fmt.Printf("\n【卖家】\n")
	fmt.Printf("  昵称: %s\n", detail.SellerNick)
	fmt.Printf("  ID: %s\n", detail.SellerID)
	if detail.SellerCredit != "" {
		fmt.Printf("  芝麻信用: %s\n", detail.SellerCredit)
	}
	if detail.ShopLevel != "" {
		fmt.Printf("  店铺级别: %s\n", detail.ShopLevel)
	}
	if detail.SellerSoldCount > 0 {
		fmt.Printf("  已售: %d 件\n", detail.SellerSoldCount)
	}
	if detail.SellerItemCount > 0 {
		fmt.Printf("  在售: %d 件\n", detail.SellerItemCount)
	}
	if detail.SellerRegDays > 0 {
		fmt.Printf("  注册天数: %d 天\n", detail.SellerRegDays)
	}
	if detail.SellerSignature != "" {
		fmt.Printf("  签名: %s\n", detail.SellerSignature)
	}

	// 商品状态
	fmt.Printf("\n【状态】\n")
	fmt.Printf("  商品状态: %s\n", detail.Status)
	if detail.WantCount > 0 {
		fmt.Printf("  想要人数: %d\n", detail.WantCount)
	}
	if detail.ViewCount > 0 {
		fmt.Printf("  浏览次数: %d\n", detail.ViewCount)
	}
	if detail.CollectCount > 0 {
		fmt.Printf("  收藏次数: %d\n", detail.CollectCount)
	}

	// 库存信息
	if detail.TotalStock > 0 {
		fmt.Printf("\n【库存】\n")
		fmt.Printf("  总库存: %d\n", detail.TotalStock)
	}

	// SKU 信息
	if detail.HasSKU && len(detail.SKUList) > 0 {
		fmt.Printf("\n【规格】(共%d种)\n", len(detail.SKUList))
		for i, sku := range detail.SKUList {
			fmt.Printf("  %d. ¥%.2f (库存:%d)", i+1, float64(sku.PriceInCent)/100, sku.Quantity)
			// 打印属性
			for _, prop := range sku.PropertyList {
				fmt.Printf(" %s:%s", prop.PropertyText, prop.ValueText)
			}
			fmt.Println()
		}
	}

	// 地址信息
	if detail.Location != "" {
		fmt.Printf("\n【地址】\n")
		fmt.Printf("  位置: %s\n", detail.Location)
	}

	// 商品属性
	fmt.Printf("\n【属性】\n")
	if detail.Condition != "" {
		fmt.Printf("  成色: %s", detail.Condition)
		if detail.IsNew {
			fmt.Printf(" (全新)")
		}
		fmt.Println()
	}
	if detail.FreeShipping {
		fmt.Printf("  包邮: 是\n")
	}
	if len(detail.Tags) > 0 {
		fmt.Printf("  标签: %s\n", strings.Join(detail.Tags, ", "))
	}

	// 时间信息
	if detail.PublishTime != "" {
		fmt.Printf("\n【时间】\n")
		fmt.Printf("  发布时间: %s\n", detail.PublishTime)
	}

	// 图片列表
	if len(detail.ImageList) > 0 {
		fmt.Printf("\n【图片】(%d张)\n", len(detail.ImageList))
		for i, img := range detail.ImageList {
			fmt.Printf("  %d. %s\n", i+1, img)
		}
	}

	// 详细描述
	if detail.Description != "" {
		fmt.Printf("\n【描述】\n")
		fmt.Printf("%s\n", detail.Description)
	}

	fmt.Printf("\n=============================\n")
}

// AnalyzeItemDetailForDataAnalysis 输出商品详情的数据分析字段报告
// 分析每个字段的数据分析价值和可能的分析维度
func AnalyzeItemDetailForDataAnalysis(detail *ItemDetail) {
	fmt.Printf("\n")
	fmt.Printf("╔════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║           闲鱼商品详情 - 数据分析字段报告                       ║\n")
	fmt.Printf("╚════════════════════════════════════════════════════════════════╝\n")

	fmt.Printf("\n📊 当前商品样本: %s\n", detail.ItemID)
	fmt.Printf("   标题: %s\n", truncateText(detail.Title, 50))

	// ==================== 高价值字段 ====================
	fmt.Printf("\n")
	fmt.Printf("┌─ 【高价值字段】可直接用于数值分析和可视化 ─────────────────────┐\n")
	fmt.Printf("│                                                                      │\n")

	printFieldAnalysis("Price", detail.Price, "当前价格", "价格分布、区间分析、趋势预测")
	printFieldAnalysis("CategoryID", fmt.Sprintf("%d", detail.CategoryID), "分类ID", "分类统计、各品类价格对比")
	printFieldAnalysis("PublishTimeTS", formatTimestamp(detail.PublishTimeTS), "发布时间戳", "时间序列分析、周期性规律")
	printFieldAnalysis("WantCount", fmt.Sprintf("%d", detail.WantCount), "想要人数", "需求热度、受欢迎程度、成交预测")
	printFieldAnalysis("ViewCount", fmt.Sprintf("%d", detail.ViewCount), "浏览次数", "曝光度、计算转化率")
	printFieldAnalysis("SellerID", detail.SellerID, "卖家ID", "卖家活跃度、商品数量分布")
	printFieldAnalysis("Location", detail.Location, "地理位置", "地域分布、价格地域差异")
	printFieldAnalysis("FreeShipping", fmt.Sprintf("%t", detail.FreeShipping), "是否包邮", "包邮对价格和转化率的影响")
	printFieldAnalysis("SellerRegDays", fmt.Sprintf("%d", detail.SellerRegDays), "卖家注册天数", "卖家资历分析")

	fmt.Printf("│                                                                      │\n")
	fmt.Printf("└──────────────────────────────────────────────────────────────────────┘\n")

	// ==================== 中价值字段 ====================
	fmt.Printf("\n")
	fmt.Printf("┌─ 【中价值字段】需要预处理或作为辅助维度 ─────────────────────────┐\n")
	fmt.Printf("│                                                                      │\n")

	printFieldAnalysis("Title", truncateText(detail.Title, 30), "商品标题", "NLP分析: 品牌提取、关键词、分类")
	printFieldAnalysis("SellerNick", detail.SellerNick, "卖家昵称", "卖家去重标识")
	printFieldAnalysis("Status", detail.Status, "商品状态", "筛选有效数据(online/sold/offline)")
	printFieldAnalysis("CollectCount", fmt.Sprintf("%d", detail.CollectCount), "收藏次数", "收藏热度辅助指标")
	printFieldAnalysis("Condition", detail.Condition, "成色", "需标准化(99新/95新)后分析")
	printFieldAnalysis("IsNew", fmt.Sprintf("%t", detail.IsNew), "是否全新", "新旧商品类别对比")
	printFieldAnalysis("Tags", fmt.Sprintf("%v", detail.Tags), "标签", "特征提取、聚类分析")
	printFieldAnalysis("ImageList", fmt.Sprintf("%d张", len(detail.ImageList)), "图片列表", "图片数与浏览量相关性")
	printFieldAnalysis("Description", truncateText(detail.Description, 30), "详细描述", "NLP关键词提取、情感分析")
	printFieldAnalysis("ShopLevel", detail.ShopLevel, "店铺级别", "卖家信誉分析维度")
	printFieldAnalysis("SellerCredit", detail.SellerCredit, "芝麻信用", "卖家芝麻信用分析维度")

	fmt.Printf("│                                                                      │\n")
	fmt.Printf("└──────────────────────────────────────────────────────────────────────┘\n")

	// ==================== 低价值字段 ====================
	fmt.Printf("\n")
	fmt.Printf("┌─ 【低价值字段】一般不用于直接数据分析 ────────────────────────────┐\n")
	fmt.Printf("│                                                                      │\n")

	printFieldAnalysis("ItemID", detail.ItemID, "商品ID", "仅用于数据关联和去重")
	printFieldAnalysis("SubTitle", detail.SubTitle, "副标题", "补充信息，非核心")
	printFieldAnalysis("Desc", detail.Desc, "简述", "简短描述，非核心")
	printFieldAnalysis("ImageURL", detail.ImageURL, "主图URL", "URL不用于分析")
	printFieldAnalysis("VideoURL", detail.VideoURL, "视频URL", "URL不用于分析")
	printFieldAnalysis("AvatarURL", detail.AvatarURL, "头像URL", "URL不用于分析")
	printFieldAnalysis("PublishTime", detail.PublishTime, "发布时间(字符串)", "建议使用时间戳字段")

	fmt.Printf("│                                                                      │\n")
	fmt.Printf("└──────────────────────────────────────────────────────────────────────┘\n")

	// ==================== 推荐的分析维度 ====================
	fmt.Printf("\n")
	fmt.Printf("┌─ 【推荐的分析维度】基于当前字段可进行的分析 ──────────────────────┐\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 1. 价格分析                                                          │\n")
	fmt.Printf("│    - 价格分布直方图、价格区间统计                                    │\n")
	fmt.Printf("│    - 包邮 vs 不包邮的价格差异                                        │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 2. 热度分析                                                          │\n")
	fmt.Printf("│    - 转化率 = WantCount / ViewCount                                  │\n")
	fmt.Printf("│    - 想要人数分布、浏览次数分布                                      │\n")
	fmt.Printf("│    - 收藏率、咨询率分析                                              │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 3. 地域分析                                                          │\n")
	fmt.Printf("│    - 城市商品数量分布                                                │\n")
	fmt.Printf("│    - 城市级别价格差异                                                │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 4. 时间分析                                                          │\n")
	fmt.Printf("│    - 发布时间分布（小时/星期几/月份）                                │\n")
	fmt.Printf("│    - 商品上架时长分析                                                │\n")
	fmt.Printf("│    - 上新频率趋势                                                    │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 5. 卖家分析                                                          │\n")
	fmt.Printf("│    - 卖家商品数量分布                                                │\n")
	fmt.Printf("│    - 卖家芝麻信用与价格/热度的相关性                                 │\n")
	fmt.Printf("│    - 卖家注册天数与商品表现的关系                                    │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ 6. 分类分析                                                          │\n")
	fmt.Printf("│    - 各品类价格分布                                                  │\n")
	fmt.Printf("│    - 分类热度排行                                                    │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("└──────────────────────────────────────────────────────────────────────┘\n")

	// ==================== 数据处理建议 ====================
	fmt.Printf("\n")
	fmt.Printf("┌─ 【数据处理建议】字段预处理 ──────────────────────────────────────┐\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("│ • Price 字段: 字符串转浮点数，去除¥符号                            │\n")
	fmt.Printf("│ • 时间戳: 转换为datetime格式便于分析                                │\n")
	fmt.Printf("│ • Location: 解析提取省份、城市                                      │\n")
	fmt.Printf("│ • Condition: 标准化为枚举值（全新/99新/95新/等）                   │\n")
	fmt.Printf("│ • Tags: 提取为独热编码(One-Hot)或计数                               │\n")
	fmt.Printf("│ • Title/Description: NLP分词、关键词提取                            │\n")
	fmt.Printf("│ • SellerRegDays: 过滤为0的值，仅保留有效注册天数                   │\n")
	fmt.Printf("│                                                                      │\n")
	fmt.Printf("└──────────────────────────────────────────────────────────────────────┘\n")

	fmt.Printf("\n")
}

// printFieldAnalysis 辅助函数：打印字段分析信息
func printFieldAnalysis(fieldName, value, meaning, usage string) {
	if value == "" || value == "0" {
		value = "(空)"
	}
	// 截断过长的值
	if len(value) > 25 {
		value = value[:22] + "..."
	}
	fmt.Printf("│  • %-12s = %-25s │ 分析: %s\n", fieldName, value, usage)
}

// truncateText 截断文本
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// formatTimestamp 格式化时间戳
func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "(空)"
	}
	t := time.Unix(ts/1000, 0)
	return t.Format("2006-01-02 15:04:05")
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
