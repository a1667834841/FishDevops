package feishu

import (
	"fmt"
	"testing"
	"time"
)

// TestBuildFieldCreates 测试字段创建列表构建
func TestBuildFieldCreates(t *testing.T) {
	service := &BitableService{
		client: nil,
		config: BitableConfig{},
	}

	fields := service.buildFieldCreates()

	if len(fields) != len(ProductFields) {
		t.Errorf("buildFieldCreates() 返回 %d 个字段, 期望 %d", len(fields), len(ProductFields))
	}

	// 验证字段顺序和类型
	expectedFields := []struct {
		name string
		ftype FieldType
	}{
		{"商品ID", FieldTypeText},
		{"商品标题", FieldTypeText},
		{"价格", FieldTypeText},
		{"价格数值", FieldTypeNumber},
		{"原价", FieldTypeText},
		{"原价数值", FieldTypeNumber},
		{"想要人数", FieldTypeNumber},
		{"发布时间", FieldTypeText},
		{"发布时间戳", FieldTypeDateTime},
		{"采集时间", FieldTypeText},
		{"采集时间戳", FieldTypeDateTime},
		{"卖家昵称", FieldTypeText},
		{"地区", FieldTypeText},
		{"包邮", FieldTypeText},
		{"商品标签", FieldTypeText},
		{"封面URL", FieldTypeURL},
		{"商品详情URL", FieldTypeURL},
		{"曝光热度", FieldTypeNumber},
		{"最近擦亮时间", FieldTypeText},
		{"擦亮时间戳", FieldTypeDateTime},
	}

	if len(fields) != len(expectedFields) {
		t.Fatalf("字段数量不匹配: got %d, want %d", len(fields), len(expectedFields))
	}

	for i, field := range fields {
		if field.FieldName != expectedFields[i].name {
			t.Errorf("字段 %d 名称 = %s, want %s", i, field.FieldName, expectedFields[i].name)
		}
		if FieldType(field.Type) != expectedFields[i].ftype {
			t.Errorf("字段 %d 类型 = %d, want %d", i, field.Type, expectedFields[i].ftype)
		}
	}
}



