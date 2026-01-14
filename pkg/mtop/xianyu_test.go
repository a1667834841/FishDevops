package mtop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
			name:   "空数据",
			token:  "test123",
			appKey: "34839810",
			data:   `{}`,
		},
		{
			name:   "复杂数据",
			token:  "abc123",
			appKey: "34839810",
			data:   `{"itemId":"","pageSize":30,"pageNumber":4,"machId":""}`,
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
				WantCount:     5,
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
				WantCount:     15,
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
				WantCount:     5,
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
				WantCount:     5,
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
				WantCount:     5,
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
				WantCount:     15,
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
				WantCount:     5,
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
				WantCount:     15,
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
		Price: "100.00",

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
						"itemId":  "item123",
						"picUrl":  "https://example.com/image.jpg",
						"title":   "测试商品",
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
		t.Errorf("CategoryID = %d, want 50023914", item.CategoryID)
	}
	if item.Location != "上海" {
		t.Errorf("Location = %s, want 上海", item.Location)
	}

	// 验证价格字段
	if item.Price != "100.00" {
		t.Errorf("Price = %s, want 100.00", item.Price)
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

// ==================== 商品详情 API 测试 ====================

// TestFetchItemDetail 测试获取商品详情
func TestFetchItemDetail(t *testing.T) {
	// 创建模拟响应数据 - 匹配实际 API 返回结构 (itemDO + sellerDO)
	mockDetailData := map[string]interface{}{
		"itemDO": map[string]interface{}{
			"itemId":        int64(123), // API 返回的是数字
			"title":         "测试商品标题",
			"desc":          "这是商品简述",
			"categoryId":    50023914,
			"soldPrice":     "100.00",
			"priceUnit":     "元",
			"itemStatus":    0,
			"itemStatusStr": "online",
			"wantCnt":       25,
			"browseCnt":     150,
			"collectCnt":    10,
			"gmtCreate":     int64(1736476800000),
			"quantity":      5,
			"imageInfos": []map[string]interface{}{
				{"url": "https://example.com/img1.jpg", "major": true, "widthSize": 800, "heightSize": 600},
				{"url": "https://example.com/img2.jpg", "major": false, "widthSize": 800, "heightSize": 600},
				{"url": "https://example.com/img3.jpg", "major": false, "widthSize": 800, "heightSize": 600},
			},
			"cpvLabels": []map[string]interface{}{
				{"propertyId": int64(1), "propertyName": "成色", "valueId": int64(1), "valueName": "95新"},
			},
			"commonTags": []map[string]interface{}{
				{"text": "包邮"},
				{"text": "验货宝"},
			},
			"transportFee": "0",
		},
		"sellerDO": map[string]interface{}{
			"sellerId":          int64(456),
			"nick":              "测试卖家",
			"uniqueName":        "seller123",
			"city":              "上海",
			"portraitUrl":       "https://example.com/avatar.jpg",
			"signature":         "欢迎选购",
			"itemCount":         10,
			"hasSoldNumInteger": 50,
			"userRegDay":        365,
			"zhimaAuth":         true,
			"zhumaLevelInfo": map[string]interface{}{
				"levelCode": "excellent",
				"levelName": "信用极好",
			},
		},
	}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求参数
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		query := r.URL.Query()
		if query.Get("api") != "mtop.taobao.idle.pc.detail" {
			t.Errorf("api = %s, want mtop.taobao.idle.pc.detail", query.Get("api"))
		}

		// 返回模拟响应
		response := Response{
			Ret:  []string{"SUCCESS::调用成功"},
			V:    "1.0",
			Data: json.RawMessage(mustMarshalJSON(mockDetailData)),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient("test_token_123", "34839810",
		WithBaseURL(server.URL),
	)

	// 获取详情
	detail, err := client.FetchItemDetail("item123")
	if err != nil {
		t.Fatalf("FetchItemDetail() error = %v", err)
	}

	// 验证基础字段
	if detail.ItemID != "123" {
		t.Errorf("ItemID = %s, want 123", detail.ItemID)
	}
	if detail.Title != "测试商品标题" {
		t.Errorf("Title = %s, want 测试商品标题", detail.Title)
	}
	if detail.Desc != "这是商品简述" {
		t.Errorf("Desc = %s, want 这是商品简述", detail.Desc)
	}
	if detail.CategoryID != 50023914 {
		t.Errorf("CategoryID = %d, want 50023914", detail.CategoryID)
	}

	// 验证价格字段
	if detail.Price != "100.00" {
		t.Errorf("Price = %s, want 100.00", detail.Price)
	}
	if detail.SoldPrice != "100.00" {
		t.Errorf("SoldPrice = %s, want 100.00", detail.SoldPrice)
	}

	// 验证卖家字段
	if detail.SellerID != "456" {
		t.Errorf("SellerID = %s, want 456", detail.SellerID)
	}
	if detail.SellerNick != "测试卖家" {
		t.Errorf("SellerNick = %s, want 测试卖家", detail.SellerNick)
	}
	if detail.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("AvatarURL = %s, want https://example.com/avatar.jpg", detail.AvatarURL)
	}
	if detail.SellerCredit != "信用极好" {
		t.Errorf("SellerCredit = %s, want 信用极好", detail.SellerCredit)
	}
	if detail.SellerRegDays != 365 {
		t.Errorf("SellerRegDays = %d, want 365", detail.SellerRegDays)
	}
	if detail.SellerItemCount != 10 {
		t.Errorf("SellerItemCount = %d, want 10", detail.SellerItemCount)
	}
	if detail.SellerSoldCount != 50 {
		t.Errorf("SellerSoldCount = %d, want 50", detail.SellerSoldCount)
	}

	// 验证状态字段
	if detail.Status != "online" {
		t.Errorf("Status = %s, want online", detail.Status)
	}
	if detail.WantCount != 25 {
		t.Errorf("WantCount = %d, want 25", detail.WantCount)
	}
	if detail.ViewCount != 150 {
		t.Errorf("ViewCount = %d, want 150", detail.ViewCount)
	}
	if detail.CollectCount != 10 {
		t.Errorf("CollectCount = %d, want 10", detail.CollectCount)
	}

	// 验证地址字段 (从 sellerDO.city 获取)
	if detail.Location != "上海" {
		t.Errorf("Location = %s, want 上海", detail.Location)
	}

	// 验证时间字段
	if detail.PublishTimeTS != 1736476800000 {
		t.Errorf("PublishTimeTS = %d, want 1736476800000", detail.PublishTimeTS)
	}
	// PublishTime 从时间戳转换，验证格式正确即可
	if detail.PublishTime == "" {
		t.Error("PublishTime should not be empty")
	}

	// 验证商品属性
	if detail.Condition != "95新" {
		t.Errorf("Condition = %s, want 95新", detail.Condition)
	}
	if detail.IsNew {
		t.Error("IsNew = true, want false")
	}
	if !detail.FreeShipping {
		t.Error("FreeShipping = false, want true")
	}

	// 验证库存
	if detail.TotalStock != 5 {
		t.Errorf("TotalStock = %d, want 5", detail.TotalStock)
	}

	// 验证图片列表
	if len(detail.ImageList) != 3 {
		t.Fatalf("ImageList length = %d, want 3", len(detail.ImageList))
	}
	if detail.ImageList[0] != "https://example.com/img1.jpg" {
		t.Errorf("ImageList[0] = %s, want https://example.com/img1.jpg", detail.ImageList[0])
	}
	// 验证主图 (major=true)
	if detail.ImageURL != "https://example.com/img1.jpg" {
		t.Errorf("ImageURL = %s, want https://example.com/img1.jpg", detail.ImageURL)
	}

	// 验证描述 (从 desc 字段获取)
	if detail.Description != "这是商品简述" {
		t.Errorf("Description = %s, want 这是商品简述", detail.Description)
	}

	// 验证标签 (从 commonTags 获取)
	if len(detail.Tags) != 2 {
		t.Fatalf("Tags length = %d, want 2", len(detail.Tags))
	}
	if !contains(detail.Tags, "包邮") {
		t.Error("Tags should contain '包邮'")
	}
	if !contains(detail.Tags, "验货宝") {
		t.Error("Tags should contain '验货宝'")
	}

	// 验证 CPV 属性标签
	if len(detail.CPVLabels) != 1 {
		t.Fatalf("CPVLabels length = %d, want 1", len(detail.CPVLabels))
	}
	if detail.CPVLabels[0].PropertyName != "成色" {
		t.Errorf("CPVLabels[0].PropertyName = %s, want 成色", detail.CPVLabels[0].PropertyName)
	}
}

// TestFetchItemDetailWithEmptyItemID 测试空商品ID
func TestFetchItemDetailWithEmptyItemID(t *testing.T) {
	client := NewClient("test_token_123", "34839810")

	_, err := client.FetchItemDetail("")
	if err == nil {
		t.Fatal("FetchItemDetail() should return error for empty itemID")
	}

	expectedErrMsg := "itemID 不能为空"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message = %s, want %s", err.Error(), expectedErrMsg)
	}
}

