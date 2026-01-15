package mtop

import (
	"testing"
	"time"
)

func TestUAPool_GetRandomUA(t *testing.T) {
	ua := globalUAPool.GetRandomUA()
	if ua == "" {
		t.Error("GetRandomUA 返回空字符串")
	}

	// 验证返回的 UA 在预定义列表中
	found := false
	for _, expected := range globalUAPool.uas {
		if ua == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("返回的 UA 不在预定义列表中: %s", ua)
	}
}

func TestDelayManager_Wait(t *testing.T) {
	tests := []struct {
		name     string
		minMs    int
		maxMs    int
		minDelay time.Duration
		maxDelay time.Duration
	}{
		{
			name:     "正常延迟范围",
			minMs:    100,
			maxMs:    200,
			minDelay: 100 * time.Millisecond,
			maxDelay: 250 * time.Millisecond,
		},
		{
			name:     "零延迟",
			minMs:    0,
			maxMs:    0,
			minDelay: 0,
			maxDelay: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := NewDelayManager(tt.minMs, tt.maxMs)

			start := time.Now()
			dm.Wait()
			elapsed := time.Since(start)

			if elapsed < tt.minDelay {
				t.Errorf("延迟时间过短: %v < %v", elapsed, tt.minDelay)
			}
			if elapsed > tt.maxDelay {
				t.Errorf("延迟时间过长: %v > %v", elapsed, tt.maxDelay)
			}
		})
	}
}

func TestHeaderBuilder_BuildRandomHeaders(t *testing.T) {
	hb := NewHeaderBuilder(globalUAPool)
	headers := hb.BuildRandomHeaders()

	// 验证必需的请求头存在
	requiredHeaders := []string{"User-Agent", "Accept-Language", "Referer"}
	for _, h := range requiredHeaders {
		if _, ok := headers[h]; !ok {
			t.Errorf("缺少必需的请求头: %s", h)
		}
	}

	// 验证 User-Agent 不为空
	if headers["User-Agent"] == "" {
		t.Error("User-Agent 不应为空")
	}

	// 验证 Accept-Language 在预定义列表中
	validLang := false
	for _, lang := range hb.acceptLangs {
		if headers["Accept-Language"] == lang {
			validLang = true
			break
		}
	}
	if !validLang {
		t.Errorf("Accept-Language 不在预定义列表中: %s", headers["Accept-Language"])
	}

	// 验证 Referer 在预定义列表中
	validReferer := false
	for _, ref := range hb.referers {
		if headers["Referer"] == ref {
			validReferer = true
			break
		}
	}
	if !validReferer {
		t.Errorf("Referer 不在预定义列表中: %s", headers["Referer"])
	}
}

func TestNewDelayManager(t *testing.T) {
	dm := NewDelayManager(1000, 3000)

	if dm == nil {
		t.Fatal("NewDelayManager 返回 nil")
	}

	if dm.minMs != 1000 {
		t.Errorf("minMs 设置错误: %d", dm.minMs)
	}

	if dm.maxMs != 3000 {
		t.Errorf("maxMs 设置错误: %d", dm.maxMs)
	}
}

func TestNewHeaderBuilder(t *testing.T) {
	hb := NewHeaderBuilder(globalUAPool)

	if hb == nil {
		t.Fatal("NewHeaderBuilder 返回 nil")
	}

	if hb.uaPool == nil {
		t.Error("uaPool 未初始化")
	}

	if len(hb.acceptLangs) == 0 {
		t.Error("acceptLangs 为空")
	}

	if len(hb.referers) == 0 {
		t.Error("referers 为空")
	}
}
