package app

import (
	"fmt"

	"xianyu_aner/internal/config"
	"xianyu_aner/pkg/mtop"
)

// CookieManager Cookie 管理器
type CookieManager struct {
	config config.BrowserConfig
}

// NewCookieManager 创建 Cookie 管理器
func NewCookieManager(cfg config.BrowserConfig) *CookieManager {
	return &CookieManager{config: cfg}
}

// GetCookies 获取 Cookies 并更新配置
func (c *CookieManager) GetCookies(cfg *config.Config) error {
	// 打印启动信息
	mtop.PrintStartupInfo("")
	fmt.Println("正在初始化...")

	// 获取 Cookie
	cookieResult, err := mtop.GetCookiesWithBrowser(mtop.BrowserConfig{
		Headless: c.config.Headless,
	})
	if err != nil {
		return fmt.Errorf("获取Cookie失败: %w", err)
	}

	mtop.PrintStartupInfo(cookieResult.Token)

	// 将Cookie注入到配置中
	cfg.MTOP.Token = cookieResult.Token
	cfg.MTOP.Cookies = cookieResult.Cookies

	return nil
}