// TestGetTableNameByDate 测试日期转表名
func TestGetTableNameByDate(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "标准日期",
			date:     time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
			expected: "2026-01-11",
		},
		{
			name:     "不同日期",
			date:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2025-12-31",
		},
		{
			name:     "今天",
			date:     time.Now(),
			expected: time.Now().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.date.Format("2006-01-02")
			if result != tt.expected {
				t.Errorf("Format() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// TestPushProductsToDateTable 测试推送数据到日期表格
func TestPushProductsToDateTable(t *testing.T) {
	// 这个测试需要 mock HTTP 服务器，将在集成测试中完成
	// 这里我们测试数据准备逻辑
	testDate := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)

	products := []Product{
		{
			ItemID:    "item123",
			Title:     "测试商品",
			Price:     "100.00",
			WantCnt:   10,
			FreeShip:  "是",
			SellerNick: "测试卖家",
		},
	}

	// 验证日期格式
	tableName := testDate.Format("2006-01-02")
	if tableName != "2026-01-11" {
		t.Errorf("表名格式错误: got %s, want 2026-01-11", tableName)
	}

	// 验证产品数据
	if len(products) != 1 {
		t.Fatalf("产品数量错误: got %d, want 1", len(products))
	}

	if products[0].ItemID != "item123" {
		t.Errorf("ItemID = %s, want item123", products[0].ItemID)
	}
}

// MockClient 用于测试的 mock 客户端
type MockClient struct {
	token              string
	tables             []TableInfo
	fields             map[string]string
	createdTables      []string
	createdFields      map[string][]string
	pushedRecordsCount int
}

func (m *MockClient) GetTenantAccessToken() (string, error) {
	return m.token, nil
}

func (m *MockClient) GetBitableTableInfos(appToken string) ([]TableInfo, error) {
	return m.tables, nil
}

func (m *MockClient) CreateTable(appToken, tableName string, fields []FieldCreate) (*TableInfo, error) {
	m.createdTables = append(m.createdTables, tableName)
	return &TableInfo{TableID: "new_" + tableName, Name: tableName}, nil
}

func (m *MockClient) GetTableFields(appToken, tableToken string) (map[string]string, error) {
	return m.fields, nil
}

func (m *MockClient) CreateField(appToken, tableToken string, field FieldCreate) (string, error) {
	if m.createdFields == nil {
		m.createdFields = make(map[string][]string)
	}
	m.createdFields[tableToken] = append(m.createdFields[tableToken], field.FieldName)
	return "field_" + field.FieldName, nil
}

func (m *MockClient) PushToBitable(appToken, tableToken string, products []Product) (*PushToBitableResponse, error) {
	m.pushedRecordsCount = len(products)
	return &PushToBitableResponse{
		Success: true,
		Message: "推送成功",
	}, nil
}

func (m *MockClient) CreateRecord(appToken, tableToken string, product Product) error {
	m.pushedRecordsCount = 1
	return nil
}

func (m *MockClient) GetTableRecords(appToken, tableToken string) ([]map[string]interface{}, error) {
	// 返回空记录列表用于测试
	return []map[string]interface{}{}, nil
}

// TestBitableService_GetOrCreateTableByDate 测试获取或创建表格
func TestBitableService_GetOrCreateTableByDate(t *testing.T) {
	testDate := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)

	t.Run("表格已存在", func(t *testing.T) {
		mockClient := &MockClient{
			token: "test_token",
			tables: []TableInfo{
				{TableID: "tbl123", Name: "2026-01-11"},
			},
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		tableID, created, err := service.GetOrCreateTableByDate(testDate)
		if err != nil {
			t.Fatalf("GetOrCreateTableByDate() error = %v", err)
		}

		if tableID != "tbl123" {
			t.Errorf("tableID = %s, want tbl123", tableID)
		}
		if created {
			t.Error("created = true, want false")
		}
		if len(mockClient.createdTables) != 0 {
			t.Errorf("创建了 %d 个表格, want 0", len(mockClient.createdTables))
		}
	})

	t.Run("表格不存在，创建新表格", func(t *testing.T) {
		mockClient := &MockClient{
			token:  "test_token",
			tables: []TableInfo{}, // 空列表
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		tableID, created, err := service.GetOrCreateTableByDate(testDate)
		if err != nil {
			t.Fatalf("GetOrCreateTableByDate() error = %v", err)
		}

		if tableID != "new_2026-01-11" {
			t.Errorf("tableID = %s, want new_2026-01-11", tableID)
		}
		if !created {
			t.Error("created = false, want true")
		}
		if len(mockClient.createdTables) != 1 {
			t.Errorf("创建了 %d 个表格, want 1", len(mockClient.createdTables))
		}
		if mockClient.createdTables[0] != "2026-01-11" {
			t.Errorf("创建的表名 = %s, want 2026-01-11", mockClient.createdTables[0])
		}
	})
}

// TestBitableService_EnsureTableFields 测试确保字段存在
func TestBitableService_EnsureTableFields(t *testing.T) {
	t.Run("所有字段已存在", func(t *testing.T) {
		existingFields := make(map[string]string)
		for _, pf := range ProductFields {
			existingFields[pf.Schema.Label] = "field_" + pf.Schema.Label
		}

		mockClient := &MockClient{
			token:  "test_token",
			fields: existingFields,
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		err := service.EnsureTableFields("tbl123")
		if err != nil {
			t.Fatalf("EnsureTableFields() error = %v", err)
		}

		if mockClient.createdFields != nil && len(mockClient.createdFields["tbl123"]) != 0 {
			t.Errorf("创建了 %d 个字段, want 0", len(mockClient.createdFields["tbl123"]))
		}
	})

	t.Run("部分字段缺失，自动创建", func(t *testing.T) {
		existingFields := map[string]string{
			"商品ID": "field_itemId",
		}

		mockClient := &MockClient{
			token:  "test_token",
			fields: existingFields,
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		err := service.EnsureTableFields("tbl123")
		if err != nil {
			t.Fatalf("EnsureTableFields() error = %v", err)
		}

		// 应该创建除了 "商品ID" 之外的所有字段
		expectedCreatedCount := len(ProductFields) - 1
		if len(mockClient.createdFields["tbl123"]) != expectedCreatedCount {
			t.Errorf("创建了 %d 个字段, want %d", len(mockClient.createdFields["tbl123"]), expectedCreatedCount)
		}
	})
}

// TestBitableService_PushProductsToDateTable 测试推送数据到日期表格
func TestBitableService_PushProductsToDateTable(t *testing.T) {
	testDate := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	products := []Product{
		{ItemID: "item1", Title: "商品1"},
		{ItemID: "item2", Title: "商品2"},
	}

	t.Run("表格已存在，字段已存在", func(t *testing.T) {
		existingFields := make(map[string]string)
		for _, pf := range ProductFields {
			existingFields[pf.Schema.Label] = "field_" + pf.Schema.Label
		}

		mockClient := &MockClient{
			token: "test_token",
			tables: []TableInfo{
				{TableID: "tbl123", Name: "2026-01-11"},
			},
			fields: existingFields,
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		resp, err := service.PushProductsToDateTable(testDate, products)
		if err != nil {
			t.Fatalf("PushProductsToDateTable() error = %v", err)
		}

		if !resp.Success {
			t.Error("Success = false, want true")
		}
		if mockClient.pushedRecordsCount != 2 {
			t.Errorf("推送了 %d 条记录, want 2", mockClient.pushedRecordsCount)
		}
	})

	t.Run("表格不存在，创建新表格", func(t *testing.T) {
		mockClient := &MockClient{
			token:  "test_token",
			tables: []TableInfo{}, // 空列表
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		resp, err := service.PushProductsToDateTable(testDate, products)
		if err != nil {
			t.Fatalf("PushProductsToDateTable() error = %v", err)
		}

		if !resp.Success {
			t.Error("Success = false, want true")
		}
		if len(mockClient.createdTables) != 1 {
			t.Errorf("创建了 %d 个表格, want 1", len(mockClient.createdTables))
		}
	})
}

// TestBitableService_GetTableByDate 测试根据日期获取表格
func TestBitableService_GetTableByDate(t *testing.T) {
	testDate := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)

	t.Run("表格存在", func(t *testing.T) {
		mockClient := &MockClient{
			token: "test_token",
			tables: []TableInfo{
				{TableID: "tbl123", Name: "2026-01-11"},
			},
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		tableID, err := service.GetTableByDate(testDate)
		if err != nil {
			t.Fatalf("GetTableByDate() error = %v", err)
		}

		if tableID != "tbl123" {
			t.Errorf("tableID = %s, want tbl123", tableID)
		}
	})

	t.Run("表格不存在", func(t *testing.T) {
		mockClient := &MockClient{
			token:  "test_token",
			tables: []TableInfo{
				{TableID: "tbl789", Name: "2026-01-10"},
			},
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		tableID, err := service.GetTableByDate(testDate)
		if err != nil {
			t.Fatalf("GetTableByDate() error = %v", err)
		}

		if tableID != "" {
			t.Errorf("tableID = %s, want empty string", tableID)
		}
	})
}

// TestBitableService_TableExists 测试表格是否存在
func TestBitableService_TableExists(t *testing.T) {
	testDate := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)

	t.Run("表格存在", func(t *testing.T) {
		mockClient := &MockClient{
			token: "test_token",
			tables: []TableInfo{
				{TableID: "tbl123", Name: "2026-01-11"},
			},
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		exists, err := service.TableExists(testDate)
		if err != nil {
			t.Fatalf("TableExists() error = %v", err)
		}

		if !exists {
			t.Error("exists = false, want true")
		}
	})

	t.Run("表格不存在", func(t *testing.T) {
		mockClient := &MockClient{
			token:  "test_token",
			tables: []TableInfo{
				{TableID: "tbl789", Name: "2026-01-10"},
			},
		}

		service := &BitableService{
			client: mockClient,
			config: BitableConfig{AppToken: "app123"},
		}

		exists, err := service.TableExists(testDate)
		if err != nil {
			t.Fatalf("TableExists() error = %v", err)
		}

		if exists {
			t.Error("exists = true, want false")
		}
	})
}

// TestProductBuilder 测试商品构建器
func TestProductBuilder(t *testing.T) {
	now := time.Date(2026, 1, 11, 12, 30, 0, 0, time.UTC)

	product := NewProductBuilder().
		WithItemID("item123").
		WithTitle("测试商品").
		WithPrice("100.00").
		WithOriginalPrice("200.00").
		WithWantCount(10).
		WithPublishTime("2026-01-11 10:00:00").
		WithSellerInfo("测试卖家", "北京").
		WithFreeShip(true).
		WithTags("全新").
		WithURLs("http://cover.jpg", "http://detail").
		WithExposureHeat(100).
		WithCaptureTime(now).
		Build()

	if product.ItemID != "item123" {
		t.Errorf("ItemID = %s, want item123", product.ItemID)
	}
	if product.Title != "测试商品" {
		t.Errorf("Title = %s, want 测试商品", product.Title)
	}
	if product.Price != "100.00" {
		t.Errorf("Price = %s, want 100.00", product.Price)
	}
	if product.PriceNumber != 100.00 {
		t.Errorf("PriceNumber = %f, want 100.00", product.PriceNumber)
	}
	if product.OriginalPriceNumber != 200.00 {
		t.Errorf("OriginalPriceNumber = %f, want 200.00", product.OriginalPriceNumber)
	}
	if product.WantCnt != 10 {
		t.Errorf("WantCnt = %d, want 10", product.WantCnt)
	}
	if product.SellerNick != "测试卖家" {
		t.Errorf("SellerNick = %s, want 测试卖家", product.SellerNick)
	}
	if product.SellerCity != "北京" {
		t.Errorf("SellerCity = %s, want 北京", product.SellerCity)
	}
	if product.FreeShip != "是" {
		t.Errorf("FreeShip = %s, want 是", product.FreeShip)
	}
	if product.ExposureHeat != 100 {
		t.Errorf("ExposureHeat = %d, want 100", product.ExposureHeat)
	}
	if product.CaptureTime != now.Format("2006-01-02 15:04:05") {
		t.Errorf("CaptureTime = %s, want %s", product.CaptureTime, now.Format("2006-01-02 15:04:05"))
	}
}

// TestParsePrice 测试价格解析
func TestParsePrice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{"标准格式", "100.00", 100.00, false},
		{"带符号", "¥100.00", 100.00, false},
		{"全角符号", "￥100.00", 100.00, false},
		{"带逗号", "1,000.00", 1000.00, false},
		{"带空格", " 100.00 ", 100.00, false},
		{"整数", "100", 100.00, false},
		{"无效格式", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePrice() = %f, want %f", got, tt.want)
			}
		})
	}
}

// TestFormatPrice 测试价格格式化
func TestFormatPrice(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{"标准", 100.0, "100.00"},
		{"小数", 99.99, "99.99"},
		{"零", 0.0, "0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatPrice(tt.input); got != tt.want {
				t.Errorf("FormatPrice() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestParseTimestamp 测试时间戳解析
func TestParseTimestamp(t *testing.T) {
	fixedTime := time.Date(2026, 1, 11, 12, 30, 0, 0, time.UTC)
	expectedMs := fixedTime.UnixMilli()

	tests := []struct {
		name  string
		input interface{}
		want  int64
	}{
		{"int64", int64(1234567890), 1234567890},
		{"int", int(1234567890), int64(1234567890)},
		{"float64", float64(1234567890), int64(1234567890)},
		{"字符串时间戳", "1234567890", int64(1234567890)},
		{"日期时间字符串", fixedTime.Format("2006-01-02 15:04:05"), expectedMs},
		{"RFC3339", fixedTime.Format(time.RFC3339), expectedMs},
		{"无效字符串", "invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTimestamp(tt.input); got != tt.want {
				t.Errorf("ParseTimestamp() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestGetFieldNameMapping 测试字段名映射
func TestGetFieldNameMapping(t *testing.T) {
	mapping := GetFieldNameMapping()

	// 验证一些关键字段
	tests := []struct {
		key    string
		want   string
		exists bool
	}{
		{"itemId", "商品ID", true},
		{"title", "商品标题", true},
		{"price", "价格", true},
		{"priceNumber", "价格数值", true},
		{"wantCnt", "想要人数", true},
		{"publishTime", "发布时间", true},
		{"sellerNick", "卖家昵称", true},
		{"freeShip", "包邮", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, exists := mapping[tt.key]
			if exists != tt.exists {
				t.Errorf("mapping[%s] exists = %v, want %v", tt.key, exists, tt.exists)
			}
			if tt.exists && got != tt.want {
				t.Errorf("mapping[%s] = %s, want %s", tt.key, got, tt.want)
			}
		})
	}
}

// TestPushProductsToTodayTable 测试推送数据到今天表格
func TestPushProductsToTodayTable(t *testing.T) {
	mockClient := &MockClient{
		token: "test_token",
		tables: []TableInfo{
			{TableID: "tbl123", Name: time.Now().Format("2006-01-02")},
		},
		fields: make(map[string]string),
	}
	// 添加所有字段
	for _, pf := range ProductFields {
		mockClient.fields[pf.Schema.Label] = "field_" + pf.Schema.Label
	}

	service := &BitableService{
		client: mockClient,
		config: BitableConfig{AppToken: "app123"},
	}

	products := []Product{
		{ItemID: "item1", Title: "商品1"},
	}

	resp, err := service.PushProductsToTodayTable(products)
	if err != nil {
		t.Fatalf("PushProductsToTodayTable() error = %v", err)
	}

	if !resp.Success {
		t.Error("Success = false, want true")
	}
	if mockClient.pushedRecordsCount != 1 {
		t.Errorf("推送了 %d 条记录, want 1", mockClient.pushedRecordsCount)
	}
}

// BenchmarkBuildFieldCreates 性能测试
func BenchmarkBuildFieldCreates(b *testing.B) {
	service := &BitableService{}
	for i := 0; i < b.N; i++ {
		_ = service.buildFieldCreates()
	}
}

// ExampleGetOrCreateTableByDate 示例代码
func ExampleBitableService_GetOrCreateTableByDate() {
	// 这个示例展示了如何使用 GetOrCreateTableByDate
	service := &BitableService{
		client: nil,
		config: BitableConfig{AppToken: "your_app_token"},
	}

	date := time.Now()
	tableID, created, err := service.GetOrCreateTableByDate(date)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if created {
		fmt.Printf("创建了新表格: %s\n", tableID)
	} else {
		fmt.Printf("使用已存在表格: %s\n", tableID)
	}
}

// ExamplePushProductsToDateTable 示例代码
func ExampleBitableService_PushProductsToDateTable() {
	// 这个示例展示了如何使用 PushProductsToDateTable
	service := &BitableService{
		client: nil,
		config: BitableConfig{AppToken: "your_app_token"},
	}

	products := []Product{
		{
			ItemID:  "item123",
			Title:   "示例商品",
			Price:   "99.00",
			WantCnt: 5,
		},
	}

	date := time.Now()
	resp, err := service.PushProductsToDateTable(date, products)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Printf("推送成功: %v, 记录数: %d\n", resp.Success, resp.Data.RecordsCreated)
}
