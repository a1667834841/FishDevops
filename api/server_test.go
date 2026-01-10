package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	// 测试模式
	gin.SetMode(gin.TestMode)
}

// TestHandleHealth 测试健康检查接口
func TestHandleHealth(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:    8080,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	})

	// 设置路由
	server.SetupRoutes()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected application/json content type, got %s", contentType)
	}
}

// TestHandleRoot 测试根路径
func TestHandleRoot(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:    8080,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	})

	server.SetupRoutes()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected text/html content type, got %s", contentType)
	}
}

// TestHandleFeedInvalidPagesParam 测试无效的 pages 参数
func TestHandleFeedInvalidPagesParam(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:    8080,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	})

	server.SetupRoutes()

	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"负数", "/api/v1/feed?pages=-1", http.StatusBadRequest},
		{"零", "/api/v1/feed?pages=0", http.StatusInternalServerError}, // 0 通过验证，设置为默认值1，但API调用失败
		{"超过上限", "/api/v1/feed?pages=11", http.StatusBadRequest},
		{"非数字", "/api/v1/feed?pages=abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			server.engine.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

// TestHandleFeedValidParams 测试有效参数
func TestHandleFeedValidParams(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:    8080,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	})

	server.SetupRoutes()

	// 只测试参数验证，不测试实际API调用（需要真实token）
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"1页", "/api/v1/feed?pages=1", 500}, // 会返回500因为没有真实token
		{"最大值", "/api/v1/feed?pages=10", 500},
		{"带machId", "/api/v1/feed?pages=2&machId=test123", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			server.engine.ServeHTTP(w, req)

			if w.Code != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, w.Code)
			}
		})
	}
}

// TestHandleFeedDefaultPages 测试默认页数
func TestHandleFeedDefaultPages(t *testing.T) {
	server := NewServer(ServerConfig{
		Port:    8080,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	})

	server.SetupRoutes()

	req := httptest.NewRequest("GET", "/api/v1/feed", nil)
	w := httptest.NewRecorder()

	server.engine.ServeHTTP(w, req)

	// 应该返回500（没有真实token）但参数应该通过验证
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 (no token), got %d", w.Code)
	}
}

// TestNewServer 测试服务器创建
func TestNewServer(t *testing.T) {
	config := ServerConfig{
		Port:    9090,
		Token:   "test_token",
		Cookies: []*http.Cookie{},
		Mode:    gin.TestMode,
	}

	server := NewServer(config)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.port != 9090 {
		t.Errorf("Expected port 9090, got %d", server.port)
	}

	if server.token != "test_token" {
		t.Errorf("Expected token test_token, got %s", server.token)
	}

	if server.engine == nil {
		t.Error("Engine is nil")
	}
}

// TestServerConfigDefaults 测试配置默认值
func TestServerConfigDefaults(t *testing.T) {
	server := NewServer(ServerConfig{
		Token:   "test_token",
		Cookies: []*http.Cookie{},
	})

	if server.port != 8080 {
		t.Errorf("Expected default port 8080, got %d", server.port)
	}
}
