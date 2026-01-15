package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"xianyu_aner/internal/config"
	"xianyu_aner/pkg/feishu"
	"xianyu_aner/pkg/mtop"
)

func main() {
	// 命令行参数
	var (
		configPath  = flag.String("configs", "", "配置文件路径")
		pages       = flag.Int("pages", 10, "爬取页数")
		minWant     = flag.Int("min-want", 1, "最低想要人数")
		days        = flag.Int("days", 14, "发布时间范围（天数）")
		output      = flag.String("output", "feed_result.json", "输出文件路径")
		pushFeishu  = flag.Bool("push-feishu", false, "是否推送到飞书")
		headless    = flag.Bool("headless", true, "是否使用无头浏览器")
		showVersion = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("xianyu_aner crawl v1.0.0")
		return
	}

	// 加载配置
	var cfg config.Config
	if *configPath != "" {
		// 使用指定的配置文件
		cfg = config.Load()
	} else {
		// 自动查找配置文件
		cfg = config.Load()
	}

	// 覆盖浏览器配置
	cfg.Browser.Headless = *headless

	// 打印启动信息
	fmt.Println("========================================")
	fmt.Println("  闲鱼数据爬取工具")
	fmt.Println("========================================")

	// 步骤1: 获取 Cookie
	fmt.Printf("\n[步骤 1/4] 获取登录 Cookie (无头模式: %v)...\n", cfg.Browser.Headless)
	cookieResult, err := mtop.GetCookiesWithBrowser(mtop.BrowserConfig{
		Headless: cfg.Browser.Headless,
	})
	if err != nil {
		log.Fatalf("获取 Cookie 失败: %v", err)
	}
	fmt.Printf("成功获取 Token: %s...\n", maskToken(cookieResult.Token))

	// 步骤2: 爬取数据
	fmt.Printf("\n[步骤 2/4] 爬取猜你喜欢数据 (页数: %d)...\n", *pages)
	client := mtop.NewClient(cookieResult.Token, "34839810",
		mtop.WithCookies(cookieResult.Cookies),
	)

	startTime := time.Now()
	items, err := client.GuessYouLike("", *pages, mtop.GuessYouLikeOptions{
		MaxPages:     *pages,
		StartPage:    1,
		MinWantCount: *minWant,
		DaysWithin:   *days,
	})
	if err != nil {
		log.Fatalf("爬取失败: %v", err)
	}
	duration := time.Since(startTime)

	fmt.Printf("爬取完成！获取到 %d 条数据，耗时 %.2f 秒\n", len(items), duration.Seconds())

	// 步骤3: 保存到文件
	fmt.Printf("\n[步骤 3/4] 保存数据到文件: %s\n", *output)
	data, _ := json.MarshalIndent(items, "", "  ")
	if err := os.WriteFile(*output, data, 0644); err != nil {
		log.Printf("保存文件失败: %v", err)
	} else {
		fmt.Printf("成功保存 %d 条数据\n", len(items))
	}

	// 步骤4: 推送到飞书（可选）
	if *pushFeishu {
		fmt.Printf("\n[步骤 4/4] 推送到飞书多维表格...\n")
		if err := pushToFeishu(cfg, client, items); err != nil {
			log.Printf("推送失败: %v", err)
		}
	} else {
		fmt.Println("\n[步骤 4/4] 跳过飞书推送")
	}

	// 打印统计信息
	fmt.Println("\n========================================")
	fmt.Println("  任务完成统计")
	fmt.Println("========================================")
	fmt.Printf("爬取商品数: %d\n", len(items))
	fmt.Printf("总耗时: %.2f 秒\n", duration.Seconds())
	fmt.Println("========================================")
}

