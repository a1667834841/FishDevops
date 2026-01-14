package mtop

import (
	"flag"
	"net/http"
	"strings"
	"testing"
)

// browserIntegrationFlag 集成测试标志
var browserIntegration = flag.Bool("browser", false, "run browser integration tests")

// skipIfNotBrowserIntegration 跳过非浏览器集成测试
func skipIfNotBrowserIntegration(t *testing.T) {
	if !*browserIntegration {
		t.Skip("skipping browser integration test; use -browser to run")
	}
}

// TestGetCookiesWithBrowser 测试使用浏览器获取Cookie
// 运行方式: go test -v ./pkg/mtop -browser
func TestGetCookiesWithBrowser(t *testing.T) {
	skipIfNotBrowserIntegration(t)

	t.Log("========== GetCookiesWithBrowser 测试 ==========")

	// 测试无头模式
	t.Run("HeadlessMode", func(t *testing.T) {
		t.Log("【测试】无头模式获取Cookie...")

		result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
		if err != nil {
			t.Fatalf("获取Cookie失败: %v", err)
		}

		// 验证结果不为nil
		if result == nil {
			t.Fatal("返回结果为nil")
		}

		// 验证Token不为空
		if result.Token == "" {
			t.Error("Token为空，可能需要先登录闲鱼")
		} else {
			t.Logf("✓ 获取到Token: %s... (长度: %d)", result.Token[:min(10, len(result.Token))], len(result.Token))
		}

		// 验证Cookies不为空
		if len(result.Cookies) == 0 {
			t.Error("Cookies为空")
		} else {
			t.Logf("✓ 获取到 %d 个Cookie", len(result.Cookies))

			// 打印关键Cookie信息
			for _, cookie := range result.Cookies {
				if strings.Contains(cookie.Name, "_m_h5_tk") || strings.Contains(cookie.Name, "cna") || strings.Contains(cookie.Name, "session") {
					t.Logf("  - %s: %s...", cookie.Name, truncateString(cookie.Value, 30))
				}
			}
		}

		// 验证Token一致性
		parsedToken := GetTokenFromCookies(result.Cookies)
		if result.Token != "" && result.Token != parsedToken {
			t.Errorf("Token不一致: result.Token=%s, parsedToken=%s", result.Token, parsedToken)
		}

		// 验证Cookie域名
		validDomain := false
		for _, cookie := range result.Cookies {
			if cookie.Domain == ".goofish.com" || cookie.Domain == "www.goofish.com" || strings.Contains(cookie.Domain, "goofish") {
				validDomain = true
				break
			}
		}
		if !validDomain {
			t.Logf("警告: 未找到 goofish.com 域名的Cookie")
		}
	})

	t.Log("========== 测试完成 ==========")
}

// TestGetCookiesWithBrowserWithDetailedLogging 详细日志测试Cookie获取
func TestGetCookiesWithBrowserWithDetailedLogging(t *testing.T) {
	skipIfNotBrowserIntegration(t)

	t.Log("========== Cookie详细信息测试 ==========")

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	t.Logf("\n【Cookie结果详情】")
	t.Logf("Token: %s...", result.Token[:min(20, len(result.Token))])
	t.Logf("Cookies总数: %d", len(result.Cookies))

	// 分类统计Cookie
	t.Logf("\n【Cookie分类统计】")
	tokenCookies := 0
	sessionCookies := 0
	trackingCookies := 0
	otherCookies := 0

	for _, cookie := range result.Cookies {
		switch {
		case strings.Contains(cookie.Name, "token") || strings.Contains(cookie.Name, "tk"):
			tokenCookies++
		case strings.Contains(cookie.Name, "session") || strings.Contains(cookie.Name, "sid"):
			sessionCookies++
		case strings.Contains(cookie.Name, "track") || strings.Contains(cookie.Name, "cna"):
			trackingCookies++
		default:
			otherCookies++
		}
	}

	t.Logf("  Token相关: %d", tokenCookies)
	t.Logf("  Session相关: %d", sessionCookies)
	t.Logf("  追踪相关: %d", trackingCookies)
	t.Logf("  其他: %d", otherCookies)

	// 获取所有Token信息
	t.Logf("\n【所有Token信息】")
	allTokens := GetAllTokens(result.Cookies)
	for key, val := range allTokens {
		if val != "" {
			t.Logf("  %s: %s...", key, truncateString(val, 20))
		} else {
			t.Logf("  %s: (空)", key)
		}
	}

	// 验证关键字段存在
	t.Run("验证关键字段", func(t *testing.T) {
		hasM_h5_tk := false
		hasM_h5_tk_enc := false
		hasCna := false

		for _, cookie := range result.Cookies {
			switch cookie.Name {
			case "_m_h5_tk":
				hasM_h5_tk = true
				if cookie.Value == "" {
					t.Error("_m_h5_tk 值为空")
				}
			case "_m_h5_tk_enc":
				hasM_h5_tk_enc = true
			case "cna":
				hasCna = true
			}
		}

		t.Logf("✓ _m_h5_tk 存在: %v", hasM_h5_tk)
		t.Logf("✓ _m_h5_tk_enc 存在: %v", hasM_h5_tk_enc)
		t.Logf("✓ cna 存在: %v", hasCna)

		if !hasM_h5_tk && !hasM_h5_tk_enc {
			t.Error("未找到 _m_h5_tk 或 _m_h5_tk_enc Cookie")
		}
	})

	t.Log("\n========== 测试完成 ==========")
}

