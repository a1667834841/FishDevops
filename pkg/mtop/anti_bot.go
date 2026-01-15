package mtop

import (
	"math/rand"
	"sync"
	"time"
)

// UAPool User-Agent 池（线程安全）
type UAPool struct {
	uas []string
	mu  sync.RWMutex
}

// globalUAPool 全局 UA 池实例
var globalUAPool = &UAPool{
	uas: []string{
		// Chrome on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
		// Chrome on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		// Edge on Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",
		// Safari on macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
		// Firefox
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		// Chrome on Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	},
}

// GetRandomUA 从池中随机获取一个 UA
func (p *UAPool) GetRandomUA() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(p.uas))
	return p.uas[idx]
}

// DelayManager 延迟管理器
type DelayManager struct {
	minMs int
	maxMs int
}

// NewDelayManager 创建延迟管理器
func NewDelayManager(minMs, maxMs int) *DelayManager {
	return &DelayManager{minMs: minMs, maxMs: maxMs}
}

// Wait 执行延迟等待
func (d *DelayManager) Wait() {
	if d.minMs <= 0 || d.maxMs <= 0 {
		return
	}
	min := time.Duration(d.minMs) * time.Millisecond
	max := time.Duration(d.maxMs) * time.Millisecond
	rand.Seed(time.Now().UnixNano())
	delay := min + time.Duration(rand.Int63n(int64(max-min)))
	time.Sleep(delay)
}

// HeaderBuilder 请求头构建器
type HeaderBuilder struct {
	uaPool      *UAPool
	acceptLangs []string
	referers    []string
}

// NewHeaderBuilder 创建请求头构建器
func NewHeaderBuilder(uaPool *UAPool) *HeaderBuilder {
	return &HeaderBuilder{
		uaPool: uaPool,
		acceptLangs: []string{
			"zh-CN,zh;q=0.9,en;q=0.8",
			"zh-CN,zh;q=0.9",
			"zh-CN,zh;q=0.9,en;q=0.8,ja;q=0.7",
			"en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
		},
		referers: []string{
			"https://www.goofish.com/",
			"https://www.goofish.com/item/",
			"https://www.taobao.com/",
			"https://h5.m.goofish.com/",
		},
	}
}

// BuildRandomHeaders 构建随机请求头
func (h *HeaderBuilder) BuildRandomHeaders() map[string]string {
	headers := make(map[string]string)
	rand.Seed(time.Now().UnixNano())

	// User-Agent
	headers["User-Agent"] = h.uaPool.GetRandomUA()

	// Accept-Language
	idx := rand.Intn(len(h.acceptLangs))
	headers["Accept-Language"] = h.acceptLangs[idx]

	// 注意：不设置 Accept-Encoding，让 Go 的 http.Client 自动处理 gzip 解压

	// Referer
	idx = rand.Intn(len(h.referers))
	headers["Referer"] = h.referers[idx]

	return headers
}
