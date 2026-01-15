package config

import (
	"fmt"
	"os"
	"path/filepath"

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
		AntiBot: AntiBotConfig{
			Enabled: true,
			Delay: DelayConfig{
				MinMs: 1000,
				MaxMs: 3000,
			},
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
	loader := newEnvLoader()

	// Server配置
	loader.setInt("SERVER_PORT", &cfg.Server.Port)
	loader.setString("SERVER_MODE", &cfg.Server.Mode)
	loader.setInt("SERVER_TIMEOUT", &cfg.Server.Timeout)

	// Browser配置
	loader.setBool("BROWSER_HEADLESS", &cfg.Browser.Headless)
	loader.setInt("BROWSER_TIMEOUT", &cfg.Browser.Timeout)

	// Feishu配置
	loader.setBool("FEISHU_ENABLED", &cfg.Feishu.Enabled)
	loader.setString("FEISHU_APP_ID", &cfg.Feishu.AppID)
	loader.setString("FEISHU_APP_SECRET", &cfg.Feishu.AppSecret)
	loader.setString("FEISHU_APP_TOKEN", &cfg.Feishu.AppToken)
	loader.setString("FEISHU_TABLE_TOKEN", &cfg.Feishu.TableToken)

	// Logging配置
	loader.setString("LOGGING_LEVEL", &cfg.Logging.Level)
	loader.setString("LOGGING_FORMAT", &cfg.Logging.Format)

	// AntiBot配置
	loader.setBool("ANTI_BOT_ENABLED", &cfg.AntiBot.Enabled)
	loader.setInt("ANTI_BOT_DELAY_MIN_MS", &cfg.AntiBot.Delay.MinMs)
	loader.setInt("ANTI_BOT_DELAY_MAX_MS", &cfg.AntiBot.Delay.MaxMs)
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
