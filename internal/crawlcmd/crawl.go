package crawlcmd

import (
	"fmt"
	"log"
	"time"

	"xianyu_aner/internal/config"
	"xianyu_aner/internal/service"
)

// CrawlCommand 爬取命令
type CrawlCommand struct {
	flags *Flags
}

// NewCrawlCommand 创建爬取命令
func NewCrawlCommand(flags *Flags) *CrawlCommand {
	return &CrawlCommand{flags: flags}
}

// Run 执行爬取命令
func (c *CrawlCommand) Run() error {
	// 显示版本
	if c.flags.ShowVersion {
		fmt.Println("xianyu_aner crawl v1.0.0")
		return nil
	}

	// 加载配置
	cfg := config.Load()
	cfg.Browser.Headless = c.flags.Headless

	// 打印启动信息
	printBanner()

	// 创建服务
	fetcher := service.NewFetcher(cfg)
	pusher := service.NewPusher(cfg)

	// 执行爬取流程
	result, err := c.executeCrawl(cfg, fetcher, pusher)
	if err != nil {
		return err
	}

	// 打印统计
	printSummary(result)
	return nil
}

func (c *CrawlCommand) executeCrawl(cfg config.Config, fetcher *service.Fetcher, pusher *service.Pusher) (*service.Result, error) {
	startTime := time.Now()

	// 步骤1: 获取 Cookie
	fmt.Printf("\n[步骤 1/4] 获取登录 Cookie (无头模式: %v)...\n", cfg.Browser.Headless)
	mtopClient, err := fetcher.InitClient()
	if err != nil {
		return nil, fmt.Errorf("初始化客户端失败: %w", err)
	}

	// 步骤2: 爬取数据
	fmt.Printf("\n[步骤 2/4] 爬取猜你喜欢数据 (页数: %d)...\n", c.flags.Pages)
	items, err := fetcher.Fetch(mtopClient, c.flags.Pages, c.flags.MinWant, c.flags.Days)
	if err != nil {
		return nil, fmt.Errorf("爬取失败: %w", err)
	}
	fmt.Printf("爬取完成！获取到 %d 条数据\n", len(items))

	// 步骤3: 保存到文件
	fmt.Printf("\n[步骤 3/4] 保存数据到文件: %s\n", c.flags.Output)
	if err := service.SaveToFile(items, c.flags.Output); err != nil {
		log.Printf("保存文件失败: %v", err)
	}

	// 步骤4: 推送到飞书（可选）
	if c.flags.PushFeishu {
		fmt.Printf("\n[步骤 4/4] 推送到飞书多维表格...\n")
		if err := pusher.Push(mtopClient, items); err != nil {
			log.Printf("推送失败: %v", err)
		}
	} else {
		fmt.Println("\n[步骤 4/4] 跳过飞书推送")
	}

	return &service.Result{
		TotalItems: len(items),
		Duration:   time.Since(startTime),
	}, nil
}

func printBanner() {
	fmt.Println("========================================")
	fmt.Println("  闲鱼数据爬取工具")
	fmt.Println("========================================")
}

func printSummary(result *service.Result) {
	fmt.Println("\n========================================")
	fmt.Println("  任务完成统计")
	fmt.Println("========================================")
	fmt.Printf("爬取商品数: %d\n", result.TotalItems)
	fmt.Printf("总耗时: %.2f 秒\n", result.Duration.Seconds())
	fmt.Println("========================================")
}
