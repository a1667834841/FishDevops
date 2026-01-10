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
		if err := pushToFeishu(cfg, items); err != nil {
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
func pushToFeishu(cfg config.Config, items []mtop.FeedItem) error {
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

	// 转换数据
	products := convertToProducts(items)
	fmt.Printf("成功转换 %d 条数据\n", len(products))

	// 推送到今天的表格
	resp, err := bitableService.PushProductsToTodayTable(products)
	if err != nil {
		return err
	}

	fmt.Printf("推送成功！创建记录数: %d\n", resp.Data.RecordsCreated)
	return nil
}

// convertToProducts 将 FeedItem 转换为 Product
func convertToProducts(items []mtop.FeedItem) []feishu.Product {
	products := make([]feishu.Product, 0, len(items))
	now := time.Now()

	for _, item := range items {
		// 解析价格数值
		priceNumber := parsePrice(item.Price)

		// 构建标签字符串
		tagsStr := strings.Join(item.Tags, ", ")

		// 判断是否包邮
		freeShip := "否"
		for _, tag := range item.Tags {
			if tag == "包邮" {
				freeShip = "是"
				break
			}
		}

		// 构建商品详情URL
		detailURL := fmt.Sprintf("https://2.taobao.com/item.htm?id=%s", item.ItemID)

		product := feishu.Product{
			ItemID:              item.ItemID,
			Title:               item.Title,
			Price:               item.Price,
			PriceNumber:         priceNumber,
			OriginalPrice:       item.PriceOriginal,
			OriginalPriceNumber: parsePrice(item.PriceOriginal),
			WantCnt:             item.WantCount,
			PublishTime:         item.PublishTime,
			PublishTimeMs:       item.PublishTimeTS,
			CaptureTime:         now.Format("2006-01-02 15:04:05"),
			CaptureTimeMs:       now.UnixMilli(),
			SellerCity:          item.Location,
			FreeShip:            freeShip,
			Tags:                tagsStr,
			CoverURL:            item.ImageURL,
			DetailURL:           detailURL,
			ProPolishTime:       item.ProPolishTime,
			ProPolishTimeMs:     item.ProPolishTimeTS,
		}

		products = append(products, product)
	}

	return products
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
