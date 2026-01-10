package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load 加载配置
// 优先级: 环境变量 > 配置文件 > 默认值
func Load() Config {
	cfg := defaultConfig()

	// 尝试从配置文件加载
	if configPath := findConfigFile(); configPath != "" {
		if err := loadFromFile(configPath, &cfg); err == nil {
			fmt.Printf("✅ 加载配置文件: %s\n", configPath)
		}
	}

	// 环境变量覆盖
	loadFromEnv(&cfg)

	return cfg
}

// defaultConfig 返回默认配置
func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:    8080,
			Mode:    "release",
			Timeout: 30,
		},
		Browser: BrowserConfig{
			Headless: true,
			Timeout:  60,
		},
		Feishu: FeishuConfig{
			Enabled: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// findConfigFile 查找配置文件
func findConfigFile() string {
	// 按优先级查找配置文件
	paths := []string{
		"config.yaml",
		"configs/config.yaml",
		"config.local.yaml",
		"configs/config.local.yaml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}

// loadFromFile 从YAML文件加载配置
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	// Server配置
	if v := os.Getenv("SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("SERVER_MODE"); v != "" {
		cfg.Server.Mode = v
	}
	if v := os.Getenv("SERVER_TIMEOUT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Timeout)
	}

	// Browser配置
	if v := os.Getenv("BROWSER_HEADLESS"); v != "" {
		cfg.Browser.Headless = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("BROWSER_TIMEOUT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Browser.Timeout)
	}

	// Feishu配置
	if v := os.Getenv("FEISHU_ENABLED"); v != "" {
		cfg.Feishu.Enabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("FEISHU_APP_ID"); v != "" {
		cfg.Feishu.AppID = v
	}
	if v := os.Getenv("FEISHU_APP_SECRET"); v != "" {
		cfg.Feishu.AppSecret = v
	}
	if v := os.Getenv("FEISHU_APP_TOKEN"); v != "" {
		cfg.Feishu.AppToken = v
	}
	if v := os.Getenv("FEISHU_TABLE_TOKEN"); v != "" {
		cfg.Feishu.TableToken = v
	}

	// Logging配置
	if v := os.Getenv("LOGGING_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOGGING_FORMAT"); v != "" {
		cfg.Logging.Format = v
	}
}

// Validate 验证配置
func (c Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", c.Server.Port)
	}

	if c.Server.Mode != "debug" && c.Server.Mode != "release" && c.Server.Mode != "test" {
		return fmt.Errorf("无效的服务器模式: %s", c.Server.Mode)
	}

	if c.Feishu.Enabled {
		if c.Feishu.AppID == "" || c.Feishu.AppSecret == "" {
			return fmt.Errorf("飞书功能已启用，但缺少必要的配置（app_id 或 app_secret）")
		}
	}

	return nil
}
