package main

import (
	"log"

	"xianyu_aner/cmd/crawl/commands"
)

func main() {
	flags := commands.ParseFlags()
	cmd := commands.NewCrawlCommand(flags)

	if err := cmd.Run(); err != nil {
		log.Fatalf("执行失败: %v", err)
	}
}
