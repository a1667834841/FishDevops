package feishu

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// 集成测试标志
var integration = flag.Bool("integration", false, "run integration tests")

// skipIfNotIntegration 跳过非集成测试
func skipIfNotIntegration(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration test; use -integration to run")
	}
}

// getFeishuTestConfig 从环境变量获取测试配置
func getFeishuTestConfig(t *testing.T) (appID, appSecret, appToken string) {
	appID = os.Getenv("FEISHU_APP_ID")
	appSecret = os.Getenv("FEISHU_APP_SECRET")
	appToken = os.Getenv("FEISHU_APP_TOKEN")

	if appID == "" || appSecret == "" || appToken == "" {
		t.Skip("设置 FEISHU_APP_ID, FEISHU_APP_SECRET, FEISHU_APP_TOKEN 环境变量来运行飞书集成测试")
	}

	return appID, appSecret, appToken
}

// TestIntegrationGetTenantAccessToken 集成测试：获取租户访问令牌
func TestIntegrationGetTenantAccessToken(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, _ := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	token, err := client.GetTenantAccessToken()
	if err != nil {
		t.Fatalf("获取租户访问令牌失败: %v", err)
	}

	if token == "" {
		t.Fatal("Token为空")
	}

	t.Logf("✓ 成功获取租户访问令牌: %s...", token[:20])
}

// TestIntegrationGetBitableTableInfos 集成测试：获取多维表格列表
func TestIntegrationGetBitableTableInfos(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	tables, err := client.GetBitableTableInfos(appToken)
	if err != nil {
		t.Fatalf("获取表格列表失败: %v", err)
	}

	t.Logf("✓ 成功获取表格列表，共 %d 个表格", len(tables))

	for i, table := range tables {
		t.Logf("  %d. 表名: %s, TableID: %s", i+1, table.Name, table.TableID)
	}
}

// TestIntegrationGetTableFields 集成测试：获取表格字段
func TestIntegrationGetTableFields(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	// 先获取表格列表
	tables, err := client.GetBitableTableInfos(appToken)
	if err != nil {
		t.Fatalf("获取表格列表失败: %v", err)
	}

	if len(tables) == 0 {
		t.Fatal("没有可用的表格")
	}

	// 使用第一个表格
	tableID := tables[0].TableID
	t.Logf("使用表格: %s (%s)", tables[0].Name, tableID)

	fields, err := client.GetTableFields(appToken, tableID)
	if err != nil {
		t.Fatalf("获取字段列表失败: %v", err)
	}

	t.Logf("✓ 成功获取字段列表，共 %d 个字段", len(fields))

	for fieldName, fieldID := range fields {
		t.Logf("  字段名: %s, FieldID: %s", fieldName, fieldID)
	}
}

// TestIntegrationGetOrCreateTableByDate 集成测试：根据日期获取或创建表格
func TestIntegrationGetOrCreateTableByDate(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 使用今天的日期进行测试
	today := time.Now()
	tableName := today.Format("2006-01-02")

	t.Run("获取或创建今天的表格", func(t *testing.T) {
		tableID, created, err := service.GetOrCreateTableByDate(today)
		if err != nil {
			t.Fatalf("获取或创建表格失败: %v", err)
		}

		if tableID == "" {
			t.Fatal("TableID为空")
		}

		if created {
			t.Logf("✓ 成功创建新表格: %s, TableID: %s", tableName, tableID)
		} else {
			t.Logf("✓ 使用已存在表格: %s, TableID: %s", tableName, tableID)
		}

		// 验证可以再次获取同一个表格
		tableID2, created2, err := service.GetOrCreateTableByDate(today)
		if err != nil {
			t.Fatalf("再次获取表格失败: %v", err)
		}

		if tableID != tableID2 {
			t.Errorf("两次获取的TableID不一致: %s vs %s", tableID, tableID2)
		}

		if created2 {
			t.Error("第二次获取不应该创建新表格")
		}

		t.Logf("✓ 验证通过: 两次获取的TableID一致")
	})

	// 测试获取不存在的日期
	t.Run("获取不存在的表格", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 0, 7) // 7天后
		tableID, err := service.GetTableByDate(futureDate)
		if err != nil {
			t.Fatalf("获取表格失败: %v", err)
		}

		if tableID != "" {
			t.Errorf("不存在的表格应该返回空字符串，得到: %s", tableID)
		}

		t.Logf("✓ 正确处理不存在的表格")
	})

	// 测试表格是否存在
	t.Run("检查表格是否存在", func(t *testing.T) {
		exists, err := service.TableExists(today)
		if err != nil {
			t.Fatalf("检查表格存在失败: %v", err)
		}

		if !exists {
			t.Error("今天的表格应该存在")
		}

		t.Logf("✓ 表格存在检查通过")
	})
}

