package service

import (
	"encoding/json"
	"fmt"
	"os"

	"xianyu_aner/internal/config"
	"xianyu_aner/pkg/mtop"
)

// Fetcher 数据获取服务（通用，可供 server 和 crawl 使用）
type Fetcher struct {
	cfg config.Config
}

// NewFetcher 创建获取服务
func NewFetcher(cfg config.Config) *Fetcher {
	return &Fetcher{cfg: cfg}
}

// InitClient 初始化 MTOP 客户端
func (f *Fetcher) InitClient() (*mtop.Client, error) {
	cookieResult, err := mtop.GetCookiesWithBrowser(mtop.BrowserConfig{
		Headless: f.cfg.Browser.Headless,
	})
	if err != nil {
		return nil, fmt.Errorf("获取Cookie失败: %w", err)
	}

	return mtop.NewClient(cookieResult.Token, "34839810",
		mtop.WithCookies(cookieResult.Cookies),
	), nil
}

// Fetch 获取猜你喜欢数据
func (f *Fetcher) Fetch(mtopClient *mtop.Client, pages, minWant, days int) ([]mtop.FeedItem, error) {
	return mtopClient.GuessYouLike("", pages, mtop.GuessYouLikeOptions{
		MaxPages:     pages,
		StartPage:    1,
		MinWantCount: minWant,
		DaysWithin:   days,
	})
}

// SaveToFile 保存数据到文件
func SaveToFile(items []mtop.FeedItem, filepath string) error {
	data, _ := json.MarshalIndent(items, "", "  ")
	return os.WriteFile(filepath, data, 0644)
}
