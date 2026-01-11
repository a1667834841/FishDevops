package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// 飞书 API 基础 URL
	baseURL = "https://open.feishu.cn"
	// API 路径
	authPath       = "/open-apis/auth/v3/tenant_access_token/internal"
	bitablePath    = "/open-apis/bitable/v1/apps/%s/tables/%s/records"
	bitableBatchPath = "/open-apis/bitable/v1/apps/%s/tables/%s/records/batch_create"
)

// Client 飞书客户端
type Client struct {
	appID     string
	appSecret string
	httpCli   *http.Client
	token     string
	tokenExp  int64
}

// FeishuClient 飞书客户端接口，用于测试 mock
type FeishuClient interface {
	GetTenantAccessToken() (string, error)
	GetBitableTableInfos(appToken string) ([]TableInfo, error)
	CreateTable(appToken, tableName string, fields []FieldCreate) (*TableInfo, error)
	GetTableFields(appToken, tableToken string) (map[string]string, error)
	CreateField(appToken, tableToken string, field FieldCreate) (string, error)
	PushToBitable(appToken, tableToken string, products []Product) (*PushToBitableResponse, error)
	CreateRecord(appToken, tableToken string, product Product) error
	GetTableRecords(appToken, tableToken string) ([]map[string]interface{}, error)
}

// ClientConfig 客户端配置
type ClientConfig struct {
	AppID     string
	AppSecret string
}

// NewClient 创建飞书客户端
func NewClient(config ClientConfig) *Client {
	return &Client{
		appID:     config.AppID,
		appSecret: config.AppSecret,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TenantAccessTokenResponse 租户访问令牌响应
type TenantAccessTokenResponse struct {
	Code               int    `json:"code"`
	TenantAccessToken  string `json:"tenant_access_token"`
	Expire             int    `json:"expire"`
	Msg                string `json:"msg"`
}

// GetTenantAccessToken 获取租户访问令牌
func (c *Client) GetTenantAccessToken() (string, error) {
	// 如果 token 未过期，直接返回
	if c.token != "" && c.tokenExp > time.Now().Unix() {
		return c.token, nil
	}

	// 构建请求
	reqBody := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %w", err)
	}

	// 发送请求
	resp, err := c.httpCli.Post(baseURL+authPath, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var tokenResp TenantAccessTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("获取 token 失败: %s", tokenResp.Msg)
	}

	// 缓存 token
	c.token = tokenResp.TenantAccessToken
	c.tokenExp = time.Now().Unix() + int64(tokenResp.Expire) - 300 // 提前 5 分钟过期

	return c.token, nil
}

// Record 单条记录
type Record struct {
	Fields map[string]interface{} `json:"fields"`
}

// CreateRecordsResponse 创建记录响应
type CreateRecordsResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Records []struct {
			RecordID string `json:"record_id"`
			Fields   map[string]interface{} `json:"fields"`
		} `json:"records"`
	} `json:"data"`
}

