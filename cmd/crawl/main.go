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

// pushToFeishu 推送数据到飞书
func pushToFeishu(cfg config.Config, client *mtop.Client, items []mtop.FeedItem) error {
	// if !cfg.Feishu.Enabled {
	// 	return fmt.Errorf("飞书功能未启用")
	// }

	// 检查必要配置
	if cfg.Feishu.AppID == "" || cfg.Feishu.AppSecret == "" {
		return fmt.Errorf("缺少飞书配置（app_id 或 app_secret）")
	}

	// 创建飞书客户端
	fsClient := feishu.NewClient(feishu.ClientConfig{
		AppID:     cfg.Feishu.AppID,
		AppSecret: cfg.Feishu.AppSecret,
	})

	bitableConfig := feishu.BitableConfig{
		AppToken:  cfg.Feishu.AppToken,
		TableToken: cfg.Feishu.TableToken,
	}
	bitableService := feishu.NewBitableService(fsClient, bitableConfig)

	// 转换数据（传入 client 参数）
	products := convertToProducts(client, items)
	fmt.Printf("成功转换 %d 条数据\n", len(products))

	// 推送到今天的表格
	resp, err := bitableService.PushProductsToTodayTable(products)
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

	// 解析价格数值
	priceNumber := parsePrice(item.Price)

	return feishu.Product{
		// 基础字段（来自 FeedItem）
		ItemID:        item.ItemID,
		Title:         item.Title,
		Price:         item.Price,
		PriceNumber:   priceNumber,
		WantCnt:       item.WantCount,
		PublishTime:   item.PublishTime,
		PublishTimeMs: item.PublishTimeTS,
		CaptureTime:   captureTime.Format("2006-01-02 15:04:05"),
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
	// 数组字段转 JSON
	imageListJSON, _ := json.Marshal(detail.ImageList)
	skuListJSON, _ := json.Marshal(detail.SKUList)
	cpvLabelsJSON, _ := json.Marshal(detail.CPVLabels)
	itemTagsJSON, _ := json.Marshal(detail.ItemTags)

	// 标签数组转字符串
	tagsStr := strings.Join(detail.Tags, ", ")

	// 包邮判断
	freeShip := "否"
	if detail.FreeShipping {
		freeShip = "是"
	}

	// 构建商品详情URL
	detailURL := fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", detail.ItemID)

	// 解析价格数值
	priceNumber := parsePrice(detail.Price)

	return feishu.Product{
		// 现有字段
		ItemID:        detail.ItemID,
		Title:         detail.Title,
		Price:         detail.Price,
		PriceNumber:   priceNumber,
		WantCnt:       detail.WantCount,
		PublishTime:   detail.PublishTime,
		PublishTimeMs: detail.PublishTimeTS,
		CaptureTime:   captureTime.Format("2006-01-02 15:04:05"),
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
		IsNew:     detail.IsNew,

		// 新增：卖家信息
		SellerID:        detail.SellerID,
		SellerCredit:    detail.SellerCredit,
		ShopLevel:       detail.ShopLevel,
		SellerRegDays:   detail.SellerRegDays,
		SellerItemCount: detail.SellerItemCount,
		SellerSoldCount: detail.SellerSoldCount,
		SellerSignature: detail.SellerSignature,

		// 新增：商品描述
		Description: detail.Description,
		Desc:        detail.Desc,
		SubTitle:    detail.SubTitle,

		// 新增：媒体资源
		VideoURL: detail.VideoURL,

		// 新增：分类信息
		CategoryID: detail.CategoryID,

		// 新增：商品状态
		Status:        detail.Status,
		ItemStatus:    detail.ItemStatus,
		ItemStatusStr: detail.ItemStatusStr,

		// 新增：数组字段（JSON格式）
		ImageListJSON: string(imageListJSON),
		SKUListJSON:   string(skuListJSON),
		CPVLabelsJSON: string(cpvLabelsJSON),
		ItemTagsJSON:  string(itemTagsJSON),

		// 新增：其他
		HasSKU:     detail.HasSKU,
		TotalStock: detail.TotalStock,
		PriceInCent: detail.PriceInCent,
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
