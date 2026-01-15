package feishu

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// BitableConfig 多维表格配置
type BitableConfig struct {
	AppToken   string // 应用 token
	TableToken string // 数据表 token
}

// BitableService 多维表格服务
type BitableService struct {
	client FeishuClient
	config BitableConfig
}

// NewBitableService 创建多维表格服务
func NewBitableService(client FeishuClient, config BitableConfig) *BitableService {
	return &BitableService{
		client: client,
		config: config,
	}
}

// PushProducts 推送商品列表
func (s *BitableService) PushProducts(products []Product) (*PushToBitableResponse, error) {
	return s.client.PushToBitable(s.config.AppToken, s.config.TableToken, products)
}

// PushProduct 推送单个商品
func (s *BitableService) PushProduct(product Product) error {
	return s.client.CreateRecord(s.config.AppToken, s.config.TableToken, product)
}

// ParsePrice 从字符串解析价格数值
func ParsePrice(priceStr string) (float64, error) {
	// 移除常见的价格符号和空格
	priceStr = strings.TrimSpace(priceStr)
	priceStr = strings.ReplaceAll(priceStr, "¥", "")
	priceStr = strings.ReplaceAll(priceStr, "￥", "")
	priceStr = strings.ReplaceAll(priceStr, ",", "")
	priceStr = strings.ReplaceAll(priceStr, " ", "")

	// 解析
	return strconv.ParseFloat(priceStr, 64)
}

// FormatPrice 格式化价格字符串
func FormatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}

