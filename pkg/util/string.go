package util

import "strings"

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// MaskToken 隐藏 Token 的中间部分
func MaskToken(token string) string {
	if len(token) <= 10 {
		return token
	}
	return token[:8] + "..." + token[len(token)-4:]
}

// BoolToYesNo 布尔值转"是/否"
func BoolToYesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// StringsJoin 字符串数组连接
func StringsJoin(strs []string, sep string) string {
	return strings.Join(strs, sep)
}
