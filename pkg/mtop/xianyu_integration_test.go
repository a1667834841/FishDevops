package mtop

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
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
// 运行方式: go test -v ./pkg/mtop -integration
func TestIntegrationGuessYouLike(t *testing.T) {
	skipIfNotIntegration(t)

	// 使用浏览器获取Cookie
	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	if result.Token == "" {
		t.Fatal("Token为空，可能需要先登录闲鱼")
	}

	t.Logf("获取到Token: %s", result.Token)

	// 创建客户端
	client := NewClient(result.Token, "34839810",
		WithCookies(result.Cookies),
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

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	client := NewClient(result.Token, "34839810",
		WithCookies(result.Cookies),
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

// TestIntegrationFullFlow 完整流程集成测试
func TestIntegrationFullFlow(t *testing.T) {
	skipIfNotIntegration(t)

	t.Log("开始完整流程集成测试...")

	// 步骤1: 获取Cookie
	t.Log("步骤1: 获取Cookie")
	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}
	if result.Token == "" {
		t.Fatal("未获取到Token")
	}
	t.Logf("✓ 获取到Token: %s...", result.Token[:10])

	// 步骤2: 创建客户端
	t.Log("步骤2: 创建客户端")
	client := NewClient(result.Token, "34839810", WithCookies(result.Cookies))
	t.Log("✓ 客户端创建成功")

	// 步骤3: 生成签名
	t.Log("步骤3: 生成签名")
	reqData := `{"itemId":"","pageSize":30,"pageNumber":1,"machId":""}`
	signResult, err := Generate(reqData, GenerateOptions{
		Token:  result.Token,
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
	var respData struct {
		Data struct {
			FeedsCount int `json:"feedsCount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &respData.Data); err == nil {
		t.Logf("✓ 解析成功，商品数量: %d", respData.Data.FeedsCount)
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

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	t.Logf("获取到 %d 个Cookie", len(result.Cookies))

	// 测试解析token
	parsedToken := GetTokenFromCookies(result.Cookies)
	if parsedToken != result.Token {
		t.Errorf("Token解析不一致: %s vs %s", parsedToken, result.Token)
	}
	t.Logf("✓ Token解析正确: %s", parsedToken)

	// 测试获取所有token信息
	allTokens := GetAllTokens(result.Cookies)
	t.Logf("✓ 获取到以下token信息:")
	for key, val := range allTokens {
		if val != "" {
			t.Logf("  %s: %s...", key, truncateString(val, 20))
		}
	}

	// 测试Cookie字符串解析
	cookieStr := "_m_h5_tk=" + result.Token + "_1234567890; cna=test_value; other=value"
	parsedFromStr := GetTokenFromCookieString(cookieStr)
	if parsedFromStr != result.Token {
		t.Errorf("从字符串解析Token不一致: %s vs %s", parsedFromStr, result.Token)
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

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Fatalf("获取Cookie失败: %v", err)
	}

	client := NewClient(result.Token, "34839810", WithCookies(result.Cookies))

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

// TestIntegrationFetchItemDetail 集成测试：获取商品详情
// 测试真实商品ID: 963943643587
// 运行方式: go test -v ./pkg/mtop -integration -run TestIntegrationFetchItemDetail
func TestIntegrationFetchItemDetail(t *testing.T) {
	// skipIfNotIntegration(t)

	t.Log("========== 商品详情集成测试 ==========")
	t.Log("测试商品ID: 963943643587")

	// 步骤1: 获取Cookie和Token
	t.Log("\n【步骤1】打开闲鱼网站获取Token...")

	result, err := GetCookiesWithBrowser(BrowserConfig{Headless: true})
	if err != nil {
		t.Logf("❌ 获取Cookie失败: %v", err)
		t.Logf("\n提示: 如果没有登录，可能无法获取有效的token")
		t.Logf("建议使用 TestIntegrationFetchItemDetailWithManualToken 测试")
		t.Logf("\n获取Token的方法:")
		t.Logf("1. 打开浏览器访问 https://www.goofish.com")
		t.Logf("2. 登录后打开开发者工具 (F12)")
		t.Logf("3. 进入 Application/存储 → Cookies → https://www.goofish.com")
		t.Logf("4. 找到 _m_h5_tk 的值，格式为 token_timestamp")
		t.Logf("5. 复制下划线前的部分作为 Token")
		t.Skip("需要手动提供Token来运行此测试")
	}

	if result.Token == "" {
		t.Log("❌ 未获取到Token")
		t.Log("获取到的Cookie列表:")
		for _, c := range result.Cookies {
			t.Logf("  - %s: %s", c.Name, truncateString(c.Value, 30))
		}
		t.Skip("未找到 _m_h5_tk Cookie，需要先登录闲鱼账号")
	}

	t.Logf("✓ 成功获取Token: %s... (长度: %d)", result.Token[:10], len(result.Token))
	t.Logf("✓ 获取到 %d 个Cookie", len(result.Cookies))

	// 步骤2: 创建客户端
	t.Log("\n【步骤2】创建MTOP客户端...")
	client := NewClient(result.Token, "34839810", WithCookies(result.Cookies))
	t.Log("✓ 客户端创建成功")

	// 步骤3: 获取商品详情
	testItemID := "963943643587"
	t.Logf("\n【步骤3】获取商品详情...")
	t.Logf("正在请求商品ID: %s", testItemID)

	detail, err := client.FetchItemDetail(testItemID)
	if err != nil {
		t.Logf("❌ 获取商品详情失败: %v", err)
		t.Fatalf("API请求失败")
	}

	t.Log("✓ 成功获取商品详情")

	// 步骤4: 打印商品信息
	t.Log("\n【步骤4】商品详情数据:")
	t.Log("==============================================")
	t.Logf("【基础信息】")
	t.Logf("  商品ID: %s", detail.ItemID)
	t.Logf("  标题: %s", detail.Title)
	if detail.SubTitle != "" {
		t.Logf("  副标题: %s", detail.SubTitle)
	}
	if detail.Desc != "" {
		t.Logf("  简述: %s", detail.Desc)
	}
	t.Logf("  分类ID: %d", detail.CategoryID)

	t.Logf("\n【价格信息】")
	t.Logf("  售价: %s", detail.Price)
	if detail.PriceOriginal != "" {
		t.Logf("  原价: %s", detail.PriceOriginal)
	}
	if detail.UnitPrice != "" {
		t.Logf("  单价: %s", detail.UnitPrice)
	}

	t.Logf("\n【卖家信息】")
	t.Logf("  卖家ID: %s", detail.SellerID)
	t.Logf("  卖家昵称: %s", detail.SellerNick)
	if detail.AvatarURL != "" {
		t.Logf("  头像: %s", detail.AvatarURL)
	}
	if detail.ShopLevel != "" {
		t.Logf("  店铺级别: %s", detail.ShopLevel)
	}

	t.Logf("\n【商品状态】")
	t.Logf("  商品状态: %s", detail.Status)
	if detail.WantCount > 0 {
		t.Logf("  想要人数: %d", detail.WantCount)
	}
	if detail.ViewCount > 0 {
		t.Logf("  浏览次数: %d", detail.ViewCount)
	}
	if detail.CollectCount > 0 {
		t.Logf("  收藏次数: %d", detail.CollectCount)
	}
	if detail.ChatCount > 0 {
		t.Logf("  咨询次数: %d", detail.ChatCount)
	}

	if detail.Location != "" {
		t.Logf("\n【地址信息】")
		t.Logf("  位置: %s", detail.Location)
		if detail.Area != "" {
			t.Logf("  区域: %s", detail.Area)
		}
	}

	t.Logf("\n【商品属性】")
	if detail.Condition != "" {
		condition := detail.Condition
		if detail.IsNew {
			condition += " (全新)"
		}
		t.Logf("  成色: %s", condition)
	}
	t.Logf("  包邮: %v", detail.FreeShipping)
	if len(detail.Tags) > 0 {
		t.Logf("  标签: %s", formatTags(detail.Tags))
	}

	if detail.PublishTime != "" {
		t.Logf("\n【时间信息】")
		t.Logf("  发布时间: %s", detail.PublishTime)
		if detail.ModifiedTime != "" {
			t.Logf("  修改时间: %s", detail.ModifiedTime)
		}
	}

	if len(detail.ImageList) > 0 {
		t.Logf("\n【图片列表】(%d张)", len(detail.ImageList))
		for i, img := range detail.ImageList {
			if i < 3 { // 只显示前3张
				t.Logf("  %d. %s", i+1, img)
			}
		}
		if len(detail.ImageList) > 3 {
			t.Logf("  ... 还有 %d 张图片", len(detail.ImageList)-3)
		}
	}

	t.Log("==============================================")

	// 步骤5: 验证关键字段
	t.Log("\n【步骤5】验证关键字段...")
	t.Run("验证关键字段", func(t *testing.T) {
		// 基础字段验证
		if detail.ItemID == "" {
			t.Error("❌ ItemID 为空")
		} else {
			t.Logf("✓ ItemID: %s", detail.ItemID)
		}

		if detail.ItemID != testItemID {
			t.Errorf("❌ ItemID 不匹配: 期望 %s, 实际 %s", testItemID, detail.ItemID)
		} else {
			t.Logf("✓ ItemID 匹配")
		}

		if detail.Title == "" {
			t.Error("❌ Title 为空")
		} else {
			t.Logf("✓ Title: %s", detail.Title)
		}

		if detail.Status == "" {
			t.Error("❌ Status 为空")
		} else {
			t.Logf("✓ Status: %s", detail.Status)
		}

		if detail.Price == "" {
			t.Error("❌ Price 为空")
		} else {
			t.Logf("✓ Price: %s", detail.Price)
		}

		// 卖家信息验证
		if detail.SellerID == "" {
			t.Error("❌ SellerID 为空")
		} else {
			t.Logf("✓ SellerID: %s", detail.SellerID)
		}

		if detail.SellerNick == "" {
			t.Error("❌ SellerNick 为空")
		} else {
			t.Logf("✓ SellerNick: %s", detail.SellerNick)
		}

		// 地址信息验证
		if detail.Location == "" {
			t.Error("❌ Location 为空")
		} else {
			t.Logf("✓ Location: %s", detail.Location)
		}

		// 时间戳验证
		if detail.PublishTimeTS == 0 {
			t.Error("❌ PublishTimeTS 为 0")
		} else {
			t.Logf("✓ PublishTimeTS: %d", detail.PublishTimeTS)
		}

		t.Logf("\n✅ 所有关键字段验证通过")
	})

	// 步骤6: 验证数据合理性
	t.Log("\n【步骤6】验证数据合理性...")
	t.Run("验证数据合理性", func(t *testing.T) {
		// 想要人数应该是非负数
		if detail.WantCount < 0 {
			t.Errorf("❌ WantCount 不合理: %d", detail.WantCount)
		} else {
			t.Logf("✓ WantCount: %d", detail.WantCount)
		}

		// 浏览次数应该大于想要人数（通常情况）
		if detail.ViewCount > 0 && detail.ViewCount < detail.WantCount {
			t.Logf("⚠️  警告: ViewCount(%d) 小于 WantCount(%d)", detail.ViewCount, detail.WantCount)
		}

		// 图片列表应该至少包含主图
		if detail.ImageURL == "" && len(detail.ImageList) == 0 {
			t.Error("❌ 没有任何图片信息")
		} else {
			if detail.ImageURL != "" {
				t.Logf("✓ 主图URL: %s", detail.ImageURL)
			}
			if len(detail.ImageList) > 0 {
				t.Logf("✓ 图片数量: %d", len(detail.ImageList))
			}
		}

		// 分类ID应该在合理范围内
		if detail.CategoryID <= 0 {
			t.Errorf("❌ CategoryID 不合理: %d", detail.CategoryID)
		} else {
			t.Logf("✓ CategoryID: %d", detail.CategoryID)
		}

		t.Logf("\n✅ 数据合理性验证通过")
	})

	// 步骤7: 验证JSON序列化
	t.Log("\n【步骤7】验证JSON序列化...")
	t.Run("验证JSON序列化", func(t *testing.T) {
		data, err := json.MarshalIndent(detail, "", "  ")
		if err != nil {
			t.Fatalf("❌ JSON序列化失败: %v", err)
		}

		t.Logf("✓ JSON序列化成功，数据长度: %d 字节", len(data))

		// 反序列化验证
		var decoded ItemDetail
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("❌ JSON反序列化失败: %v", err)
		}

		// 验证关键字段
		if decoded.ItemID != detail.ItemID {
			t.Errorf("❌ 序列化后ItemID不一致: %s vs %s", decoded.ItemID, detail.ItemID)
		}
		if decoded.Title != detail.Title {
			t.Errorf("❌ 序列化后Title不一致: %s vs %s", decoded.Title, detail.Title)
		}
		if decoded.Price != detail.Price {
			t.Errorf("❌ 序列化后Price不一致: %s vs %s", decoded.Price, detail.Price)
		}

		t.Logf("✓ JSON序列化/反序列化验证通过")
	})

	t.Log("\n==============================================")
	t.Log("✅ 商品详情集成测试全部通过！")
	t.Log("==============================================")

	// 步骤8: 数据分析字段报告
	t.Log("\n【步骤8】数据分析字段报告...")
	AnalyzeItemDetailForDataAnalysis(detail)
}


// formatTags 格式化标签列表
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return "无"
	}
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}
	return result
}
