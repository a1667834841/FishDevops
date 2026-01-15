package mtop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestBuilder 请求构建器
type RequestBuilder struct {
	client *Client
	req    Request
}

// BuildRequest 构建HTTP请求
func (c *Client) BuildRequest(req Request) (*http.Request, error) {
	builder := &RequestBuilder{client: c, req: req}

	// 序列化数据
	dataStr := builder.serializeData()

	// 生成签名
	timestamp, sign := builder.generateSignature(dataStr)

	// 构建 URL
	apiURL := fmt.Sprintf("%s/%s/1.0/", c.baseURL, req.API)
	values := builder.buildQueryParams(timestamp, sign)
	formData := builder.buildFormData(dataStr)

	// 创建 HTTP 请求
	httpRequest, err := http.NewRequest("POST", apiURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpRequest.URL.RawQuery = values.Encode()

	// 如果未启用反爬虫，使用固定请求头；否则在 Do() 方法中应用随机请求头
	if c.antiBot == nil || !c.antiBot.enabled {
		builder.setHeaders(httpRequest)
	}

	builder.addCookies(httpRequest)

	return httpRequest, nil
}

// serializeData 序列化请求数据
func (b *RequestBuilder) serializeData() string {
	switch v := b.req.Data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	}
}

// generateSignature 生成签名
func (b *RequestBuilder) generateSignature(dataStr string) (string, string) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	signResult, _ := Generate(dataStr, GenerateOptions{
		Token:     b.client.token,
		Timestamp: timestamp,
		AppKey:    b.client.appKey,
	})
	return timestamp, signResult.Sign
}

// buildQueryParams 构建查询参数
func (b *RequestBuilder) buildQueryParams(timestamp, sign string) url.Values {
	values := url.Values{}
	values.Set("jsv", "2.7.2")
	values.Set("appKey", b.client.appKey)
	values.Set("t", timestamp)
	values.Set("sign", sign)
	values.Set("v", "1.0")
	values.Set("type", "originaljson")
	values.Set("accountSite", "xianyu")
	values.Set("dataType", "json")
	values.Set("timeout", "20000")
	values.Set("api", b.req.API)
	values.Set("sessionOption", "AutoLoginOnly")

	// 根据API类型设置不同的 spm_cnt
	spmCnt := "a21ybx.home.0.0"
	if strings.Contains(b.req.API, "detail") {
		spmCnt = "a21ybx.item.0.0"
	}
	values.Set("spm_cnt", spmCnt)

	return values
}

// buildFormData 构建表单数据
func (b *RequestBuilder) buildFormData(dataStr string) url.Values {
	formData := url.Values{}
	formData.Set("data", dataStr)
	return formData
}

// setHeaders 设置请求头
func (b *RequestBuilder) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Origin", "https://www.goofish.com")
	req.Header.Set("Referer", "https://www.goofish.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
}

// addCookies 添加 Cookies
func (b *RequestBuilder) addCookies(req *http.Request) {
	for _, cookie := range b.client.cookies {
		req.AddCookie(cookie)
	}
}