// TestIntegrationEnsureTableFields 集成测试：确保表格字段存在
func TestIntegrationEnsureTableFields(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 获取或创建表格
	today := time.Now()
	tableID, created, err := service.GetOrCreateTableByDate(today)
	if err != nil {
		t.Fatalf("获取或创建表格失败: %v", err)
	}

	t.Logf("使用表格: %s (新创建: %v)", tableID, created)

	// 确保所有字段存在
	err = service.EnsureTableFields(tableID)
	if err != nil {
		t.Fatalf("确保字段存在失败: %v", err)
	}

	t.Logf("✓ 成功确保表格字段存在")

	// 验证字段确实存在
	fields, err := client.GetTableFields(appToken, tableID)
	if err != nil {
		t.Fatalf("获取字段列表失败: %v", err)
	}

	t.Logf("✓ 表格当前共有 %d 个字段", len(fields))

	// 检查关键字段是否存在
	expectedFields := []string{"商品ID", "商品标题", "价格", "想要人数"}
	missingFields := []string{}
	for _, expected := range expectedFields {
		if _, exists := fields[expected]; !exists {
			missingFields = append(missingFields, expected)
		}
	}

	if len(missingFields) > 0 {
		t.Errorf("缺少字段: %v", missingFields)
	} else {
		t.Logf("✓ 所有关键字段都存在")
	}
}

// TestIntegrationPushProductsToDateTable 集成测试：推送数据到日期表格
func TestIntegrationPushProductsToDateTable(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 创建测试商品数据
	testProducts := []Product{
		{
			ItemID:     "integration_test_item_001",
			Title:      "集成测试商品",
			Price:      "99.99",
			PriceNumber: 99.99,
			WantCnt:    5,
			PublishTime: time.Now().Format("2006-01-02 15:04:05"),
			SellerNick: "测试卖家",
			SellerCity: "测试城市",
			FreeShip:   "是",
			Tags:       "测试标签",
			CoverURL:   "https://example.com/cover.jpg",
			DetailURL:  "https://example.com/detail",
		},
		{
			ItemID:     "integration_test_item_002",
			Title:      "集成测试商品2",
			Price:      "199.99",
			PriceNumber: 199.99,
			WantCnt:    10,
			PublishTime: time.Now().Format("2006-01-02 15:04:05"),
			SellerNick: "测试卖家2",
			SellerCity: "测试城市2",
			FreeShip:   "否",
			Tags:       "测试标签2",
			CoverURL:   "https://example.com/cover2.jpg",
			DetailURL:  "https://example.com/detail2",
		},
	}

	today := time.Now()

	t.Run("推送数据到今天的表格", func(t *testing.T) {
		resp, err := service.PushProductsToDateTable(today, testProducts)
		if err != nil {
			t.Fatalf("推送数据到日期表格失败: %v", err)
		}

		if !resp.Success {
			t.Fatalf("推送失败: %s", resp.Message)
		}

		t.Logf("✓ 成功推送数据到表格")
		t.Logf("  创建记录数: %d", resp.Data.RecordsCreated)
		t.Logf("  TableToken: %s", resp.Data.TableToken)

		if resp.Data.RecordsCreated != len(testProducts) {
			t.Errorf("创建记录数不匹配: got %d, want %d", resp.Data.RecordsCreated, len(testProducts))
		}
	})

	// 测试推送单个商品
	t.Run("推送单个商品", func(t *testing.T) {
		singleProduct := []Product{
			{
				ItemID:     "integration_test_item_single",
				Title:      "单个测试商品",
				Price:      "88.88",
				PriceNumber: 88.88,
				WantCnt:    3,
				SellerNick: "单个测试卖家",
			},
		}

		resp, err := service.PushProductsToDateTable(today, singleProduct)
		if err != nil {
			t.Fatalf("推送单个商品失败: %v", err)
		}

		if !resp.Success {
			t.Fatalf("推送失败: %s", resp.Message)
		}

		t.Logf("✓ 成功推送单个商品")
	})
}

