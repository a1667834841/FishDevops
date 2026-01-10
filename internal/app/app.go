package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xianyu_aner/internal/config"
	"xianyu_aner/internal/server"
	"xianyu_aner/pkg/mtop"
)

// Run 启动应用，包含所有业务逻辑
func Run(cfg config.Config) error {
	// 打印启动信息
	mtop.PrintStartupInfo("")
	fmt.Println("正在初始化...")

	// 获取 Cookie
	cookieResult, err := mtop.GetCookiesWithBrowser(mtop.BrowserConfig{
		Headless: cfg.Browser.Headless,
	})
	if err != nil {
		return fmt.Errorf("获取Cookie失败: %w", err)
	}

	mtop.PrintStartupInfo(cookieResult.Token)

	// 将Cookie注入到配置中
	cfg.MTOP.Token = cookieResult.Token
	cfg.MTOP.Cookies = cookieResult.Cookies

	// 创建服务器
	srv := server.New(cfg)

	// 启动服务器（在 goroutine 中）
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n\n🛑 正在关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("❌ 服务器关闭失败: %v", err)
		return err
	}

	fmt.Println("✅ 服务器已关闭")
	return nil
}
