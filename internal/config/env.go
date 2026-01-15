package config

import (
	"os"
	"strconv"
	"strings"
)

// envLoader 环境变量加载辅助函数
type envLoader struct{}

// newEnvLoader 创建环境变量加载器
func newEnvLoader() *envLoader {
	return &envLoader{}
}

// getString 获取字符串环境变量
func (e *envLoader) getString(key string, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getInt 获取整数环境变量
func (e *envLoader) getInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// getBool 获取布尔环境变量
func (e *envLoader) getBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		return strings.ToLower(v) == "true" || v == "1"
	}
	return defaultVal
}

// setString 如果环境变量存在，设置字符串值
func (e *envLoader) setString(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

// setInt 如果环境变量存在，设置整数值
func (e *envLoader) setInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			*target = i
		}
	}
}

// setBool 如果环境变量存在，设置布尔值
func (e *envLoader) setBool(key string, target *bool) {
	if v := os.Getenv(key); v != "" {
		*target = strings.ToLower(v) == "true" || v == "1"
	}
}