// TestIntegrationPushProductsToTodayTable 集成测试：推送数据到今天表格
func TestIntegrationPushProductsToTodayTable(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	testProducts := []Product{
		{
			ItemID:     "today_test_item_001",
			Title:      "今天测试商品",
			Price:      "66.66",
			PriceNumber: 66.66,
			WantCnt:    2,
			SellerNick: "今天测试卖家",
		},
	}

	resp, err := service.PushProductsToTodayTable(testProducts)
	if err != nil {
		t.Fatalf("推送数据到今天表格失败: %v", err)
	}

	if !resp.Success {
		t.Fatalf("推送失败: %s", resp.Message)
	}

	t.Logf("✓ 成功推送数据到今天表格")
	t.Logf("  创建记录数: %d", resp.Data.RecordsCreated)
}

// TestIntegrationFullFlow 完整流程集成测试
func TestIntegrationFullFlow(t *testing.T) {
	skipIfNotIntegration(t)

	t.Log("开始飞书多维表格完整流程集成测试...")

	appID, appSecret, appToken := getFeishuTestConfig(t)

	// 步骤1: 创建客户端
	t.Log("步骤1: 创建飞书客户端")
	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})
	t.Log("✓ 客户端创建成功")

	// 步骤2: 获取访问令牌
	t.Log("步骤2: 获取租户访问令牌")
	token, err := client.GetTenantAccessToken()
	if err != nil {
		t.Fatalf("获取访问令牌失败: %v", err)
	}
	t.Logf("✓ 成功获取访问令牌: %s...", token[:20])

	// 步骤3: 获取表格列表
	t.Log("步骤3: 获取现有表格列表")
	tables, err := client.GetBitableTableInfos(appToken)
	if err != nil {
		t.Fatalf("获取表格列表失败: %v", err)
	}
	t.Logf("✓ 成功获取表格列表，共 %d 个表格", len(tables))

	// 步骤4: 创建 BitableService
	t.Log("步骤4: 创建 BitableService")
	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})
	t.Log("✓ BitableService 创建成功")

	// 步骤5: 获取或创建今天的表格
	t.Log("步骤5: 获取或创建今天的表格")
	today := time.Now()
	tableID, created, err := service.GetOrCreateTableByDate(today)
	if err != nil {
		t.Fatalf("获取或创建表格失败: %v", err)
	}
	if created {
		t.Logf("✓ 成功创建新表格: %s", tableID)
	} else {
		t.Logf("✓ 使用已存在表格: %s", tableID)
	}

	// 步骤6: 确保字段存在
	t.Log("步骤6: 确保所有字段存在")
	err = service.EnsureTableFields(tableID)
	if err != nil {
		t.Fatalf("确保字段存在失败: %v", err)
	}
	t.Log("✓ 字段检查完成")

	// 步骤7: 推送测试数据
	t.Log("步骤7: 推送测试数据到表格")
	testProducts := []Product{
		{
			ItemID:     "full_flow_test_001",
			Title:      "完整流程测试商品",
			Price:      "123.45",
			PriceNumber: 123.45,
			WantCnt:    7,
			PublishTime: time.Now().Format("2006-01-02 15:04:05"),
			SellerNick: "完整流程测试卖家",
			SellerCity: "完整流程测试城市",
			FreeShip:   "是",
			Tags:       "完整流程测试",
		},
	}

	resp, err := service.PushProductsToDateTable(today, testProducts)
	if err != nil {
		t.Fatalf("推送数据失败: %v", err)
	}

	if !resp.Success {
		t.Fatalf("推送失败: %s", resp.Message)
	}

	t.Logf("✓ 成功推送 %d 条记录", resp.Data.RecordsCreated)

	// 步骤8: 验证数据
	t.Log("步骤8: 验证推送的数据")
	fields, err := client.GetTableFields(appToken, tableID)
	if err != nil {
		t.Fatalf("获取字段失败: %v", err)
	}
	t.Logf("✓ 表格共有 %d 个字段", len(fields))

	t.Log("\n✅ 飞书多维表格完整流程测试通过！")
}

