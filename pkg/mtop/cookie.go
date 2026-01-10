package mtop

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetTokenFromCookies 从 Cookie 列表中获取 MTOP token
// Cookie 名: _m_h5_tk
// 格式: token_timestamp (需要取下划线前的部分)
func GetTokenFromCookies(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "_m_h5_tk" {
			return parseToken(cookie.Value)
		}
	}
	return ""
}

// GetTokenFromCookieString 从 Cookie 字符串中获取 MTOP token
// 输入格式: "key1=value1; key2=value2; ..."
func GetTokenFromCookieString(cookieStr string) string {
	// 使用正则表达式匹配 _m_h5_tk 的值
	re := regexp.MustCompile(`_m_h5_tk=([^;\s]+)`)
	matches := re.FindStringSubmatch(cookieStr)
	if len(matches) >= 2 {
		return parseToken(matches[1])
	}
	return ""
}

// GetTokenFromJar 从 http.CookieJar 中获取 MTOP token
func GetTokenFromJar(jar *cookiejar.Jar, urlStr string) string {
	if jar == nil {
		return ""
	}

	// 解析 URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	cookies := jar.Cookies(parsedURL)
	return GetTokenFromCookies(cookies)
}

// parseToken 解析 token 值
// Token 格式通常是 "xxx_timestamp"，取下划线前的部分
func parseToken(fullToken string) string {
	parts := strings.Split(fullToken, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return fullToken
}

// GetAllTokens 从 Cookie 中获取所有相关的 token 信息
func GetAllTokens(cookies []*http.Cookie) map[string]string {
	result := make(map[string]string)

	for _, cookie := range cookies {
		switch cookie.Name {
		case "_m_h5_tk":
			result["token"] = parseToken(cookie.Value)
			result["fullToken"] = cookie.Value
		case "_m_h5_tk_enc":
			result["tokenEnc"] = cookie.Value
		case "cookie2":
			result["cookie2"] = cookie.Value
		case "sgcookie":
			result["sgcookie"] = cookie.Value
		case "unb":
			result["unb"] = cookie.Value
		case "umt":
			result["umt"] = cookie.Value
		case "cna":
			result["cna"] = cookie.Value
		case "isg":
			result["isg"] = cookie.Value
		}
	}

	return result
}

// GetAllTokensFromString 从 Cookie 字符串中获取所有相关的 token 信息
func GetAllTokensFromString(cookieStr string) map[string]string {
	result := make(map[string]string)

	// 定义需要提取的 Cookie 名称
	patterns := map[string]string{
		"token":     `_m_h5_tk=([^;\s]+)`,
		"tokenEnc":  `_m_h5_tk_enc=([^;\s]+)`,
		"cookie2":   `cookie2=([^;\s]+)`,
		"sgcookie":  `sgcookie=([^;\s]+)`,
		"unb":       `unb=([^;\s]+)`,
		"umt":       `umt=([^;\s]+)`,
		"cna":       `cna=([^;\s]+)`,
		"isg":       `isg=([^;\s]+)`,
		"sessionId": `sessionId=([^;\s]+)`,
	}

	for key, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(cookieStr)
		if len(matches) >= 2 {
			value := matches[1]
			if key == "token" {
				result["fullToken"] = value
				result[key] = parseToken(value)
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// CookieCookie 简单的 Cookie 结构，用于兼容各种类型
type SimpleCookie struct {
	Name   string
	Value  string
	Domain string
	Path   string

	// 可选字段
	Expires *time.Time
	Secure   bool
	HTTPOnly bool
	SameSite string
}

// GetTokenFromSimpleCookies 从 SimpleCookie 列表中获取 MTOP token
func GetTokenFromSimpleCookies(cookies []SimpleCookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "_m_h5_tk" {
			return parseToken(cookie.Value)
		}
	}
	return ""
}

// GetAllTokensFromSimpleCookies 从 SimpleCookie 列表中获取所有 token
func GetAllTokensFromSimpleCookies(cookies []SimpleCookie) map[string]string {
	result := make(map[string]string)

	for _, cookie := range cookies {
		switch cookie.Name {
		case "_m_h5_tk":
			result["token"] = parseToken(cookie.Value)
			result["fullToken"] = cookie.Value
		case "_m_h5_tk_enc":
			result["tokenEnc"] = cookie.Value
		case "cookie2":
			result["cookie2"] = cookie.Value
		case "sgcookie":
			result["sgcookie"] = cookie.Value
		case "unb":
			result["unb"] = cookie.Value
		case "umt":
			result["umt"] = cookie.Value
		case "cna":
			result["cna"] = cookie.Value
		case "isg":
			result["isg"] = cookie.Value
		}
	}

	return result
}

// ConvertMapToSimpleCookies 将 map[string]string 转换为 SimpleCookie 列表
// 适用于 Playwright 的 Cookies 格式
func ConvertMapToSimpleCookies(cookieMaps []map[string]string) []SimpleCookie {
	result := make([]SimpleCookie, 0, len(cookieMaps))

	for _, cm := range cookieMaps {
		cookie := SimpleCookie{
			Name:  cm["name"],
			Value: cm["value"],
		}

		if domain, ok := cm["domain"]; ok {
			cookie.Domain = domain
		}
		if path, ok := cm["path"]; ok {
			cookie.Path = path
		}
		if expires, ok := cm["expires"]; ok && expires != "" {
			if exp, err := strconv.ParseInt(expires, 10, 64); err == nil {
				t := time.Unix(exp, 0)
				cookie.Expires = &t
			}
		}
		if secure, ok := cm["secure"]; ok && secure == "true" {
			cookie.Secure = true
		}
		if httpOnly, ok := cm["httpOnly"]; ok && httpOnly == "true" {
			cookie.HTTPOnly = true
		}
		if sameSite, ok := cm["sameSite"]; ok {
			cookie.SameSite = sameSite
		}

		result = append(result, cookie)
	}

	return result
}

// ConvertMapSliceToHTTPCookies 将 map 列表转换为 http.Cookie 列表
func ConvertMapSliceToHTTPCookies(cookieMaps []map[string]string) []*http.Cookie {
	result := make([]*http.Cookie, 0, len(cookieMaps))

	for _, cm := range cookieMaps {
		cookie := &http.Cookie{
			Name:  cm["name"],
			Value: cm["value"],
		}

		if domain, ok := cm["domain"]; ok {
			cookie.Domain = domain
		}
		if path, ok := cm["path"]; ok {
			cookie.Path = path
		}
		if expires, ok := cm["expires"]; ok && expires != "" {
			if exp, err := strconv.ParseFloat(expires, 64); err == nil {
				cookie.Expires = time.Unix(int64(exp), 0)
			}
		}
		if secure, ok := cm["secure"]; ok && secure == "true" {
			cookie.Secure = true
		}
		if httpOnly, ok := cm["httpOnly"]; ok && httpOnly == "true" {
			cookie.HttpOnly = true
		}
		if sameSite, ok := cm["sameSite"]; ok {
			switch sameSite {
			case "None":
				cookie.SameSite = http.SameSiteNoneMode
			case "Lax":
				cookie.SameSite = http.SameSiteLaxMode
			case "Strict":
				cookie.SameSite = http.SameSiteStrictMode
			}
		}

		result = append(result, cookie)
	}

	return result
}