// ParseTimestamp 解析时间戳
func ParseTimestamp(timestamp interface{}) int64 {
	switch v := timestamp.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		// 尝试解析字符串格式的时间戳
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		// 尝试解析日期时间字符串
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t.UnixMilli()
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

// ConvertFeedItems 转换 FeedItem 到 Product 列表
// 这是一个辅助函数，用于从现有的 mtop.FeedItem 转换
func ConvertFeedItems(items interface{}) []Product {
	products := make([]Product, 0)
	now := time.Now()

	// 这里需要根据实际的 FeedItem 结构进行转换
	// 由于 mtop.FeedItem 是外部包的类型，这里提供一个示例
	_ = now // 避免未使用变量警告

	return products
}

// ProductBuilder 商品构建器
type ProductBuilder struct {
	product Product
}

// NewProductBuilder 创建商品构建器
func NewProductBuilder() *ProductBuilder {
	return &ProductBuilder{
		product: Product{},
	}
}

// WithItemID 设置商品ID
func (b *ProductBuilder) WithItemID(id string) *ProductBuilder {
	b.product.ItemID = id
	return b
}

// WithTitle 设置标题
func (b *ProductBuilder) WithTitle(title string) *ProductBuilder {
	b.product.Title = title
	return b
}

// WithPrice 设置价格
func (b *ProductBuilder) WithPrice(priceStr string) *ProductBuilder {
	b.product.Price = priceStr
	return b
}

// WithOriginalPrice 设置原价
func (b *ProductBuilder) WithOriginalPrice(originalPriceStr string) *ProductBuilder {
	b.product.OriginalPrice = originalPriceStr
	return b
}

// WithWantCount 设置想要人数
func (b *ProductBuilder) WithWantCount(count int) *ProductBuilder {
	b.product.WantCnt = count
	return b
}

// WithPublishTime 设置发布时间（时间戳）
func (b *ProductBuilder) WithPublishTime(publishTime interface{}) *ProductBuilder {
	switch v := publishTime.(type) {
	case string:
		if timestamp := ParseTimestamp(v); timestamp > 0 {
			b.product.PublishTimeMs = timestamp
		}
	case int64:
		b.product.PublishTimeMs = v
	}
	return b
}

// WithSellerInfo 设置卖家信息
func (b *ProductBuilder) WithSellerInfo(nick, city string) *ProductBuilder {
	b.product.SellerNick = nick
	b.product.SellerCity = city
	return b
}

// WithFreeShip 设置包邮
func (b *ProductBuilder) WithFreeShip(freeShip bool) *ProductBuilder {
	if freeShip {
		b.product.FreeShip = "是"
	} else {
		b.product.FreeShip = "否"
	}
	return b
}

// WithTags 设置标签
func (b *ProductBuilder) WithTags(tags string) *ProductBuilder {
	b.product.Tags = tags
	return b
}

// WithURLs 设置 URL
func (b *ProductBuilder) WithURLs(coverURL, detailURL string) *ProductBuilder {
	b.product.CoverURL = coverURL
	b.product.DetailURL = detailURL
	return b
}

// WithExposureHeat 设置曝光热度
func (b *ProductBuilder) WithExposureHeat(heat int) *ProductBuilder {
	b.product.ExposureHeat = heat
	return b
}

// WithCaptureTime 设置采集时间（时间戳）
func (b *ProductBuilder) WithCaptureTime(captureTime time.Time) *ProductBuilder {
	b.product.CaptureTimeMs = captureTime.UnixMilli()
	return b
}

// Build 构建商品
func (b *ProductBuilder) Build() Product {
	// 如果没有设置采集时间，使用当前时间
	if b.product.CaptureTimeMs == 0 {
		b.WithCaptureTime(time.Now())
	}
	return b.product
}

// GetOrCreateTableByDate 根据日期获取或创建表格
// 返回表格ID和是否为新创建的表格
func (s *BitableService) GetOrCreateTableByDate(date time.Time) (tableID string, created bool, err error) {
	tableName := date.Format("2006-01-02")

	// 获取所有表格
	tables, err := s.client.GetBitableTableInfos(s.config.AppToken)
	if err != nil {
		return "", false, fmt.Errorf("获取表格列表失败: %w", err)
	}

	// 打印所有现有表格名（调试用）
	var existingNames []string
	for _, t := range tables {
		existingNames = append(existingNames, t.Name)
		// 查找是否存在同名表格（精确匹配）
		if t.Name == tableName {
			fmt.Printf("[DEBUG] 找到已存在表格: %s (ID: %s)\n", tableName, t.TableID)
			return t.TableID, false, nil
		}
	}
	fmt.Printf("[DEBUG] 现有表格: %v, 查找: %s\n", existingNames, tableName)

	// 创建新表格，包含所有需要的字段
	fields := s.buildFieldCreates()
	fmt.Printf("[DEBUG] 开始创建表格: %s\n", tableName)
	tableInfo, err := s.client.CreateTable(s.config.AppToken, tableName, fields)
	if err != nil {
		// 检查是否是表格名重复错误（并发情况下可能被其他进程创建）
		if strings.Contains(err.Error(), "Duplicated") || strings.Contains(err.Error(), "重复") {
			fmt.Printf("[DEBUG] 表格名重复，重试查询: %s\n", tableName)
			// 等待一小段时间让数据同步
			time.Sleep(500 * time.Millisecond)

			// 再次查询获取表格ID
			tables, retryErr := s.client.GetBitableTableInfos(s.config.AppToken)
			if retryErr != nil {
				return "", false, fmt.Errorf("创建表格失败且重试查询失败: %w, 原错误: %v", retryErr, err)
			}

			// 打印重试时的表格列表
			var retryNames []string
			for _, t := range tables {
				retryNames = append(retryNames, t.Name)
				if t.Name == tableName {
					fmt.Printf("[DEBUG] 重试找到表格: %s (ID: %s)\n", tableName, t.TableID)
					return t.TableID, false, nil
				}
			}
			return "", false, fmt.Errorf("表格名重复但未找到表格。查找: %s, 现有表格: %v, 错误: %v", tableName, retryNames, err)
		}
		return "", false, fmt.Errorf("创建表格失败: %w", err)
	}

	fmt.Printf("[DEBUG] 表格创建成功: %s (ID: %s)\n", tableName, tableInfo.TableID)
	return tableInfo.TableID, true, nil
}

// buildFieldCreates 根据 ProductFields 构建字段创建列表
// 使用英文Key作为field_name，中文Label作为options.label
func (s *BitableService) buildFieldCreates() []FieldCreate {
	fields := make([]FieldCreate, 0, len(ProductFields))
	for _, pf := range ProductFields {
		field := FieldCreate{
			FieldName: pf.Key, // 使用英文Key作为field_name
			Type:      int(pf.Schema.Type),
			Options: map[string]interface{}{
				"label": pf.Schema.Label, // 设置中文显示名称
			},
		}
		fields = append(fields, field)
	}
	return fields
}

// EnsureTableFields 确保表格包含所有需要的字段，如果不存在则创建
func (s *BitableService) EnsureTableFields(tableID string) error {
	// 获取现有字段
	existingFields, err := s.client.GetTableFields(s.config.AppToken, tableID)
	if err != nil {
		return fmt.Errorf("获取现有字段失败: %w", err)
	}

	fmt.Printf("[DEBUG] 现有字段数: %d\n", len(existingFields))

	// 统计字段创建结果
	successCount := 0
	failCount := 0
	skipCount := 0
	var failedFields []string

	// 检查并创建缺失的字段
	for _, pf := range ProductFields {
		// 使用 Key（英文标识符）作为字段名
		fieldName := pf.Key
		if _, exists := existingFields[fieldName]; !exists {
			// 字段不存在，创建它
			field := FieldCreate{
				FieldName: fieldName,
				Type:      int(pf.Schema.Type),
				Options: map[string]interface{}{
					"label": pf.Schema.Label, // 设置中文名称
				},
			}

			fmt.Printf("[创建字段] %s (fieldName=%s, type=%d)\n", pf.Schema.Label, fieldName, pf.Schema.Type)

			fieldID, err := s.client.CreateField(s.config.AppToken, tableID, field)
			if err != nil {
				// 打印详细错误用于调试
				fmt.Printf("[ERROR] 创建字段失败!\n")
				fmt.Printf("  - 中文名称: %s\n", pf.Schema.Label)
				fmt.Printf("  - 字段标识: %s\n", fieldName)
				fmt.Printf("  - 字段类型: %d\n", pf.Schema.Type)
				fmt.Printf("  - 错误信息: %v\n", err)
				failCount++
				failedFields = append(failedFields, pf.Schema.Label)
				// 继续创建其他字段，不直接返回错误
				continue
			}

			successCount++
			fmt.Printf("[SUCCESS] 创建字段成功: %s (fieldID=%s)\n", pf.Schema.Label, fieldID)
		} else {
			skipCount++
		}
	}

	// 打印统计信息
	fmt.Printf("\n[字段创建统计]\n")
	fmt.Printf("  成功: %d\n", successCount)
	fmt.Printf("  跳过(已存在): %d\n", skipCount)
	fmt.Printf("  失败: %d\n", failCount)

	if failCount > 0 {
		fmt.Printf("\n[失败字段列表]\n")
		for i, name := range failedFields {
			fmt.Printf("  %d. %s\n", i+1, name)
		}
		return fmt.Errorf("有 %d 个字段创建失败", failCount)
	}

	return nil
}

// PushProductsToDateTable 推送商品数据到指定日期的表格
// 如果表格不存在则创建，如果字段不存在则自动创建
func (s *BitableService) PushProductsToDateTable(date time.Time, products []Product) (*PushToBitableResponse, error) {
	// 获取或创建表格
	tableID, created, err := s.GetOrCreateTableByDate(date)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[DEBUG] 推送数据到表格: tableID=%s, 商品数=%d\n", tableID, len(products))

	// 如果是新创建的表格，字段已在创建时定义，无需再检查
	// 如果是已存在的表格，确保所有字段都存在
	if !created {
		fmt.Printf("[DEBUG] 表格已存在，检查字段...\n")
		if err := s.EnsureTableFields(tableID); err != nil {
			return nil, fmt.Errorf("确保字段存在失败: %w", err)
		}

		// 去重：过滤掉已存在的商品
		products, err = s.DeduplicateProducts(tableID, products)
		if err != nil {
			return nil, fmt.Errorf("去重失败: %w", err)
		}
		fmt.Printf("[DEBUG] 去重后商品数: %d\n", len(products))
	}

	// 如果去重后没有数据需要推送，直接返回
	if len(products) == 0 {
		fmt.Println("[去重] 所有商品都已存在，无需推送")
		resp := &PushToBitableResponse{
			Success: true,
			Message: "所有商品都已存在",
		}
		resp.Data.RecordsCreated = 0
		return resp, nil
	}

	// 推送数据
	resp, err := s.client.PushToBitable(s.config.AppToken, tableID, products)
	if err != nil {
		return nil, fmt.Errorf("推送数据失败: %w", err)
	}

	return resp, nil
}

// PushProductsToTodayTable 推送商品数据到今天的表格
func (s *BitableService) PushProductsToTodayTable(products []Product) (*PushToBitableResponse, error) {
	return s.PushProductsToDateTable(time.Now(), products)
}

// GetTableByDate 根据日期获取表格ID，如果不存在则返回空字符串
func (s *BitableService) GetTableByDate(date time.Time) (string, error) {
	tableName := date.Format("2006-01-02")

	tables, err := s.client.GetBitableTableInfos(s.config.AppToken)
	if err != nil {
		return "", fmt.Errorf("获取表格列表失败: %w", err)
	}

	for _, table := range tables {
		if table.Name == tableName {
			return table.TableID, nil
		}
	}

	return "", nil
}

// TableExists 检查指定日期的表格是否存在
func (s *BitableService) TableExists(date time.Time) (bool, error) {
	tableID, err := s.GetTableByDate(date)
	if err != nil {
		return false, err
	}
	return tableID != "", nil
}

// ProductKey 商品的唯一标识（用于去重）
type ProductKey struct {
	ItemID  string // 商品ID
	Price   string // 价格
	WantCnt int    // 想要人数
}

// buildProductKey 构建商品唯一标识
func (s *BitableService) buildProductKey(itemID, price string, wantCnt int) ProductKey {
	return ProductKey{
		ItemID:  itemID,
		Price:   price,
		WantCnt: wantCnt,
	}
}

// extractProductKeyFromRecord 从飞书记录中提取商品唯一标识
func (s *BitableService) extractProductKeyFromRecord(record map[string]interface{}, fieldNameMapping map[string]string) (ProductKey, error) {
	// 获取商品ID
	itemID := ""
	if v, ok := record[fieldNameMapping["itemId"]]; ok {
		if str, ok := v.(string); ok {
			itemID = str
		}
	}

	// 获取价格
	price := ""
	if v, ok := record[fieldNameMapping["price"]]; ok {
		if str, ok := v.(string); ok {
			price = str
		}
	}

	// 获取想要人数
	wantCnt := 0
	if v, ok := record[fieldNameMapping["wantCnt"]]; ok {
		switch val := v.(type) {
		case float64:
			wantCnt = int(val)
		case int:
			wantCnt = val
		}
	}

	return ProductKey{
		ItemID:  itemID,
		Price:   price,
		WantCnt: wantCnt,
	}, nil
}

// DeduplicateProducts 对商品列表进行去重
// 改进的去重逻辑：逐个根据商品ID查询，而不是一次性获取所有记录
// 去重标准：tableId + itemId + 价格 + 想要人数
func (s *BitableService) DeduplicateProducts(tableID string, products []Product) ([]Product, error) {
	fieldNameMapping := GetFieldNameMapping()
	itemIDFieldName := fieldNameMapping["itemId"]

	filteredProducts := make([]Product, 0, len(products))
	duplicateCount := 0

	// 按商品ID分组，同时去除本地重复
	productsByItemID := make(map[string][]Product)
	for _, product := range products {
		productsByItemID[product.ItemID] = append(productsByItemID[product.ItemID], product)
	}

	// 对每个商品ID进行查询
	for itemID, itemProducts := range productsByItemID {
		// 构建过滤条件：根据商品ID查询
		filter := FilterInfo{
			Conjunction: "and",
			Conditions: []Condition{
				{
					FieldName: itemIDFieldName,
					Operator:  "is",
					Value:     []string{itemID},
				},
			},
		}

		// 查询该商品ID的现有记录
		existingRecords, err := s.client.SearchRecords(s.config.AppToken, tableID, filter)
		if err != nil {
			return nil, fmt.Errorf("查询商品ID=%s的记录失败: %w", itemID, err)
		}

		// 构建已存在记录的 key 集合（用于同一 itemID 下的去重）
		existingKeys := make(map[ProductKey]bool)
		for _, record := range existingRecords {
			key, err := s.extractProductKeyFromRecord(record, fieldNameMapping)
			if err != nil {
				continue
			}
			existingKeys[key] = true
		}

		// 检查每个商品是否重复
		for _, product := range itemProducts {
			key := s.buildProductKey(product.ItemID, product.Price, product.WantCnt)
			if !existingKeys[key] {
				filteredProducts = append(filteredProducts, product)
				existingKeys[key] = true // 标记为已存在，防止本次推送中重复
			} else {
				duplicateCount++
			}
		}
	}

	if duplicateCount > 0 {
		fmt.Printf("[去重] 过滤掉 %d 条重复记录\n", duplicateCount)
	}

	return filteredProducts, nil
}
