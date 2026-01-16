package crawlcmd

import "flag"

// Flags 命令行参数
type Flags struct {
	ConfigPath  string
	Pages       int
	MinWant     int
	Days        int
	Output      string
	PushFeishu  bool
	Headless    bool
	ShowVersion bool
}

// ParseFlags 解析命令行参数
func ParseFlags() *Flags {
	var (
		configPath  = flag.String("config", "", "配置文件路径")
		pages       = flag.Int("pages", 10, "爬取页数")
		minWant     = flag.Int("min-want", 1, "最低想要人数")
		days        = flag.Int("days", 14, "发布时间范围（天数）")
		output      = flag.String("output", "feed_result.json", "输出文件路径")
		pushFeishu  = flag.Bool("push-feishu", false, "是否推送到飞书")
		headless    = flag.Bool("headless", true, "是否使用无头浏览器")
		showVersion = flag.Bool("version", false, "显示版本信息")
	)
	flag.Parse()

	return &Flags{
		ConfigPath:  *configPath,
		Pages:       *pages,
		MinWant:     *minWant,
		Days:        *days,
		Output:      *output,
		PushFeishu:  *pushFeishu,
		Headless:    *headless,
		ShowVersion: *showVersion,
	}
}
