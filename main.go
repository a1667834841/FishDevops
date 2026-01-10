package main

import (
	"log"

	"xianyu_aner/internal/app"
	"xianyu_aner/internal/config"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ 配置验证失败: %v", err)
	}

	// 运行应用
	if err := app.Run(cfg); err != nil {
		log.Fatalf("❌ 应用运行失败: %v", err)
	}
}
