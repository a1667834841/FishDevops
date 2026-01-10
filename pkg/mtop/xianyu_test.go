package mtop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestSignatureGeneration 测试签名生成
func TestSignatureGeneration(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		appKey   string
		data     string
		wantSign string // 预期的MD5签名
	}{
		{
			name:     "基本签名",
			token:    "test123",
			appKey:   "34839810",
			data:     `{"pageNum":1,"pageSize":20}`,
			wantSign: "", // 实际运行时会计算
		},
		{
			name:     "空数据",
			token:    "test123",
			appKey:   "34839810",
			data:     `{}`,
		},
		{
			name:     "复杂数据",
			token:    "abc123",
			appKey:   "34839810",
			data:     `{"itemId":"","pageSize":30,"pageNumber":4,"machId":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Generate(tt.data, GenerateOptions{
				Token:  tt.token,
				AppKey: tt.appKey,
			})
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// 验证签名字符串格式
			expectedSignStr := tt.token + "&" + result.T + "&" + tt.appKey + "&" + tt.data
			if result.SignString != expectedSignStr {
				t.Errorf("SignString = %s, want %s", result.SignString, expectedSignStr)
			}

			// 验证签名长度（MD5应该是32位）
			if len(result.Sign) != 32 {
				t.Errorf("Sign length = %d, want 32", len(result.Sign))
			}

			// 验证返回字段
			if result.AppKey != tt.appKey {
				t.Errorf("AppKey = %s, want %s", result.AppKey, tt.appKey)
			}
			if result.Token != tt.token {
				t.Errorf("Token = %s, want %s", result.Token, tt.token)
			}
			if result.Data != tt.data {
				t.Errorf("Data = %s, want %s", result.Data, tt.data)
			}
		})
	}
}

// TestGenerateWithMap 测试使用map生成签名
func TestGenerateWithMap(t *testing.T) {
	data := map[string]interface{}{
		"pageNum":  1,
		"pageSize": 20,
		"itemId":   "",
	}

	result, err := Generate(data, GenerateOptions{
		Token:  "test123",
		AppKey: "34839810",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证数据被序列化为JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result.Data), &parsed); err != nil {
		t.Fatalf("Data is not valid JSON: %v", err)
	}

	if parsed["pageNum"].(float64) != 1 {
		t.Errorf("pageNum = %v, want 1", parsed["pageNum"])
	}
}

// TestGetTokenFromCookies 测试从Cookie解析token
func TestGetTokenFromCookies(t *testing.T) {
	tests := []struct {
		name      string
		cookies   []*http.Cookie
		wantToken string
	}{
		{
			name: "正常token",
			cookies: []*http.Cookie{
				{Name: "_m_h5_tk", Value: "abc123_1234567890"},
				{Name: "other", Value: "value"},
			},
			wantToken: "abc123",
		},
		{
			name: "无下划线",
			cookies: []*http.Cookie{
				{Name: "_m_h5_tk", Value: "abc123"},
			},
			wantToken: "abc123",
		},
		{
			name: "没有token cookie",
			cookies: []*http.Cookie{
				{Name: "other", Value: "value"},
			},
			wantToken: "",
		},
		{
			name:      "空cookie列表",
			cookies:   []*http.Cookie{},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTokenFromCookies(tt.cookies)
			if got != tt.wantToken {
				t.Errorf("GetTokenFromCookies() = %s, want %s", got, tt.wantToken)
			}
		})
	}
}

// TestGetTokenFromCookieString 测试从字符串解析token
func TestGetTokenFromCookieString(t *testing.T) {
	tests := []struct {
		name      string
		cookieStr string
		wantToken string
	}{
		{
			name:      "正常格式",
			cookieStr: "_m_h5_tk=abc123_1234567890; other=value",
			wantToken: "abc123",
		},
		{
			name:      "无分号",
			cookieStr: "_m_h5_tk=abc123_1234567890",
			wantToken: "abc123",
		},
		{
			name:      "无token",
			cookieStr: "other=value",
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTokenFromCookieString(tt.cookieStr)
			if got != tt.wantToken {
				t.Errorf("GetTokenFromCookieString() = %s, want %s", got, tt.wantToken)
			}
		})
	}
}

// TestMatchFilter 测试过滤功能
func TestMatchFilter(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		options GuessYouLikeOptions
		item    FeedItem
		want    bool
	}{
		{
			name: "无过滤条件",
			options: GuessYouLikeOptions{
				MinWantCount: 0,
				DaysWithin:   0,
			},
			item: FeedItem{
				WantCount:    5,
				PublishTimeTS: now.Add(-2 * 24 * time.Hour).UnixMilli(),
			},
			want: true,
		},
		{
			name: "想要人数过滤-通过",
			options: GuessYouLikeOptions{
				MinWantCount: 10,
				DaysWithin:   0,
			},
			item: FeedItem{
				WantCount:    15,
				PublishTimeTS: now.Add(-2 * 24 * time.Hour).UnixMilli(),
			},
			want: true,
		},
		{
			name: "想要人数过滤-不通过",
			options: GuessYouLikeOptions{
				MinWantCount: 10,
				DaysWithin:   0,
			},
			item: FeedItem{
				WantCount:    5,
				PublishTimeTS: now.Add(-2 * 24 * time.Hour).UnixMilli(),
			},
			want: false,
		},
		{
			name: "时间范围过滤-通过",
			options: GuessYouLikeOptions{
				MinWantCount: 0,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:    5,
				PublishTimeTS: now.Add(-3 * 24 * time.Hour).UnixMilli(),
			},
			want: true,
		},
		{
			name: "时间范围过滤-不通过",
			options: GuessYouLikeOptions{
				MinWantCount: 0,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:    5,
				PublishTimeTS: now.Add(-10 * 24 * time.Hour).UnixMilli(),
			},
			want: false,
		},
		{
			name: "组合过滤-通过",
			options: GuessYouLikeOptions{
				MinWantCount: 10,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:    15,
				PublishTimeTS: now.Add(-3 * 24 * time.Hour).UnixMilli(),
			},
			want: true,
		},
		{
			name: "组合过滤-想要人数不通过",
			options: GuessYouLikeOptions{
				MinWantCount: 10,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:    5,
				PublishTimeTS: now.Add(-3 * 24 * time.Hour).UnixMilli(),
			},
			want: false,
		},
		{
			name: "组合过滤-时间不通过",
			options: GuessYouLikeOptions{
				MinWantCount: 10,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:    15,
				PublishTimeTS: now.Add(-10 * 24 * time.Hour).UnixMilli(),
			},
			want: false,
		},
		{
			name: "无时间戳-应该通过",
			options: GuessYouLikeOptions{
				MinWantCount: 0,
				DaysWithin:   7,
			},
			item: FeedItem{
				WantCount:     5,
				PublishTimeTS: 0,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.options.MatchFilter(tt.item)
			if got != tt.want {
				t.Errorf("MatchFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFeedItemFields 测试FeedItem字段完整性
func TestFeedItemFields(t *testing.T) {
	// 创建一个完整的FeedItem用于测试JSON序列化
	item := FeedItem{
		// 基础信息
		ItemID:     "test_item_123",
		Title:      "测试商品标题",
		ImageURL:   "https://example.com/image.jpg",
		CategoryID: 50023914,
		Location:   "上海",

		// 价格与行情
		Price:         "100.00",
		PriceOriginal: "150.00",
		UnitPrice:     "10.00/斤",

		// 热度与流量
		WantCount: 25,
		ViewCount: 150,
		Status:    "online",
		ShopLevel: "level5",

		// 卖家与服务
		SellerNick:   "测试卖家",
		SellerCredit: "卖家信用极好",
		FreeShipping: true,

		// 时间与活跃度
		PublishTime:    "2026-01-10 10:30:00",
		PublishTimeTS:  1736476800000,
		ModifiedTime:   "2026-01-11 15:20:00",
		ModifiedTimeTS: 1736587200000,

		// 其他
		IsIdle:        true,
		VideoCoverURL: "https://example.com/video.jpg",
		VideoURL:      "https://example.com/video.mp4",
		Condition:     "全新",
		SoldOut:       false,
		Like:          false,
		Tags:          []string{"level5", "卖家信用极好", "包邮"},
		IsVideo:       true,
	}

	// 测试JSON序列化
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 验证可以反序列化
	var decoded FeedItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON反序列化失败: %v", err)
	}

	// 验证关键字段
	if decoded.ItemID != item.ItemID {
		t.Errorf("ItemID = %s, want %s", decoded.ItemID, item.ItemID)
	}
	if decoded.WantCount != item.WantCount {
		t.Errorf("WantCount = %d, want %d", decoded.WantCount, item.WantCount)
	}
	if decoded.FreeShipping != item.FreeShipping {
		t.Errorf("FreeShipping = %v, want %v", decoded.FreeShipping, item.FreeShipping)
	}
	if decoded.ShopLevel != item.ShopLevel {
		t.Errorf("ShopLevel = %s, want %s", decoded.ShopLevel, item.ShopLevel)
	}
}

// TestClientDo 测试客户端请求
func TestClientDo(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		// 验证请求头 (Go http package adds charset=UTF-8)
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" && ct != "application/x-www-form-urlencoded; charset=UTF-8" {
			t.Errorf("Content-Type = %s", ct)
		}

		// 验证URL参数
		query := r.URL.Query()
		if query.Get("appKey") != "34839810" {
			t.Errorf("appKey = %s", query.Get("appKey"))
		}
		if query.Get("api") != "mtop.taobao.idlehome.home.webpc.feed" {
			t.Errorf("api = %s", query.Get("api"))
		}

		// 验证sign存在
		sign := query.Get("sign")
		if sign == "" || len(sign) != 32 {
			t.Errorf("sign = %s, want 32 char MD5", sign)
		}

		// 返回模拟响应
		response := Response{
			Ret:  []string{"SUCCESS::调用成功"},
			V:    "1.0",
			Data: json.RawMessage(`{"data":{"hasNext":true,"items":[],"pageNum":1,"pageSize":30,"totalItem":0}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient("test_token_123", "34839810",
		WithBaseURL(server.URL),
	)

	// 发送请求
	resp, err := client.Do(Request{
		API:    "mtop.taobao.idlehome.home.webpc.feed",
		Data:   map[string]interface{}{"pageNum": 1, "pageSize": 30},
		Method: "POST",
	})

	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	// 验证响应成功
	success := false
	for _, r := range resp.Ret {
		if r == "SUCCESS" || r == "SUCCESS::调用成功" {
			success = true
			break
		}
	}
	if !success {
		t.Errorf("Ret = %v, want SUCCESS", resp.Ret)
	}
}

// TestGuessYouLike 测试获取猜你喜欢
func TestGuessYouLike(t *testing.T) {
	// 创建模拟响应数据
	mockCardData := map[string]interface{}{
		"cardList": []map[string]interface{}{
			{
				"cardData": map[string]interface{}{
					"categoryId": "50023914",
					"status":     "online",
					"viewCount":  150,
					"detailParams": map[string]interface{}{
						"itemId": "item123",
						"picUrl": "https://example.com/image.jpg",
						"title":  "测试商品",
						"isVideo": "1",
					},
					"user": map[string]interface{}{
						"userNick": "测试卖家",
					},
					"priceInfo": map[string]interface{}{
						"price":    "100.00",
						"oriPrice": "150.00",
					},
					"unitPriceInfo": map[string]interface{}{
						"price": "10.00",
					},
					"city": "上海",
					"attributeMap": map[string]string{
						"gmtShelf":     fmt.Sprintf("%d", time.Now().Add(-2*24*time.Hour).UnixMilli()),
						"gmtModified":  fmt.Sprintf("%d", time.Now().UnixMilli()),
						"freeShipping": "1",
					},
					"fishTags": map[string]interface{}{
						"r4": map[string]interface{}{
							"tagList": []map[string]interface{}{
								{
									"data": map[string]interface{}{
										"labelId": "955",
										"type":    "img",
										"content": "level5",
									},
									"utParams": map[string]interface{}{
										"args": map[string]interface{}{
											"content": "level5",
										},
									},
								},
								{
									"data": map[string]interface{}{
										"content": "卖家信用极好",
									},
								},
								{
									"data": map[string]interface{}{
										"content": "25人想要",
									},
								},
							},
						},
					},
				},
			},
		},
		"feedsCount": 1,
		"nextPage":   false,
	}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Ret:  []string{"SUCCESS::调用成功"},
			V:    "1.0",
			Data: json.RawMessage(mustMarshalJSON(mockCardData)),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient("test_token_123", "34839810",
		WithBaseURL(server.URL),
	)

	// 获取数据
	items, err := client.GuessYouLike("", 1, GuessYouLikeOptions{
		MinWantCount: 10,
		DaysWithin:   7,
	})
	if err != nil {
		t.Fatalf("GuessYouLike() error = %v", err)
	}

	// 验证返回的商品
	if len(items) != 1 {
		t.Fatalf("Got %d items, want 1", len(items))
	}

	item := items[0]

	// 验证基础字段
	if item.ItemID != "item123" {
		t.Errorf("ItemID = %s, want item123", item.ItemID)
	}
	if item.Title != "测试商品" {
		t.Errorf("Title = %s, want 测试商品", item.Title)
	}
	if item.CategoryID != 50023914 {
		t.Errorf("CategoryID = %s, want 50023914", item.CategoryID)
	}
	if item.Location != "上海" {
		t.Errorf("Location = %s, want 上海", item.Location)
	}

	// 验证价格字段
	if item.Price != "100.00" {
		t.Errorf("Price = %s, want 100.00", item.Price)
	}
	if item.PriceOriginal != "150.00" {
		t.Errorf("PriceOriginal = %s, want 150.00", item.PriceOriginal)
	}
	if item.UnitPrice != "10.00" {
		t.Errorf("UnitPrice = %s, want 10.00", item.UnitPrice)
	}

	// 验证热度字段
	if item.WantCount != 25 {
		t.Errorf("WantCount = %d, want 25", item.WantCount)
	}
	if item.ViewCount != 150 {
		t.Errorf("ViewCount = %d, want 150", item.ViewCount)
	}
	if item.Status != "online" {
		t.Errorf("Status = %s, want online", item.Status)
	}

	// 验证卖家字段
	if item.SellerNick != "测试卖家" {
		t.Errorf("SellerNick = %s, want 测试卖家", item.SellerNick)
	}
	if item.SellerCredit != "卖家信用极好" {
		t.Errorf("SellerCredit = %s, want 卖家信用极好", item.SellerCredit)
	}

	// 验证服务字段
	if !item.FreeShipping {
		t.Error("FreeShipping = false, want true")
	}
	if !item.IsVideo {
		t.Error("IsVideo = false, want true")
	}

	// 验证店铺级别
	if item.ShopLevel != "level5" {
		t.Errorf("ShopLevel = %s, want level5", item.ShopLevel)
	}

	// 验证标签
	if !contains(item.Tags, "level5") {
		t.Error("Tags should contain 'level5'")
	}
	if !contains(item.Tags, "卖家信用极好") {
		t.Error("Tags should contain '卖家信用极好'")
	}
}

// TestGuessYouLikeWithFilters 测试带过滤条件的获取
func TestGuessYouLikeWithFilters(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		options GuessYouLikeOptions
		wantLen int // 期望返回的商品数量
	}{
		{
			name: "只要想要>=20的商品",
			options: GuessYouLikeOptions{
				MinWantCount: 20,
				DaysWithin:   0,
			},
			wantLen: 1, // 只有25人想要的那个商品
		},
		{
			name: "只要想要>=30的商品",
			options: GuessYouLikeOptions{
				MinWantCount: 30,
				DaysWithin:   0,
			},
			wantLen: 0, // 没有满足条件的商品
		},
		{
			name: "无过滤条件",
			options: GuessYouLikeOptions{
				MinWantCount: 0,
				DaysWithin:   0,
			},
			wantLen: 2, // 所有商品
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟数据（2个商品，一个25人想要，一个5人想要）
			mockCardData := map[string]interface{}{
				"cardList": []map[string]interface{}{
					{
						"cardData": map[string]interface{}{
							"detailParams": map[string]interface{}{
								"itemId": "item123",
								"picUrl": "https://example.com/image.jpg",
								"title":  "高热度商品",
							},
							"priceInfo": map[string]interface{}{
								"price": "100.00",
							},
							"city": "上海",
							"attributeMap": map[string]string{
								"gmtShelf": fmt.Sprintf("%d", now.Add(-2*24*time.Hour).UnixMilli()),
							},
							"fishTags": map[string]interface{}{
								"r4": map[string]interface{}{
									"tagList": []map[string]interface{}{
										{"data": map[string]interface{}{"content": "25人想要"}},
									},
								},
							},
						},
					},
					{
						"cardData": map[string]interface{}{
							"detailParams": map[string]interface{}{
								"itemId": "item456",
								"picUrl": "https://example.com/image2.jpg",
								"title":  "低热度商品",
							},
							"priceInfo": map[string]interface{}{
								"price": "50.00",
							},
							"city": "北京",
							"attributeMap": map[string]string{
								"gmtShelf": fmt.Sprintf("%d", now.Add(-1*24*time.Hour).UnixMilli()),
							},
							"fishTags": map[string]interface{}{
								"r4": map[string]interface{}{
									"tagList": []map[string]interface{}{
										{"data": map[string]interface{}{"content": "5人想要"}},
									},
								},
							},
						},
					},
				},
				"feedsCount": 2,
				"nextPage":   false,
			}

			// 创建测试服务器
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := Response{
					Ret:  []string{"SUCCESS::调用成功"},
					V:    "1.0",
					Data: json.RawMessage(mustMarshalJSON(mockCardData)),
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client := NewClient("test_token_123", "34839810",
				WithBaseURL(server.URL),
			)

			items, err := client.GuessYouLike("", 1, tt.options)
			if err != nil {
				t.Fatalf("GuessYouLike() error = %v", err)
			}

			if len(items) != tt.wantLen {
				t.Errorf("Got %d items, want %d", len(items), tt.wantLen)
			}
		})
	}
}

// TestConvertMapSliceToHTTPCookies 测试Cookie转换
func TestConvertMapSliceToHTTPCookies(t *testing.T) {
	cookieMaps := []map[string]string{
		{"name": "_m_h5_tk", "value": "abc123_123456", "domain": ".goofish.com"},
		{"name": "cna", "value": "xyz789", "path": "/"},
	}

	cookies := ConvertMapSliceToHTTPCookies(cookieMaps)

	if len(cookies) != 2 {
		t.Fatalf("Got %d cookies, want 2", len(cookies))
	}

	if cookies[0].Name != "_m_h5_tk" {
		t.Errorf("cookies[0].Name = %s, want _m_h5_tk", cookies[0].Name)
	}
	if cookies[0].Value != "abc123_123456" {
		t.Errorf("cookies[0].Value = %s, want abc123_123456", cookies[0].Value)
	}
	if cookies[0].Domain != ".goofish.com" {
		t.Errorf("cookies[0].Domain = %s, want .goofish.com", cookies[0].Domain)
	}

	if cookies[1].Path != "/" {
		t.Errorf("cookies[1].Path = %s, want /", cookies[1].Path)
	}
}

// TestGuessYouLikeOptionsDefaults 测试默认选项
func TestGuessYouLikeOptionsDefaults(t *testing.T) {
	options := GuessYouLikeOptions{}

	// 验证默认值不是零值，因为在GuessYouLike函数中会设置默认值
	// 这里只是测试结构体本身
	if options.MaxPages != 0 {
		t.Errorf("MaxPages default = %d, want 0", options.MaxPages)
	}
	if options.StartPage != 0 {
		t.Errorf("StartPage default = %d, want 0", options.StartPage)
	}
}

// mustMarshalJSON 辅助函数：JSON序列化，失败时panic
func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// contains 辅助函数：检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
