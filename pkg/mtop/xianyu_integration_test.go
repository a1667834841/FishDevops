package mtop

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// 集成测试标志
var integration = flag.Bool("integration", false, "run integration tests")

// skipIfNotIntegration 跳过非集成测试
func skipIfNotIntegration(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration test; use -integration to run")
	}
}

// TestIntegrationGuessYouLike 集成测试：获取猜你喜欢
// 运行方式: go test -v ./mtop/... -integration
func TestIntegrationGuessYouLike(t *testing.T) {
	skipIfNotIntegration(t)

	// 使用Playwright获取Cookie
	token, cookies, err := getXianyuCookies()
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	if token == "" {
		t.Fatal("Token为空，可能需要先登录闲鱼")
	}

	t.Logf("获取到Token: %s", token)

	// 创建客户端
	client := NewClient(token, "34839810",
		WithCookies(cookies),
	)

	// 测试1: 获取第一页数据
	t.Run("GetFirstPage", func(t *testing.T) {
		items, err := client.GuessYouLike("", 1)
		if err != nil {
			t.Fatalf("获取第一页失败: %v", err)
		}

		t.Logf("获取到 %d 条商品", len(items))

		// 验证返回数据
		if len(items) == 0 {
			t.Error("未返回任何商品")
		}

		// 验证第一个商品的字段
		if len(items) > 0 {
			item := items[0]
			t.Logf("第一个商品: %s", item.Title)
			t.Logf("  ItemID: %s", item.ItemID)
			t.Logf("  Price: %s", item.Price)
			t.Logf("  CategoryID: %d", item.CategoryID)
			t.Logf("  WantCount: %d", item.WantCount)
			t.Logf("  ViewCount: %d", item.ViewCount)
			t.Logf("  Status: %s", item.Status)
			t.Logf("  ShopLevel: %s", item.ShopLevel)
			t.Logf("  SellerNick: %s", item.SellerNick)
			t.Logf("  SellerCredit: %s", item.SellerCredit)
			t.Logf("  FreeShipping: %v", item.FreeShipping)
			t.Logf("  IsVideo: %v", item.IsVideo)

			if item.ItemID == "" {
				t.Error("ItemID为空")
			}
			if item.Title == "" {
				t.Error("Title为空")
			}
		}
	})

	// 测试2: 获取多页数据
	t.Run("GetMultiplePages", func(t *testing.T) {
		pages := 2
		items, err := client.GuessYouLike("", pages)
		if err != nil {
			t.Fatalf("获取多页数据失败: %v", err)
		}

		t.Logf("获取到 %d 条商品 (期望 %d 页)", len(items), pages)

		if len(items) == 0 {
			t.Error("未返回任何商品")
		}
	})

	// 测试3: 使用machId参数
	t.Run("WithMachID", func(t *testing.T) {
		items, err := client.GuessYouLike("test_mach_id", 1)
		if err != nil {
			t.Fatalf("使用machId获取失败: %v", err)
		}

		t.Logf("使用machId获取到 %d 条商品", len(items))
	})
}

// TestIntegrationClientDo 集成测试：直接调用Do方法
func TestIntegrationClientDo(t *testing.T) {
	skipIfNotIntegration(t)

	token, cookies, err := getXianyuCookies()
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	client := NewClient(token, "34839810",
		WithCookies(cookies),
	)

	// 构建请求数据
	reqData := map[string]interface{}{
		"itemId":     "",
		"pageSize":   30,
		"pageNumber": 1,
		"machId":     "",
	}

	resp, err := client.Do(Request{
		API:    "mtop.taobao.idlehome.home.webpc.feed",
		Data:   reqData,
		Method: "POST",
	})

	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	// 验证响应
	t.Logf("Ret: %v", resp.Ret)

	// 检查是否成功
	success := false
	for _, r := range resp.Ret {
		if r == "SUCCESS" || r == "SUCCESS::调用成功" {
			success = true
			break
		}
	}

	if !success {
		t.Errorf("返回错误: %v", resp.Ret)
	}

	t.Logf("✓ 请求成功")
}