// convertFeedItemsToBasicProducts 将 FeedItem 转换为仅包含基础字段的 Product（用于去重）
// 保留字段：商品ID、价格、想要数、采集时间、发布时间、商品标题、商品链接、封面图
func convertFeedItemsToBasicProducts(items []mtop.FeedItem) []feishu.Product {
	products := make([]feishu.Product, 0, len(items))
	now := time.Now()

	for _, item := range items {
		if item.ItemID == "" {
			continue
		}

		product := feishu.Product{
			// 去重关键字段
			ItemID:        item.ItemID,
			Price:         item.Price,
			WantCnt:       item.WantCount,

			// 时间信息（采集时间使用当前时间）
			PublishTimeMs: item.PublishTimeTS,
			CaptureTimeMs: now.UnixMilli(),

			// 基础显示字段
			Title:     item.Title,
			CoverURL:  item.ImageURL,
			DetailURL: fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", item.ItemID),
		}
		products = append(products, product)
	}

	return products
}

// mergeDetailToProduct 将详情数据合并到基础产品数据中
func mergeDetailToProduct(basic feishu.Product, detail *mtop.ItemDetail) feishu.Product {
	// 保留基础数据
	result := basic

	// 补充详情数据（24个字段的完整信息）
	tagsStr := strings.Join(detail.Tags, ", ")
	freeShip := "否"
	if detail.FreeShipping {
		freeShip = "是"
	}

	// 热度指标
	result.ViewCount = detail.ViewCount
	result.CollectCount = detail.CollectCount
	// ExposureHeat 字段在 ItemDetail 中不存在，保持默认值

	// 商品属性
	result.SubTitle = detail.SubTitle
	// OriginalPrice 字段在 ItemDetail 中不存在，保持默认值
	result.Condition = detail.Condition

	// 卖家信息
	result.SellerNick = detail.SellerNick
	result.SellerCity = detail.SellerCity
	result.SellerCredit = detail.SellerCredit
	result.SellerItemCount = detail.SellerItemCount
	result.SellerSoldCount = detail.SellerSoldCount
	result.FreeShip = freeShip

	// 其他
	result.Tags = tagsStr
	result.VideoURL = detail.VideoURL
	result.ItemStatusStr = detail.ItemStatusStr
	result.Description = detail.Description

	// 链接资源（详情接口可能返回更准确的URL）
	if detail.ImageURL != "" {
		result.CoverURL = detail.ImageURL
	}

	return result
}

// enrichProductsWithDetails 对去重后的商品补充详情数据
func enrichProductsWithDetails(client *mtop.Client, basicProducts []feishu.Product) []feishu.Product {
	finalProducts := make([]feishu.Product, 0, len(basicProducts))
	successCount := 0
	failCount := 0

	fmt.Printf("\n[开始] 获取商品详情，共 %d 个新商品...\n", len(basicProducts))

	for i, basic := range basicProducts {
		fmt.Printf("[处理 %d/%d] 正在获取详情: %s (ID: %s)...\n",
			i+1, len(basicProducts), truncateString(basic.Title, 30), basic.ItemID)

		detail, err := client.FetchItemDetail(basic.ItemID)
		if err != nil {
			fmt.Printf("[失败 %d/%d] 获取详情失败，使用基础数据: %v\n", i+1, len(basicProducts), err)
			failCount++
			// 使用基础数据，不影响整体流程
			finalProducts = append(finalProducts, basic)
		} else {
			// 合并详情数据
			enriched := mergeDetailToProduct(basic, detail)
			finalProducts = append(finalProducts, enriched)
			successCount++

			fmt.Printf("[成功 %d/%d] ✓ %s - ¥%s (想要:%d, 浏览:%d)\n",
				i+1, len(basicProducts), truncateString(detail.Title, 20),
				detail.Price, detail.WantCount, detail.ViewCount)
		}
	}

	fmt.Printf("\n[完成] 详情成功: %d, 失败(使用基础数据): %d\n", successCount, failCount)
	return finalProducts
}