// PushToBitable 推送数据到飞书多维表格
func (c *Client) PushToBitable(appToken, tableToken string, products []Product) (*PushToBitableResponse, error) {
	// 调试日志（脱敏）
	maskToken := func(s string) string {
		if len(s) <= 8 {
			return s
		}
		return s[:4] + "..." + s[len(s)-4:]
	}
	fmt.Printf("[DEBUG] appToken=%s, tableToken=%s\n", maskToken(appToken), maskToken(tableToken))

	// 获取访问令牌
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	// 构建记录
	records := make([]Record, 0, len(products))
	fieldNameMapping := GetFieldNameMapping()

	for i, product := range products {
		fields := make(map[string]interface{})

		// 辅助函数：只在非空时添加字段
		addField := func(key, value string) {
			if value != "" {
				if fieldName, ok := fieldNameMapping[key]; ok {
					fields[fieldName] = value
				}
			}
		}

		// 商品ID
		addField("itemId", product.ItemID)

		// 商品标题
		addField("title", product.Title)

		// 价格
		addField("price", product.Price)

		// 价格数值
		if fieldName, ok := fieldNameMapping["priceNumber"]; ok {
			if product.PriceNumber > 0 {
				fields[fieldName] = product.PriceNumber
			}
		}

		// 原价
		addField("originalPrice", product.OriginalPrice)

		// 原价数值
		if fieldName, ok := fieldNameMapping["originalPriceNumber"]; ok {
			if product.OriginalPriceNumber > 0 {
				fields[fieldName] = product.OriginalPriceNumber
			}
		}

		// 想要人数
		if fieldName, ok := fieldNameMapping["wantCnt"]; ok {
			fields[fieldName] = product.WantCnt
		}

		// 发布时间
		addField("publishTime", product.PublishTime)

		// 发布时间戳
		if fieldName, ok := fieldNameMapping["publishTimeMs"]; ok {
			if product.PublishTimeMs > 0 {
				fields[fieldName] = product.PublishTimeMs
			}
		}

		// 采集时间
		addField("captureTime", product.CaptureTime)

		// 采集时间戳
		if fieldName, ok := fieldNameMapping["captureTimeMs"]; ok {
			if product.CaptureTimeMs > 0 {
				fields[fieldName] = product.CaptureTimeMs
			}
		}

		// 卖家昵称
		addField("sellerNick", product.SellerNick)

		// 地区
		addField("sellerCity", product.SellerCity)

		// 包邮
		addField("freeShip", product.FreeShip)

		// 商品标签
		addField("tags", product.Tags)

		// 封面URL - 飞书URL字段格式: {"link": "url"}
		if fieldName, ok := fieldNameMapping["coverUrl"]; ok {
			if product.CoverURL != "" {
				fields[fieldName] = map[string]string{"link": product.CoverURL}
			}
		}

		// 商品详情URL - 飞书URL字段格式: {"link": "url"}
		if fieldName, ok := fieldNameMapping["detailUrl"]; ok {
			if product.DetailURL != "" {
				fields[fieldName] = map[string]string{"link": product.DetailURL}
			}
		}

		// 曝光热度
		if fieldName, ok := fieldNameMapping["exposureHeat"]; ok {
			if product.ExposureHeat > 0 {
				fields[fieldName] = product.ExposureHeat
			}
		}

		// 擦亮时间
		addField("proPolishTime", product.ProPolishTime)

		// 擦亮时间戳
		if fieldName, ok := fieldNameMapping["proPolishTimeMs"]; ok {
			if product.ProPolishTimeMs > 0 {
				fields[fieldName] = product.ProPolishTimeMs
			}
		}

		// 调试日志：打印第一个商品的详细信息
		if i == 0 {
			fmt.Printf("[DEBUG] 商品数据: ID=%s, Title=%s\n", product.ItemID, product.Title)
			fmt.Printf("[DEBUG] 字段数据 (数量=%d):\n", len(fields))
			for k, v := range fields {
				fmt.Printf("  %s: %v (type: %T)\n", k, v, v)
			}
		}

		records = append(records, Record{Fields: fields})
	}

	// 构建批量创建请求
	reqBody := map[string]interface{}{
		"records": records,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	// 打印请求体（调试用）
	fmt.Printf("[DEBUG] 请求体: %s\n", string(jsonData))

	// 发送请求
	url := fmt.Sprintf(baseURL+bitableBatchPath, appToken, tableToken)
	fmt.Printf("[DEBUG] API URL: %s\n", url)  // 调试日志
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var createResp CreateRecordsResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(body))
	}

	if createResp.Code != 0 {
		// 打印详细错误信息
		fmt.Printf("[DEBUG] API 错误响应:\n")
		fmt.Printf("  Code: %d\n", createResp.Code)
		fmt.Printf("  Msg: %s\n", createResp.Msg)
		fmt.Printf("  完整响应: %s\n", string(body))
		return nil, fmt.Errorf("推送失败: %s", createResp.Msg)
	}

	// 构建响应
	result := &PushToBitableResponse{
		Success: true,
		Message: "推送成功",
	}
	result.Data.RecordsCreated = len(createResp.Data.Records)
	result.Data.TableToken = tableToken

	return result, nil
}

// CreateRecord 创建单条记录
func (c *Client) CreateRecord(appToken, tableToken string, product Product) error {
	_, err := c.PushToBitable(appToken, tableToken, []Product{product})
	return err
}

// GetBitableInfoResponse 获取多维表格信息响应
type GetBitableInfoResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		HasMore  bool `json:"has_more"`
		PageSize int  `json:"page_size"`
		PageToken string `json:"page_token"`
		Total    int `json:"total"`
		Items []struct {
			TableID  string `json:"table_id"`
			Revision int    `json:"revision"`
			Name     string `json:"name"`
		} `json:"items"`
	} `json:"data"`
}