// TestIntegrationWithManualToken 使用手动提供的Token进行测试
// 设置环境变量 XIANYU_TOKEN 后运行
func TestIntegrationWithManualToken(t *testing.T) {
	skipIfNotIntegration(t)

	token := os.Getenv("XIANYU_TOKEN")
	if token == "" {
		t.Skip("设置 XIANYU_TOKEN 环境变量来运行此测试")
	}

	// 从环境变量读取Cookie字符串
	cookieStr := os.Getenv("XIANYU_COOKIES")

	client := NewClient(token, "34839810")
	if cookieStr != "" {
		// 如果提供了Cookie字符串，解析并使用
		client = NewClient(token, "34839810",
			WithCookieString(cookieStr),
		)
	}

	items, err := client.GuessYouLike("", 1)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	t.Logf("获取到 %d 条商品", len(items))

	if len(items) > 0 {
		t.Logf("第一个商品: %s", items[0].Title)
	}
}

// TestIntegrationSignatureConsistency 测试签名一致性
func TestIntegrationSignatureConsistency(t *testing.T) {
	skipIfNotIntegration(t)

	data := `{"itemId":"","pageSize":30,"pageNumber":1,"machId":""}`
	token := "test_token_1234567890"
	appKey := "34839810"

	// 生成两次签名，验证一致性
	result1, err1 := Generate(data, GenerateOptions{
		Token:  token,
		AppKey: appKey,
	})
	if err1 != nil {
		t.Fatalf("第一次生成签名失败: %v", err1)
	}

	result2, err2 := Generate(data, GenerateOptions{
		Token:  token,
		AppKey: appKey,
	})
	if err2 != nil {
		t.Fatalf("第二次生成签名失败: %v", err2)
	}

	// 签名应该相同（token和data相同）
	// 注意：时间戳不同会导致签名不同
	if result1.Token != result2.Token {
		t.Errorf("Token不一致: %s vs %s", result1.Token, result2.Token)
	}

	if result1.Data != result2.Data {
		t.Errorf("Data不一致: %s vs %s", result1.Data, result2.Data)
	}

	t.Logf("签名1: %s (t=%s)", result1.Sign, result1.T)
	t.Logf("签名2: %s (t=%s)", result2.Sign, result2.T)
}

// getXianyuCookies 使用Playwright获取闲鱼Cookie
func getXianyuCookies() (string, []*http.Cookie, error) {
	pw, err := playwright.Run()
	if err != nil {
		return "", nil, err
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", nil, err
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		return "", nil, err
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return "", nil, err
	}
	defer page.Close()

	// 导航到闲鱼
	_, err = page.Goto("https://www.goofish.com", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return "", nil, err
	}

	// 获取Cookies
	cookies, err := context.Cookies()
	if err != nil {
		return "", nil, err
	}

	// 转换Cookie格式
	cookieMaps := make([]map[string]string, len(cookies))
	for i, c := range cookies {
		cookieMaps[i] = map[string]string{
			"name":  c.Name,
			"value": c.Value,
		}
		if c.Domain != "" {
			cookieMaps[i]["domain"] = c.Domain
		}
		if c.Path != "" {
			cookieMaps[i]["path"] = c.Path
		}
	}

	httpCookies := ConvertMapSliceToHTTPCookies(cookieMaps)
	token := GetTokenFromCookies(httpCookies)

	return token, httpCookies, nil
}