// pushToFeishu 推送数据到飞书（四阶段流程）
func pushToFeishu(cfg config.Config, client *mtop.Client, items []mtop.FeedItem) error {
	// 检查必要配置
	if cfg.Feishu.AppID == "" || cfg.Feishu.AppSecret == "" {
		return fmt.Errorf("缺少飞书配置（app_id 或 app_secret）")
	}

	// 创建飞书客户端（全程使用同一个实例）
	fsClient := feishu.NewClient(feishu.ClientConfig{
		AppID:     cfg.Feishu.AppID,
		AppSecret: cfg.Feishu.AppSecret,
	})

	bitableConfig := feishu.BitableConfig{
		AppToken:   cfg.Feishu.AppToken,
		TableToken: cfg.Feishu.TableToken,
	}
	bitableService := feishu.NewBitableService(fsClient, bitableConfig)

	// === 阶段1：转换为基础产品结构（仅用于去重）===
	fmt.Println("\n[阶段1/4] 转换为基础产品结构（去重用）...")
	basicProducts := convertFeedItemsToBasicProducts(items)
	fmt.Printf("转换完成：%d 条基础记录\n", len(basicProducts))

	// === 阶段2：去重查询 ===
	fmt.Println("\n[阶段2/4] 查询飞书表格进行去重...")
	today := time.Now()
	tableID, created, err := bitableService.GetOrCreateTableByDate(today)
	if err != nil {
		return fmt.Errorf("获取/创建今日表格失败: %w", err)
	}

	var uniqueProducts []feishu.Product
	if !created {
		// 表格已存在，执行去重
		uniqueProducts, err = bitableService.DeduplicateProducts(tableID, basicProducts)
		if err != nil {
			return fmt.Errorf("去重失败: %w", err)
		}
		fmt.Printf("去重完成：%d 条新记录（过滤掉 %d 条重复记录）\n",
			len(uniqueProducts), len(basicProducts)-len(uniqueProducts))
	} else {
		// 新创建的表格，全部保留
		uniqueProducts = basicProducts
		fmt.Println("新表格创建，保留所有记录")
	}

	// 如果没有新数据，提前退出
	if len(uniqueProducts) == 0 {
		fmt.Println("\n[完成] 没有新数据需要推送")
		return nil
	}

	// === 阶段3：获取详情（仅对新记录）===
	fmt.Println("\n[阶段3/4] 获取商品详情...")
	finalProducts := enrichProductsWithDetails(client, uniqueProducts)

	// === 阶段4：推送到飞书 ===
	fmt.Println("\n[阶段4/4] 推送到飞书...")
	resp, err := bitableService.PushProductsToDateTable(today, finalProducts)
	if err != nil {
		return err
	}

	fmt.Printf("推送成功！创建记录数: %d\n", resp.Data.RecordsCreated)
	return nil
}

// convertToProducts 将 FeedItem 转换为 Product（通过商品详情接口补充完整信息）
func convertToProducts(client *mtop.Client, items []mtop.FeedItem) []feishu.Product {
	products := make([]feishu.Product, 0, len(items))
	now := time.Now()
	detailSuccessCount := 0
	detailFailCount := 0
	fallbackCount := 0

	fmt.Printf("\n[开始] 获取商品详情，共 %d 个商品...\n", len(items))

	for i, item := range items {
		// 检查 itemId 是否为空
		if item.ItemID == "" {
			fmt.Printf("[跳过 %d/%d] 商品 itemId 为空: %s\n", i+1, len(items), truncateString(item.Title, 30))
			detailFailCount++
			continue
		}

		fmt.Printf("[处理 %d/%d] 正在获取商品详情: %s (ID: %s)...\n",
			i+1, len(items), truncateString(item.Title, 30), item.ItemID)

		// 串行调用商品详情接口
		detail, err := client.FetchItemDetail(item.ItemID)
		if err != nil {
			fmt.Printf("[降级 %d/%d] 获取商品详情失败，使用猜你喜欢数据: %v\n", i+1, len(items), err)
			detailFailCount++
			fallbackCount++

			// Fallback: 使用 FeedItem 原始数据
			product := convertFeedItemToProduct(item, now)
			products = append(products, product)
		} else {
			// 转换 ItemDetail 到 Product
			product := convertItemDetailToProduct(detail, now)
			products = append(products, product)
			detailSuccessCount++

			fmt.Printf("[成功 %d/%d] ✓ %s - ¥%s (想要:%d, 浏览:%d)\n",
				i+1, len(items), truncateString(detail.Title, 20),
				detail.Price, detail.WantCount, detail.ViewCount)
		}
	}

	fmt.Printf("\n[完成] 详情成功: %d, 降级使用猜你喜欢: %d, 跳过: %d, 总计: %d\n",
		detailSuccessCount, fallbackCount, detailFailCount-fallbackCount, len(items))

	return products
}

