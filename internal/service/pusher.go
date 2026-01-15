package service

import (
	"fmt"
	"time"

	"xianyu_aner/internal/config"
	"xianyu_aner/pkg/feishu"
	"xianyu_aner/pkg/mtop"
	"xianyu_aner/pkg/util"
)

// Pusher 飞书推送服务（通用，可供 server 和 crawl 使用）
type Pusher struct {
	cfg       config.Config
	converter *Converter
}

// NewPusher 创建推送服务
func NewPusher(cfg config.Config) *Pusher {
	return &Pusher{
		cfg:       cfg,
		converter: NewConverter(),
	}
}

// Push 推送数据到飞书（四阶段流程）
func (p *Pusher) Push(mtopClient *mtop.Client, items []mtop.FeedItem) error {
	if p.cfg.Feishu.AppID == "" || p.cfg.Feishu.AppSecret == "" {
		return fmt.Errorf("缺少飞书配置（app_id 或 app_secret）")
	}

	// 创建飞书客户端
	fsClient := feishu.NewClient(feishu.ClientConfig{
		AppID:     p.cfg.Feishu.AppID,
		AppSecret: p.cfg.Feishu.AppSecret,
	})

	bitableConfig := feishu.BitableConfig{
		AppToken:   p.cfg.Feishu.AppToken,
		TableToken: p.cfg.Feishu.TableToken,
	}
	bitableService := feishu.NewBitableService(fsClient, bitableConfig)

	// 执行四阶段推送流程
	return p.executeFourStagePush(mtopClient, bitableService, items)
}

func (p *Pusher) executeFourStagePush(mtopClient *mtop.Client, bitableService *feishu.BitableService, items []mtop.FeedItem) error {
	deduplicator := NewDeduplicator()

	// 阶段1：转换为基础产品结构
	fmt.Println("\n[阶段1/4] 转换为基础产品结构（去重用）...")
	basicProducts := p.converter.FeedItemsToBasicProducts(items)
	fmt.Printf("转换完成：%d 条基础记录\n", len(basicProducts))

	// 阶段2：去重查询
	fmt.Println("\n[阶段2/4] 查询飞书表格进行去重...")
	uniqueProducts, err := p.deduplicate(bitableService, deduplicator, basicProducts)
	if err != nil {
		return fmt.Errorf("去重失败: %w", err)
	}

	if len(uniqueProducts) == 0 {
		fmt.Println("\n[完成] 没有新数据需要推送")
		return nil
	}

	// 阶段3：获取详情
	fmt.Println("\n[阶段3/4] 获取商品详情...")
	finalProducts := p.enrichDetails(mtopClient, uniqueProducts)

	// 阶段4：推送到飞书
	fmt.Println("\n[阶段4/4] 推送到飞书...")
	resp, err := bitableService.PushProductsToDateTable(time.Now(), finalProducts)
	if err != nil {
		return err
	}

	fmt.Printf("推送成功！创建记录数: %d\n", resp.Data.RecordsCreated)
	return nil
}

func (p *Pusher) deduplicate(bitableService *feishu.BitableService, deduplicator *Deduplicator, products []feishu.Product) ([]feishu.Product, error) {
	today := time.Now()
	tableID, created, err := bitableService.GetOrCreateTableByDate(today)
	if err != nil {
		return nil, fmt.Errorf("获取/创建今日表格失败: %w", err)
	}

	var uniqueProducts []feishu.Product
	if !created {
		// 表格已存在，执行去重
		uniqueProducts, err = deduplicator.DeduplicateProducts(bitableService, tableID, products)
		if err != nil {
			return nil, err
		}
		fmt.Printf("去重完成：%d 条新记录（过滤掉 %d 条重复记录）\n",
			len(uniqueProducts), len(products)-len(uniqueProducts))
	} else {
		// 新创建的表格，全部保留
		uniqueProducts = products
		fmt.Println("新表格创建，保留所有记录")
	}

	return uniqueProducts, nil
}

func (p *Pusher) enrichDetails(mtopClient *mtop.Client, products []feishu.Product) []feishu.Product {
	finalProducts := make([]feishu.Product, 0, len(products))

	for i, basic := range products {
		fmt.Printf("[处理 %d/%d] 正在获取详情: %s (ID: %s)...\n",
			i+1, len(products), util.TruncateString(basic.Title, 30), basic.ItemID)

		detail, err := mtopClient.FetchItemDetail(basic.ItemID)
		if err != nil {
			fmt.Printf("[失败 %d/%d] 获取详情失败，使用基础数据: %v\n", i+1, len(products), err)
			finalProducts = append(finalProducts, basic)
		} else {
			enriched := p.converter.MergeDetailToProduct(basic, detail)
			finalProducts = append(finalProducts, enriched)
		}
	}

	return finalProducts
}