// TestIntegrationFullFlow 完整流程集成测试
func TestIntegrationFullFlow(t *testing.T) {
	skipIfNotIntegration(t)

	t.Log("开始完整流程集成测试...")

	// 步骤1: 获取Cookie
	t.Log("步骤1: 获取Cookie")
	token, cookies, err := getXianyuCookies()
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}
	if token == "" {
		t.Fatal("未获取到Token")
	}
	t.Logf("✓ 获取到Token: %s...", token[:10])

	// 步骤2: 创建客户端
	t.Log("步骤2: 创建客户端")
	client := NewClient(token, "34839810", WithCookies(cookies))
	t.Log("✓ 客户端创建成功")

	// 步骤3: 生成签名
	t.Log("步骤3: 生成签名")
	reqData := `{"itemId":"","pageSize":30,"pageNumber":1,"machId":""}`
	signResult, err := Generate(reqData, GenerateOptions{
		Token:  token,
		AppKey: "34839810",
	})
	if err != nil {
		t.Fatalf("生成签名失败: %v", err)
	}
	t.Logf("✓ 签名生成成功: %s", signResult.Sign)

	// 步骤4: 发送请求
	t.Log("步骤4: 发送API请求")
	resp, err := client.Do(Request{
		API:    "mtop.taobao.idlehome.home.webpc.feed",
		Data:   reqData,
		Method: "POST",
	})
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	// 检查响应是否成功
	success := false
	for _, r := range resp.Ret {
		if r == "SUCCESS" || r == "SUCCESS::调用成功" {
			success = true
			break
		}
	}

	if !success {
		t.Errorf("API返回错误: %v", resp.Ret)
	} else {
		t.Logf("✓ API请求成功")
	}

	// 步骤5: 解析数据
	t.Log("步骤5: 解析响应数据")
	var result struct {
		Data struct {
			FeedsCount int `json:"feedsCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &result.Data); err == nil {
		t.Logf("✓ 解析成功，商品数量: %d", result.Data.FeedsCount)
	}

	// 步骤6: 使用高级API
	t.Log("步骤6: 使用高级API获取数据")
	items, err := client.GuessYouLike("", 1)
	if err != nil {
		t.Fatalf("GuessYouLike失败: %v", err)
	}
	t.Logf("✓ 高级API成功，获取到 %d 条商品", len(items))

	if len(items) > 0 {
		t.Logf("示例商品: %s - ¥%s", items[0].Title, items[0].Price)
	}

	t.Log("\n✅ 完整流程测试通过！")
}

// TestIntegrationCookieParsing 测试Cookie解析功能
func TestIntegrationCookieParsing(t *testing.T) {
	skipIfNotIntegration(t)

	token, cookies, err := getXianyuCookies()
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	t.Logf("获取到 %d 个Cookie", len(cookies))

	// 测试解析token
	parsedToken := GetTokenFromCookies(cookies)
	if parsedToken != token {
		t.Errorf("Token解析不一致: %s vs %s", parsedToken, token)
	}
	t.Logf("✓ Token解析正确: %s", parsedToken)

	// 测试获取所有token信息
	allTokens := GetAllTokens(cookies)
	t.Logf("✓ 获取到以下token信息:")
	for key, val := range allTokens {
		if val != "" {
			t.Logf("  %s: %s...", key, truncateString(val, 20))
		}
	}

	// 测试Cookie字符串解析
	cookieStr := "_m_h5_tk=" + token + "_1234567890; cna=test_value; other=value"
	parsedFromStr := GetTokenFromCookieString(cookieStr)
	if parsedFromStr != token {
		t.Errorf("从字符串解析Token不一致: %s vs %s", parsedFromStr, token)
	}
	t.Logf("✓ 从字符串解析Token正确")
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TestIntegrationErrorHandling 测试错误处理
func TestIntegrationErrorHandling(t *testing.T) {
	skipIfNotIntegration(t)

	token, cookies, err := getXianyuCookies()
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	client := NewClient(token, "34839810", WithCookies(cookies))

	t.Run("InvalidAPI", func(t *testing.T) {
		_, err := client.Do(Request{
			API:    "invalid.api.name",
			Data:   map[string]string{},
			Method: "POST",
		})
		if err == nil {
			t.Error("期望返回错误，但请求成功了")
		} else {
			t.Logf("✓ 正确返回错误: %v", err)
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		invalidClient := NewClient("", "34839810")
		_, err := invalidClient.Do(Request{
			API:    "mtop.taobao.idlehome.home.webpc.feed",
			Data:   map[string]string{},
			Method: "POST",
		})
		if err == nil {
			t.Error("空token应该返回错误")
		} else {
			t.Logf("✓ 空token正确返回错误")
		}
	})
}
