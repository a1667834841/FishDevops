package main

import (
	"log"

	"xianyu_aner/internal/crawlcmd"
)

func main() {
	flags := crawlcmd.ParseFlags()
	cmd := crawlcmd.NewCrawlCommand(flags)

	if err := cmd.Run(); err != nil {
		log.Fatalf("执行失败: %v", err)
	}
}
