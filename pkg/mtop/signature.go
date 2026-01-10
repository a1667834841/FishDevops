package mtop

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// SignatureResult 签名结果
type SignatureResult struct {
	Sign       string `json:"sign"`        // MD5签名
	T          string `json:"t"`           // 时间戳
	AppKey     string `json:"appKey"`      // 应用Key
	Token      string `json:"token"`       // MTOP Token
	Data       string `json:"data"`        // 请求数据JSON字符串
	SignString string `json:"signString"`  // 签名字符串(调试用)
}

// GenerateOptions 签名生成选项
type GenerateOptions struct {
	Token     string // 自定义 token（默认从 cookie 获取）
	Timestamp string // 自定义时间戳（默认当前时间）
	AppKey    string // 应用 key（默认 34839810）
}

// Generate 生成 MTOP 签名
// data: 请求数据对象或 JSON 字符串
// options: 可选参数
func Generate(data interface{}, options GenerateOptions) (*SignatureResult, error) {
	// 处理 data 参数，转换为 JSON 字符串
	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("序列化数据失败: %w", err)
		}
		dataStr = string(jsonBytes)
	}

	// 设置默认 appKey
	appKey := options.AppKey
	if appKey == "" {
		appKey = "34839810" // 闲鱼默认 appKey
	}

	// 设置默认时间戳（13位毫秒时间戳）
	timestamp := options.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	// 获取 token（必须从外部传入）
	token := options.Token
	if token == "" {
		return nil, fmt.Errorf("token 不能为空，请从 Cookie 中获取 _m_h5_tk")
	}

	// 生成签名字符串: token&timestamp&appKey&data
	signStr := fmt.Sprintf("%s&%s&%s&%s", token, timestamp, appKey, dataStr)

	// MD5 加密
	hash := md5.New()
	hash.Write([]byte(signStr))
	sign := hex.EncodeToString(hash.Sum(nil))

	return &SignatureResult{
		Sign:       sign,
		T:          timestamp,
		AppKey:     appKey,
		Token:      token,
		Data:       dataStr,
		SignString: signStr,
	}, nil
}
