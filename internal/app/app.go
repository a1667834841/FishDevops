package app

import (
	"xianyu_aner/internal/config"
	"xianyu_aner/internal/server"
)

// Run 启动应用（仅协调各组件）
func Run(cfg config.Config) error {
	// 1. 获取 Cookie
	cookieMgr := NewCookieManager(cfg.Browser)
	if err := cookieMgr.GetCookies(&cfg); err != nil {
		return err
	}

	// 2. 创建服务器
	srv := server.New(cfg)

	// 3. 运行服务（带优雅关闭）
	return RunWithGracefulShutdown(srv)
}