// TestIntegrationTableOperations 集成测试：表格操作
func TestIntegrationTableOperations(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	t.Run("检查多个日期的表格", func(t *testing.T) {
		// 测试今天、昨天、明天
		dates := []struct {
			name time.Time
			desc string
		}{
			{time.Now().AddDate(0, 0, -1), "昨天"},
			{time.Now(), "今天"},
			{time.Now().AddDate(0, 0, 1), "明天"},
		}

		for _, d := range dates {
			exists, err := service.TableExists(d.name)
			if err != nil {
				t.Errorf("检查%s表格失败: %v", d.desc, err)
				continue
			}
			tableName := d.name.Format("2006-01-02")
			if exists {
				t.Logf("✓ %s表格(%s)存在", d.desc, tableName)
			} else {
				t.Logf("  %s表格(%s)不存在", d.desc, tableName)
			}
		}
	})

	t.Run("创建和获取不同日期的表格", func(t *testing.T) {
		// 使用未来的日期（不太可能已存在）
		futureDate := time.Now().AddDate(0, 0, 30) // 30天后
		tableName := futureDate.Format("2006-01-02")

		tableID, created, err := service.GetOrCreateTableByDate(futureDate)
		if err != nil {
			t.Fatalf("获取或创建表格失败: %v", err)
		}

		if created {
			t.Logf("✓ 成功创建未来日期表格: %s, TableID: %s", tableName, tableID)
		} else {
			t.Logf("  表格已存在: %s, TableID: %s", tableName, tableID)
		}

		// 再次获取应该返回已存在的表格
		tableID2, created2, err := service.GetOrCreateTableByDate(futureDate)
		if err != nil {
			t.Fatalf("再次获取表格失败: %v", err)
		}

		if tableID != tableID2 {
			t.Errorf("两次获取的TableID不一致")
		}

		if created2 {
			t.Error("第二次获取不应该创建新表格")
		}

		t.Logf("✓ 验证通过: 重复获取返回相同的表格")
	})
}

// TestIntegrationProductBuilder 集成测试：使用 ProductBuilder 构建商品
func TestIntegrationProductBuilder(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 使用 ProductBuilder 构建测试商品
	product := NewProductBuilder().
		WithItemID("builder_test_001").
		WithTitle("Builder测试商品").
		WithPrice("299.99").
		WithOriginalPrice("399.99").
		WithWantCount(15).
		WithPublishTime(time.Now().Format("2006-01-02 15:04:05")).
		WithSellerInfo("Builder测试卖家", "Builder测试城市").
		WithFreeShip(true).
		WithTags("Builder测试标签").
		WithURLs("https://example.com/builder_cover.jpg", "https://example.com/builder_detail").
		WithExposureHeat(500).
		Build()

	// 验证构建的商品
	if product.ItemID != "builder_test_001" {
		t.Errorf("ItemID错误: got %s", product.ItemID)
	}
	if product.PriceNumber != 299.99 {
		t.Errorf("PriceNumber错误: got %f", product.PriceNumber)
	}
	if product.OriginalPriceNumber != 399.99 {
		t.Errorf("OriginalPriceNumber错误: got %f", product.OriginalPriceNumber)
	}
	if product.FreeShip != "是" {
		t.Errorf("FreeShip错误: got %s", product.FreeShip)
	}

	t.Log("✓ ProductBuilder 构建商品成功")

	// 推送到飞书
	resp, err := service.PushProductsToTodayTable([]Product{product})
	if err != nil {
		t.Fatalf("推送商品失败: %v", err)
	}

	if !resp.Success {
		t.Fatalf("推送失败: %s", resp.Message)
	}

	t.Logf("✓ 成功推送 Builder 构建的商品到飞书")
}

// TestIntegrationErrorHandling 集成测试：错误处理
func TestIntegrationErrorHandling(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	t.Run("无效的 AppToken", func(t *testing.T) {
		invalidService := NewBitableService(client, BitableConfig{
			AppToken: "invalid_app_token",
		})

		_, err := invalidService.GetTableByDate(time.Now())
		if err == nil {
			t.Error("使用无效的 AppToken 应该返回错误")
		} else {
			t.Logf("✓ 正确处理无效 AppToken: %v", err)
		}
	})

	t.Run("空的商品列表", func(t *testing.T) {
		service := NewBitableService(client, BitableConfig{
			AppToken: appToken,
		})

		_, err := service.PushProductsToTodayTable([]Product{})
		if err == nil {
			t.Log("空商品列表被接受（取决于实现）")
		} else {
			t.Logf("✓ 正确处理空商品列表: %v", err)
		}
	})

	t.Run("无效的日期格式", func(t *testing.T) {
		// time.Time 不会无效，这里测试边界情况
		zeroTime := time.Time{}
		service := NewBitableService(client, BitableConfig{
			AppToken: appToken,
		})

		_, _, err := service.GetOrCreateTableByDate(zeroTime)
		if err != nil {
			t.Logf("✓ 正确处理零时间: %v", err)
		}
	})
}

