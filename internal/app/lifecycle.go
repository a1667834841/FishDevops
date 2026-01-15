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

	"xianyu_aner/internal/server"
)

// RunWithGracefulShutdown 运行服务器并处理优雅关闭
func RunWithGracefulShutdown(srv *server.Server) error {
	// 启动服务器
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return shutdownServer(srv)
}

// shutdownServer 关闭服务器
func shutdownServer(srv *server.Server) error {
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