// convertFeedItemToProduct 将 FeedItem 转换为 Product（Fallback 机制）
func convertFeedItemToProduct(item mtop.FeedItem, captureTime time.Time) feishu.Product {
	// 构建标签字符串
	tagsStr := strings.Join(item.Tags, ", ")

	// 包邮判断
	freeShip := "否"
	for _, tag := range item.Tags {
		if tag == "包邮" {
			freeShip = "是"
			break
		}
	}

	// 构建商品详情URL
	detailURL := fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", item.ItemID)

	return feishu.Product{
		// 基础字段（来自 FeedItem）
		ItemID:        item.ItemID,
		Title:         item.Title,
		Price:         item.Price,
		WantCnt:       item.WantCount,
		PublishTimeMs: item.PublishTimeTS,
		CaptureTimeMs: captureTime.UnixMilli(),
		SellerCity:    item.Location,
		FreeShip:      freeShip,
		Tags:          tagsStr,
		CoverURL:      item.ImageURL,
		DetailURL:     detailURL,

		// 以下字段在 FeedItem 中不可用，留空
		// 新增字段（FeedItem 中不可用，留空）
	}
}

// convertItemDetailToProduct 将 ItemDetail 转换为 Product
func convertItemDetailToProduct(detail *mtop.ItemDetail, captureTime time.Time) feishu.Product {
	// 标签数组转字符串
	tagsStr := strings.Join(detail.Tags, ", ")

	// 包邮判断
	freeShip := "否"
	if detail.FreeShipping {
		freeShip = "是"
	}

	// 构建商品详情URL
	detailURL := fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", detail.ItemID)

	return feishu.Product{
		// 基本字段
		ItemID:        detail.ItemID,
		Title:         detail.Title,
		Price:         detail.Price,
		WantCnt:       detail.WantCount,
		PublishTimeMs: detail.PublishTimeTS,
		CaptureTimeMs: captureTime.UnixMilli(),
		SellerNick:    detail.SellerNick,
		SellerCity:    detail.SellerCity,
		FreeShip:      freeShip,
		Tags:          tagsStr,
		CoverURL:      detail.ImageURL,
		DetailURL:     detailURL,

		// 新增：热度指标
		ViewCount:    detail.ViewCount,
		CollectCount: detail.CollectCount,

		// 新增：商品属性
		Condition: detail.Condition,

		// 新增：卖家信息
		SellerCredit:    detail.SellerCredit,
		SellerItemCount: detail.SellerItemCount,
		SellerSoldCount: detail.SellerSoldCount,

		// 新增：商品描述
		Description: detail.Description,
		SubTitle:    detail.SubTitle,

		// 新增：媒体资源
		VideoURL: detail.VideoURL,

		// 新增：商品状态
		ItemStatusStr: detail.ItemStatusStr,
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parsePrice 从价格字符串解析数值
func parsePrice(priceStr string) float64 {
	priceStr = strings.TrimSpace(priceStr)
	priceStr = strings.ReplaceAll(priceStr, "¥", "")
	priceStr = strings.ReplaceAll(priceStr, "￥", "")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	priceStr = strings.ReplaceAll(priceStr, " ", "")

	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	return price
}

// maskToken 隐藏 Token 的中间部分
func maskToken(token string) string {
	if len(token) <= 10 {
		return token
	}
	return token[:8] + "..." + token[len(token)-4:]
}
