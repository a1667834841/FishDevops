package config

import (
	"net/http"
	"time"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `yaml:"server" env-prefix:"SERVER_"`
	Browser BrowserConfig `yaml:"browser" env-prefix:"BROWSER_"`
	Feishu  FeishuConfig  `yaml:"feishu" env-prefix:"FEISHU_"`
	Logging LoggingConfig `yaml:"logging" env-prefix:"LOGGING_"`
	AntiBot AntiBotConfig `yaml:"anti_bot" env-prefix:"ANTI_BOT_"` // 反爬虫配置
	MTOP    MTOPConfig    `yaml:"-"`                               // MTOP配置不直接从文件加载
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    int    `yaml:"port" env:"PORT" default:"8080"`
	Mode    string `yaml:"mode" env:"MODE" default:"release"` // debug, release, test
	Timeout int    `yaml:"timeout" env:"TIMEOUT" default:"30"` // 秒
}

// BrowserConfig 浏览器配置
type BrowserConfig struct {
	Headless bool `yaml:"headless" env:"HEADLESS" default:"true"`
	Timeout  int  `yaml:"timeout" env:"TIMEOUT" default:"60"` // 秒
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enabled    bool   `yaml:"enabled" env:"ENABLED" default:"false"`
	AppID      string `yaml:"app_id" env:"APP_ID"`
	AppSecret  string `yaml:"app_secret" env:"APP_SECRET"`
	AppToken   string `yaml:"app_token" env:"APP_TOKEN"`
	TableToken string `yaml:"table_token" env:"TABLE_TOKEN"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LEVEL" default:"info"`  // debug, info, warn, error
	Format string `yaml:"format" env:"FORMAT" default:"text"` // json, text
}

// MTOPConfig MTOP客户端配置（运行时注入）
type MTOPConfig struct {
	Token   string
	Cookies []*http.Cookie
}

// GetTimeout 获取超时时间
func (c ServerConfig) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// AntiBotConfig 反爬虫配置
type AntiBotConfig struct {
	Enabled bool        `yaml:"enabled" env:"ENABLED" default:"true"` // 是否启用反爬虫
	Delay   DelayConfig `yaml:"delay"`                                // 延迟配置
}

// DelayConfig 延迟配置
type DelayConfig struct {
	MinMs int `yaml:"min_ms" env:"DELAY_MIN_MS" default:"1000"` // 最小延迟（毫秒）
	MaxMs int `yaml:"max_ms" env:"DELAY_MAX_MS" default:"3000"` // 最大延迟（毫秒）
}