// GetBitableTables 获取多维表格的数据表列表
func (c *Client) GetBitableTables(appToken string) ([]string, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables", appToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var infoResp GetBitableInfoResponse
	if err := json.Unmarshal(body, &infoResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if infoResp.Code != 0 {
		return nil, fmt.Errorf("获取表格列表失败: %s", infoResp.Msg)
	}

	tables := make([]string, 0, len(infoResp.Data.Items))
	for _, table := range infoResp.Data.Items {
		tables = append(tables, table.TableID)
	}

	return tables, nil
}

// TableInfo 表格信息
type TableInfo struct {
	TableID string
	Name    string
}

// GetBitableTableInfos 获取多维表格的数据表详细信息列表
func (c *Client) GetBitableTableInfos(appToken string) ([]TableInfo, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables", appToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var infoResp GetBitableInfoResponse
	if err := json.Unmarshal(body, &infoResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if infoResp.Code != 0 {
		return nil, fmt.Errorf("获取表格列表失败: %s", infoResp.Msg)
	}

	tables := make([]TableInfo, 0, len(infoResp.Data.Items))
	for _, table := range infoResp.Data.Items {
		tables = append(tables, TableInfo{
			TableID: table.TableID,
			Name:    table.Name,
		})
	}

	return tables, nil
}

// CreateTableRequest 创建表格请求
type CreateTableRequest struct {
	Table TableCreateSpec `json:"table"`
}

// TableCreateSpec 表格创建规范
type TableCreateSpec struct {
	Name     string        `json:"name"`
	DefaultViewID string    `json:"default_view_id,omitempty"`
	Fields   []FieldCreate `json:"fields"`
}

// FieldCreate 字段创建规范
type FieldCreate struct {
	FieldName string  `json:"field_name"`
	Type      int     `json:"type"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// CreateTableResponse 创建表格响应
type CreateTableResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Table TableInfo `json:"table"`
	} `json:"data"`
}

// CreateTable 创建数据表
func (c *Client) CreateTable(appToken, tableName string, fields []FieldCreate) (*TableInfo, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	reqBody := CreateTableRequest{
		Table: TableCreateSpec{
			Name:   tableName,
			Fields: fields,
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %w", err)
	}

	url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables", appToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var createResp CreateTableResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if createResp.Code != 0 {
		return nil, fmt.Errorf("创建表格失败: %s", createResp.Msg)
	}

	return &createResp.Data.Table, nil
}

// GetTableFieldsResponse 获取表格字段响应
type GetTableFieldsResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items []struct {
			FieldID    string `json:"field_id"`
			FieldName  string `json:"field_name"`
			Type       int    `json:"type"`
			PropertyName string `json:"property_name,omitempty"`
		} `json:"items"`
	} `json:"data"`
}

// GetTableFields 获取表格字段列表
func (c *Client) GetTableFields(appToken, tableToken string) (map[string]string, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables/%s/fields", appToken, tableToken)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var fieldsResp GetTableFieldsResponse
	if err := json.Unmarshal(body, &fieldsResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if fieldsResp.Code != 0 {
		return nil, fmt.Errorf("获取字段列表失败: %s", fieldsResp.Msg)
	}

	fields := make(map[string]string)
	for _, field := range fieldsResp.Data.Items {
		fields[field.FieldName] = field.FieldID
	}

	return fields, nil
}

// CreateFieldRequest 创建字段请求
type CreateFieldRequest struct {
	Field FieldCreate `json:"field"`
}

// CreateFieldResponse 创建字段响应
type CreateFieldResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Field struct {
			FieldID   string `json:"field_id"`
			FieldName string `json:"field_name"`
		} `json:"field"`
	} `json:"data"`
}

// CreateField 创建字段
func (c *Client) CreateField(appToken, tableToken string, field FieldCreate) (string, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return "", fmt.Errorf("获取访问令牌失败: %w", err)
	}

	reqBody := CreateFieldRequest{
		Field: field,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("构建请求失败: %w", err)
	}

	url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables/%s/fields", appToken, tableToken)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	var createResp CreateFieldResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if createResp.Code != 0 {
		return "", fmt.Errorf("创建字段失败: %s", createResp.Msg)
	}

	return createResp.Data.Field.FieldID, nil
}

// GetTableRecordsResponse 获取表格记录响应
type GetTableRecordsResponse struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		HasMore   bool `json:"has_more"`
		PageSize  int  `json:"page_size"`
		PageToken string `json:"page_token"`
		Total     int  `json:"total"`
		Items     []struct {
			RecordID string                 `json:"record_id"`
			Fields   map[string]interface{} `json:"fields"`
		} `json:"items"`
	} `json:"data"`
}

// GetTableRecords 获取表格中的所有记录（用于去重）
func (c *Client) GetTableRecords(appToken, tableToken string) ([]map[string]interface{}, error) {
	token, err := c.GetTenantAccessToken()
	if err != nil {
		return nil, fmt.Errorf("获取访问令牌失败: %w", err)
	}

	var allRecords []map[string]interface{}
	pageToken := ""

	for {
		url := fmt.Sprintf(baseURL+"/open-apis/bitable/v1/apps/%s/tables/%s/records?page_size=100", appToken, tableToken)
		if pageToken != "" {
			url += "&page_token=" + pageToken
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.httpCli.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		var recordsResp GetTableRecordsResponse
		if err := json.Unmarshal(body, &recordsResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		if recordsResp.Code != 0 {
			return nil, fmt.Errorf("获取记录失败: %s", recordsResp.Msg)
		}

		for _, item := range recordsResp.Data.Items {
			allRecords = append(allRecords, item.Fields)
		}

		if !recordsResp.Data.HasMore {
			break
		}

		pageToken = recordsResp.Data.PageToken
	}

	return allRecords, nil
}