// TestIntegrationConcurrentOperations 集成测试：并发操作
func TestIntegrationConcurrentOperations(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	today := time.Now()

	// 并发获取同一个表格
	t.Run("并发获取表格", func(t *testing.T) {
		results := make(chan string, 5)
		errors := make(chan error, 5)

		for i := 0; i < 5; i++ {
			go func() {
				tableID, _, err := service.GetOrCreateTableByDate(today)
				if err != nil {
					errors <- err
				} else {
					results <- tableID
				}
			}()
		}

		// 收集结果
		tableIDs := make([]string, 0, 5)
		errorCount := 0
		for i := 0; i < 5; i++ {
			select {
			case tableID := <-results:
				tableIDs = append(tableIDs, tableID)
			case err := <-errors:
				errorCount++
				t.Logf("并发操作中的错误: %v", err)
			}
		}

		if len(tableIDs) > 0 {
			t.Logf("✓ 并发获取表格成功，获取到 %d 个结果", len(tableIDs))
			// 验证所有结果都是同一个表格
			firstID := tableIDs[0]
			for _, id := range tableIDs {
				if id != firstID {
					t.Errorf("并发获取的 TableID 不一致: %s vs %s", id, firstID)
				}
			}
			t.Logf("✓ 所有并发请求返回相同的 TableID")
		}

		if errorCount > 0 {
			t.Logf("警告: %d 个并发请求失败", errorCount)
		}
	})
}

// TestIntegrationLargeBatch 集成测试：大批量数据
func TestIntegrationLargeBatch(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 创建大批量测试数据
	batchSize := 50
	products := make([]Product, batchSize)
	for i := 0; i < batchSize; i++ {
		products[i] = Product{
			ItemID:     fmt.Sprintf("large_batch_test_%04d", i),
			Title:      fmt.Sprintf("大批量测试商品 %d", i+1),
			Price:      fmt.Sprintf("%.2f", float64(i+1)*10),
			PriceNumber: float64(i + 1) * 10,
			WantCnt:    i + 1,
			SellerNick: fmt.Sprintf("卖家 %d", i+1),
			SellerCity: "测试城市",
		}
	}

	today := time.Now()

	resp, err := service.PushProductsToDateTable(today, products)
	if err != nil {
		t.Fatalf("推送大批量数据失败: %v", err)
	}

	if !resp.Success {
		t.Fatalf("推送失败: %s", resp.Message)
	}

	t.Logf("✓ 成功推送 %d 条记录", resp.Data.RecordsCreated)

	if resp.Data.RecordsCreated != batchSize {
		t.Logf("警告: 期望创建 %d 条记录，实际创建 %d 条", batchSize, resp.Data.RecordsCreated)
	}
}

// TestIntegrationDataTypeHandling 集成测试：数据类型处理
func TestIntegrationDataTypeHandling(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	client := NewClient(ClientConfig{
		AppID:     appID,
		AppSecret: appSecret,
	})

	service := NewBitableService(client, BitableConfig{
		AppToken: appToken,
	})

	// 测试各种数据类型
	products := []Product{
		{
			ItemID:     "type_test_text",
			Title:      "文本类型测试",
			Price:      "100.00",
			WantCnt:    0,
			SellerNick: "",
			SellerCity: "",
		},
		{
			ItemID:     "type_test_number",
			Title:      "数字类型测试",
			Price:      "200.50",
			PriceNumber: 200.50,
			WantCnt:    999,
			SellerNick: "数字卖家",
			SellerCity: "数字城市",
		},
		{
			ItemID:     "type_test_special_chars",
			Title:      "特殊字符测试: \"引号\" \t 制表符 \n 换行符",
			Price:      "¥1,234.56",
			WantCnt:    1,
			SellerNick: "特殊<卖家> & 测试",
			SellerCity: "城市/地区\\测试",
		},
	}

	resp, err := service.PushProductsToTodayTable(products)
	if err != nil {
		t.Fatalf("推送不同数据类型失败: %v", err)
	}

	if !resp.Success {
		t.Fatalf("推送失败: %s", resp.Message)
	}

	t.Logf("✓ 成功推送不同数据类型的商品")
}

// TestIntegrationEnvironmentValidation 测试环境变量验证
func TestIntegrationEnvironmentValidation(t *testing.T) {
	skipIfNotIntegration(t)

	appID, appSecret, appToken := getFeishuTestConfig(t)

	// 验证环境变量不为空
	if strings.TrimSpace(appID) == "" {
		t.Error("FEISHU_APP_ID 不能为空")
	}
	if strings.TrimSpace(appSecret) == "" {
		t.Error("FEISHU_APP_SECRET 不能为空")
	}
	if strings.TrimSpace(appToken) == "" {
		t.Error("FEISHU_APP_TOKEN 不能为空")
	}

	t.Logf("✓ 环境变量验证通过")
	t.Logf("  AppID: %s", appID)
	t.Logf("  AppToken: %s", appToken)
}
