package mtop

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client MTOP API 客户端
type Client struct {
	httpClient    *http.Client
	baseURL       string
	token         string
	appKey        string
	cookies       []*http.Cookie
	antiBot       *AntiBotMiddleware // 反爬虫中间件
	headerBuilder *HeaderBuilder     // 请求头构建器
	delayManager  *DelayManager      // 延迟管理器
}

// AntiBotMiddleware 反爬虫中间件
type AntiBotMiddleware struct {
	enabled       bool
	headerBuilder *HeaderBuilder
	delayManager  *DelayManager
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

// WithAntiBotConfig 设置反爬虫配置
func WithAntiBotConfig(enabled bool, minDelay, maxDelay int) ClientOption {
	return func(c *Client) {
		if !enabled {
			return
		}
		headerBuilder := NewHeaderBuilder(globalUAPool)
		delayManager := NewDelayManager(minDelay, maxDelay)

		c.antiBot = &AntiBotMiddleware{
			enabled:       true,
			headerBuilder: headerBuilder,
			delayManager:  delayManager,
		}
		c.headerBuilder = headerBuilder
		c.delayManager = delayManager
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
	// 如果启用反爬虫，先执行延迟
	if c.antiBot != nil && c.antiBot.enabled {
		c.antiBot.delayManager.Wait()
	}

	// 构建请求
	httpRequest, err := c.BuildRequest(req)
	if err != nil {
		return nil, err
	}

	// 如果启用反爬虫，应用随机请求头
	if c.antiBot != nil && c.antiBot.enabled {
		c.applyAntiBotHeaders(httpRequest)
	}

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

	// 解析响应
	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	return &result, nil
}

// applyAntiBotHeaders 应用反爬虫请求头
func (c *Client) applyAntiBotHeaders(req *http.Request) {
	randomHeaders := c.antiBot.headerBuilder.BuildRandomHeaders()
	for k, v := range randomHeaders {
		req.Header.Set(k, v)
	}

	// 确保基础请求头不被覆盖
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Origin", "https://www.goofish.com")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
}

// maskCookieValue 隐藏 Cookie 值用于调试输出
func maskCookieValue(value string) string {
	if len(value) <= 10 {
		return value[:2] + "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