// TestFetchItemDetailWithErrorStatus 测试API返回错误状态
func TestFetchItemDetailWithErrorStatus(t *testing.T) {
	// 创建测试服务器返回错误状态
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Ret:  []string{"ERROR::系统错误"},
			V:    "1.0",
			Data: json.RawMessage(`{}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_token_123", "34839810",
		WithBaseURL(server.URL),
	)

	_, err := client.FetchItemDetail("item123")
	if err == nil {
		t.Fatal("FetchItemDetail() should return error for failed API response")
	}

	if !strings.Contains(err.Error(), "详情API返回错误") {
		t.Errorf("Error should contain '详情API返回错误', got: %v", err)
	}
}

// TestFetchItemDetailMinimalData 测试最小数据响应
func TestFetchItemDetailMinimalData(t *testing.T) {
	// 最小必需数据的响应 - 匹配实际 API 返回结构
	mockDetailData := map[string]interface{}{
		"itemDO": map[string]interface{}{
			"itemId":        int64(789),
			"title":         "最小商品",
			"soldPrice":     "50.00",
			"itemStatusStr": "online",
			"imageInfos":    []map[string]interface{}{},
			"commonTags":    []map[string]interface{}{},
			"cpvLabels":     []map[string]interface{}{},
		},
		"sellerDO": map[string]interface{}{
			"sellerId": int64(456),
			"nick":     "卖家456",
			"city":     "北京",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Response{
			Ret:  []string{"SUCCESS::调用成功"},
			V:    "1.0",
			Data: json.RawMessage(mustMarshalJSON(mockDetailData)),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient("test_token_123", "34839810",
		WithBaseURL(server.URL),
	)

	detail, err := client.FetchItemDetail("minimal123")
	if err != nil {
		t.Fatalf("FetchItemDetail() error = %v", err)
	}

	// 验证最小字段
	if detail.ItemID != "789" {
		t.Errorf("ItemID = %s, want 789", detail.ItemID)
	}
	if detail.Title != "最小商品" {
		t.Errorf("Title = %s, want 最小商品", detail.Title)
	}
	if detail.Price != "50.00" {
		t.Errorf("Price = %s, want 50.00", detail.Price)
	}
	if detail.SellerNick != "卖家456" {
		t.Errorf("SellerNick = %s, want 卖家456", detail.SellerNick)
	}
	if detail.Location != "北京" {
		t.Errorf("Location = %s, want 北京", detail.Location)
	}

	// 验证空切片
	if len(detail.ImageList) != 0 {
		t.Errorf("ImageList length = %d, want 0", len(detail.ImageList))
	}
	if len(detail.Tags) != 0 {
		t.Errorf("Tags length = %d, want 0", len(detail.Tags))
	}
}

// TestItemDetailJSONSerialization 测试ItemDetail的JSON序列化
func TestItemDetailJSONSerialization(t *testing.T) {
	detail := &ItemDetail{
		ItemID:        "test123",
		Title:         "测试商品",
		Price:         "100.00",
		WantCount:     25,
		FreeShipping:  true,
		Tags:          []string{"包邮", "验货宝"},
		ImageList:     []string{"https://example.com/img1.jpg"},
		PublishTimeTS: 1736476800000,
	}

	// 测试序列化
	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 测试反序列化
	var decoded ItemDetail
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON反序列化失败: %v", err)
	}

	// 验证关键字段
	if decoded.ItemID != detail.ItemID {
		t.Errorf("ItemID = %s, want %s", decoded.ItemID, detail.ItemID)
	}
	if decoded.WantCount != detail.WantCount {
		t.Errorf("WantCount = %d, want %d", decoded.WantCount, detail.WantCount)
	}
	if !decoded.FreeShipping {
		t.Error("FreeShipping = false, want true")
	}
	if len(decoded.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(decoded.Tags))
	}
}

// TestItemDetailRequest 测试请求参数结构
func TestItemDetailRequest(t *testing.T) {
	req := ItemDetailRequest{
		ItemID: "test_item_123",
	}

	// 序列化
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 验证JSON内容
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON反序列化失败: %v", err)
	}

	if parsed["itemId"] != "test_item_123" {
		t.Errorf("itemId = %v, want test_item_123", parsed["itemId"])
	}
}