// TestGetCookiesWithBrowserTokenExtraction 测试Token提取功能
func TestGetCookiesWithBrowserTokenExtraction(t *testing.T) {
	skipIfNotBrowserIntegration(t)

	t.Log("========== Token提取测试 ==========")

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	// 测试1: GetTokenFromCookies
	t.Run("GetTokenFromCookies", func(t *testing.T) {
		token := GetTokenFromCookies(result.Cookies)
		if token == "" {
			t.Error("GetTokenFromCookies 返回空字符串")
		} else {
			t.Logf("✓ Token: %s...", truncateString(token, 20))
		}

		// 验证与result.Token一致
		if token != result.Token {
			t.Errorf("Token不一致: %s vs %s", token, result.Token)
		}
	})

	// 测试2: GetAllTokens
	t.Run("GetAllTokens", func(t *testing.T) {
		allTokens := GetAllTokens(result.Cookies)

		expectedKeys := []string{"_m_h5_tk", "_m_h5_tk_enc"}
		for _, key := range expectedKeys {
			if val, ok := allTokens[key]; !ok {
				t.Errorf("GetAllTokens 缺少键: %s", key)
			} else if val == "" {
				t.Logf("警告: %s 的值为空", key)
			} else {
				t.Logf("✓ %s: %s...", key, truncateString(val, 15))
			}
		}
	})

	// 测试3: Token格式验证
	t.Run("TokenFormatValidation", func(t *testing.T) {
		token := result.Token
		if token == "" {
			t.Skip("Token为空，跳过格式验证")
		}

		// Token通常是32位或更长的字符串
		if len(token) < 20 {
			t.Errorf("Token长度过短: %d, 期望至少20位", len(token))
		}

		// Token通常只包含字母和数字
		for _, r := range token {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				t.Errorf("Token包含非法字符: %c", r)
			}
		}

		t.Logf("✓ Token格式验证通过 (长度: %d)", len(token))
	})

	t.Log("\n========== 测试完成 ==========")
}

// TestGetCookiesWithBrowserConsistency 测试多次获取Cookie的一致性
func TestGetCookiesWithBrowserConsistency(t *testing.T) {
	skipIfNotBrowserIntegration(t)

	t.Log("========== Cookie一致性测试 ==========")

	// 第一次获取
	t.Log("第一次获取Cookie...")
	result1, err1 := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err1 != nil {
		t.Fatalf("第一次获取Cookie失败: %v", err1)
	}
	t.Logf("第一次: Token=%s..., Cookies=%d", truncateString(result1.Token, 15), len(result1.Cookies))

	// 第二次获取
	t.Log("第二次获取Cookie...")
	result2, err2 := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err2 != nil {
		t.Fatalf("第二次获取Cookie失败: %v", err2)
	}
	t.Logf("第二次: Token=%s..., Cookies=%d", truncateString(result2.Token, 15), len(result2.Cookies))

	// 验证Token长度一致
	if len(result1.Token) != len(result2.Token) {
		t.Logf("警告: Token长度不一致: %d vs %d", len(result1.Token), len(result2.Token))
	}

	// 验证Cookie数量一致（允许一定差异）
	if abs(len(result1.Cookies)-len(result2.Cookies)) > 5 {
		t.Logf("警告: Cookie数量差异较大: %d vs %d", len(result1.Cookies), len(result2.Cookies))
	}

	// 验证关键Cookie存在
	t.Run("验证关键Cookie存在", func(t *testing.T) {
		result1HasToken := false
		result2HasToken := false

		for _, c := range result1.Cookies {
			if c.Name == "_m_h5_tk" && c.Value != "" {
				result1HasToken = true
				break
			}
		}

		for _, c := range result2.Cookies {
			if c.Name == "_m_h5_tk" && c.Value != "" {
				result2HasToken = true
				break
			}
		}

		if !result1HasToken || !result2HasToken {
			t.Error("未找到 _m_h5_tk Cookie")
		} else {
			t.Log("✓ 两次都成功获取到 _m_h5_tk Cookie")
		}
	})

	t.Log("\n========== 测试完成 ==========")
}

