package mtop

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client MTOP API 客户端
type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	appKey     string
	cookies    []*http.Cookie
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithHTTPClient 设置自定义 HTTP 客户端
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL 设置基础 URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithCookies 设置 Cookies
func WithCookies(cookies []*http.Cookie) ClientOption {
	return func(c *Client) {
		c.cookies = cookies
	}
}

// WithCookieString 设置 Cookies（从字符串）
func WithCookieString(cookieStr string) ClientOption {
	return func(c *Client) {
		// 解析 cookie 字符串
		parts := strings.Split(cookieStr, ";")
		cookies := make([]*http.Cookie, 0, len(parts))
		for _, part := range parts {
			kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
			if len(kv) == 2 {
				cookies = append(cookies, &http.Cookie{
					Name:  kv[0],
					Value: kv[1],
				})
			}
		}
		c.cookies = cookies
	}
}

// NewClient 创建新的 MTOP 客户端
func NewClient(token string, appKey string, opts ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://h5api.m.goofish.com/h5",
		token:   token,
		appKey:  appKey,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// SetToken 设置 token
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetCookies 设置 cookies
func (c *Client) SetCookies(cookies []*http.Cookie) {
	c.cookies = cookies
}

// Request API 请求参数
type Request struct {
	API    string
	Data   interface{}
	Method string
}

// Response API 响应
type Response struct {
	Ret  []string `json:"ret"`
	V    string   `json:"v"`
	Data json.RawMessage `json:"data"`
}

// Do 发送请求
func (c *Client) Do(req Request) (*Response, error) {
	// 序列化 data
	var dataStr string
	switch v := req.Data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("序列化数据失败: %w", err)
		}
		dataStr = string(jsonBytes)
	}

	// 生成签名
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	signResult, err := Generate(dataStr, GenerateOptions{
		Token:     c.token,
		Timestamp: timestamp,
		AppKey:    c.appKey,
	})
	if err != nil {
		return nil, fmt.Errorf("生成签名失败: %w", err)
	}

	// 构建请求 URL
	apiURL := fmt.Sprintf("%s/%s/1.0/", c.baseURL, req.API)

	// 构建 query 参数
	values := url.Values{}
	values.Set("jsv", "2.7.2")
	values.Set("appKey", c.appKey)
	values.Set("t", timestamp)
	values.Set("sign", signResult.Sign)
	values.Set("v", "1.0")
	values.Set("type", "originaljson")
	values.Set("accountSite", "xianyu")
	values.Set("dataType", "json")
	values.Set("timeout", "20000")
	values.Set("api", req.API)
	values.Set("sessionOption", "AutoLoginOnly")
	values.Set("spm_cnt", "a21ybx.home.0.0")

	// 创建表单 body（data 作为表单字段）
	formData := url.Values{}
	formData.Set("data", dataStr)

	// 创建 HTTP 请求
	httpRequest, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 将参数添加到 query
	httpRequest.URL.RawQuery = values.Encode()

	// 设置请求头
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	httpRequest.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	httpRequest.Header.Set("Accept", "application/json, text/plain, */*")
	httpRequest.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	httpRequest.Header.Set("Origin", "https://www.goofish.com")
	httpRequest.Header.Set("Referer", "https://www.goofish.com/")
	httpRequest.Header.Set("Sec-Fetch-Dest", "empty")
	httpRequest.Header.Set("Sec-Fetch-Mode", "cors")
	httpRequest.Header.Set("Sec-Fetch-Site", "same-site")
	httpRequest.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	httpRequest.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	httpRequest.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)

	// 添加 cookies
	for _, cookie := range c.cookies {
		httpRequest.AddCookie(cookie)
	}

	// 打印调试信息
	// fmt.Printf("\n[调试] 请求URL: %s\n", httpRequest.URL.String())
	// fmt.Printf("[调试] 请求Body: %s\n", formData.Encode())

	// 发送请求
	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 打印调试信息
	// fmt.Printf("\n[调试] API响应: status=%d, body=%s\n", resp.StatusCode, string(body))

	// 解析响应
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	return &result, nil
}