// TestCookieResultStruct 测试CookieResult结构体
func TestCookieResultStruct(t *testing.T) {
	t.Log("========== CookieResult结构体测试 ==========")

	// 创建测试数据
	testCookies := []*http.Cookie{
		{Name: "_m_h5_tk", Value: "test_token_1234567890", Domain: ".goofish.com"},
		{Name: "cna", Value: "test_cna_value", Domain: ".goofish.com"},
		{Name: "session", Value: "test_session", Domain: ".goofish.com"},
	}

	result := &CookieResult{
		Token:   "test_token",
		Cookies: testCookies,
	}

	// 验证字段
	if result.Token != "test_token" {
		t.Errorf("Token = %s, want test_token", result.Token)
	}

	if len(result.Cookies) != 3 {
		t.Errorf("Cookies length = %d, want 3", len(result.Cookies))
	}

	// 验证Cookie内容
	if result.Cookies[0].Name != "_m_h5_tk" {
		t.Errorf("First cookie name = %s, want _m_h5_tk", result.Cookies[0].Name)
	}

	if result.Cookies[0].Domain != ".goofish.com" {
		t.Errorf("First cookie domain = %s, want .goofish.com", result.Cookies[0].Domain)
	}

	t.Log("✓ CookieResult结构体验证通过")
	t.Log("\n========== 测试完成 ==========")
}

// TestBrowserConfig 测试BrowserConfig结构体
func TestBrowserConfig(t *testing.T) {
	t.Log("========== BrowserConfig测试 ==========")

	tests := []struct {
		name     string
		config   BrowserConfig
		headless bool
	}{
		{
			name:     "默认配置",
			config:   BrowserConfig{},
			headless: false,
		},
		{
			name:     "无头模式",
			config:   BrowserConfig{Headless: true},
			headless: true,
		},
		{
			name:     "有头模式",
			config:   BrowserConfig{Headless: false},
			headless: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Headless != tt.headless {
				t.Errorf("Headless = %v, want %v", tt.config.Headless, tt.headless)
			}
			t.Logf("✓ %s: Headless=%v", tt.name, tt.config.Headless)
		})
	}

	t.Log("\n========== 测试完成 ==========")
}

// TestGetRandomUserAgent 测试随机User-Agent生成
func TestGetRandomUserAgent(t *testing.T) {
	t.Log("========== 随机User-Agent测试 ==========")

	// 生成多次User-Agent，验证多样性
	userAgents := make(map[string]bool)
	for i := 0; i < 100; i++ {
		ua := getRandomUserAgent()
		userAgents[ua] = true

		// 验证User-Agent不为空
		if ua == "" {
			t.Error("getRandomUserAgent() 返回空字符串")
		}

		// 验证User-Agent包含必要的关键字
		if !strings.Contains(ua, "Mozilla") {
			t.Errorf("User-Agent不包含Mozilla: %s", ua)
		}
	}

	// 验证有多种不同的User-Agent
	if len(userAgents) < 5 {
		t.Errorf("User-Agent种类过少: %d", len(userAgents))
	}

	t.Logf("✓ 生成了 %d 种不同的User-Agent", len(userAgents))

	// 打印所有User-Agent种类
	t.Log("\n所有User-Agent种类:")
	for ua := range userAgents {
		t.Logf("  - %s", ua)
	}

	t.Log("\n========== 测试完成 ==========")
}

// TestAntiDetectionScript 测试反检测脚本内容
func TestAntiDetectionScript(t *testing.T) {
	t.Log("========== 反检测脚本测试 ==========")

	script := getAntiDetectionScript()

	// 验证脚本不为空
	if script == "" {
		t.Fatal("反检测脚本为空")
	}

	// 验证脚本包含关键元素
	requiredElements := []string{
		"navigator",
		"webdriver",
		"window.chrome",
		"plugins",
		"languages",
		"deviceMemory",
		"hardwareConcurrency",
	}

	for _, element := range requiredElements {
		if !strings.Contains(script, element) {
			t.Errorf("反检测脚本缺少关键元素: %s", element)
		}
	}

	// 验证脚本格式正确（包含JavaScript关键字）
	jsKeywords := []string{"function", "Object.defineProperty", "return"}
	for _, keyword := range jsKeywords {
		if !strings.Contains(script, keyword) {
			t.Logf("警告: 脚本可能不包含关键字: %s", keyword)
		}
	}

	t.Logf("✓ 反检测脚本验证通过 (长度: %d 字符)", len(script))
	t.Log("\n========== 测试完成 ==========")
}

// 辅助函数

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// abs 返回整数的绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
