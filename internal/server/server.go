package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"xianyu_aner/internal/config"
	"xianyu_aner/internal/server/handlers"
	"xianyu_aner/pkg/feishu"
	"xianyu_aner/pkg/mtop"
)

// Server HTTP服务器
type Server struct {
	engine       *gin.Engine
	config       config.Config
	mtopClient   *mtop.Client
	feishuClient *feishu.Client
	feishuConfig *feishu.BitableConfig
	httpServer   *http.Server
}

// New 创建新的服务器
func New(cfg config.Config) *Server {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	s := &Server{
		engine: gin.New(),
		config: cfg,
	}

	s.initializeClients()
	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// initializeClients 初始化客户端
func (s *Server) initializeClients() {
	// 创建MTOP客户端
	s.mtopClient = mtop.NewClient(s.config.MTOP.Token, "34839810",
		mtop.WithCookies(s.config.MTOP.Cookies),
	)

	// 创建飞书客户端（如果配置了）
	if s.config.Feishu.Enabled && s.config.Feishu.AppID != "" && s.config.Feishu.AppSecret != "" {
		s.feishuClient = feishu.NewClient(feishu.ClientConfig{
			AppID:     s.config.Feishu.AppID,
			AppSecret: s.config.Feishu.AppSecret,
		})
		s.feishuConfig = &feishu.BitableConfig{
			AppToken:   s.config.Feishu.AppToken,
			TableToken: s.config.Feishu.TableToken,
		}
	}
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	s.engine.Use(cors.Default())
	s.engine.Use(gin.Recovery())
	s.engine.Use(gin.Logger())
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 创建handlers
	feedHandler := handlers.NewFeedHandler(s.mtopClient)
	healthHandler := handlers.NewHealthHandler()
	feishuHandler := handlers.NewFeishuHandler(s.feishuClient, s.feishuConfig)

	// API v1路由组
	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.HandleHealth)
		v1.GET("/feed", feedHandler.HandleFeed)
		v1.POST("/feishu/push", feishuHandler.HandleFeishuPush)
	}

	// 根路径
	s.engine.GET("/", s.handleRoot)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      s.engine,
		ReadTimeout:  s.config.Server.GetTimeout(),
		WriteTimeout: s.config.Server.GetTimeout(),
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 API服务器启动在 http://localhost:%d", s.config.Server.Port)
	log.Println("📋 可用的接口:")
	log.Println("   GET  /api/v1/health      - 健康检查")
	log.Println("   GET  /api/v1/feed        - 获取猜你喜欢")
	log.Println("   POST /api/v1/feishu/push - 推送到飞书表格")
	log.Println("   GET  /                   - API文档")

	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// GetMtopClient 获取MTOP客户端
func (s *Server) GetMtopClient() *mtop.Client {
	return s.mtopClient
}

// GetFeishuClient 获取飞书客户端
func (s *Server) GetFeishuClient() (*feishu.Client, *feishu.BitableConfig) {
	return s.feishuClient, s.feishuConfig
}

// GetConfig 获取配置
func (s *Server) GetConfig() config.Config {
	return s.config
}

// handleRoot 根路径，返回API文档
func (s *Server) handleRoot(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html>
<head>
    <title>闲鱼 API 服务</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #ff6000; margin-bottom: 10px; }
        .version { color: #999; font-size: 14px; margin-bottom: 30px; }
        .endpoint { background: #f8f9fa; padding: 20px; margin: 15px 0; border-radius: 8px; border-left: 4px solid #ff6000; }
        .method { display: inline-block; padding: 5px 12px; border-radius: 4px; color: white; font-weight: bold; font-size: 12px; margin-right: 10px; }
        .get { background: #28a745; }
        .post { background: #007bff; }
        .path { font-weight: bold; font-size: 16px; }
        .desc { margin-top: 10px; color: #666; }
        .params { margin-top: 15px; background: white; padding: 15px; border-radius: 5px; }
        .params strong { color: #333; }
        code { background: #e8e8e8; padding: 3px 8px; border-radius: 4px; font-family: 'Courier New', monospace; color: #d63384; }
        .example { margin-top: 10px; padding: 10px; background: #fff3cd; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>闲鱼 API 服务</h1>
        <p class="version">Version 1.0</p>

        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/health</span>
            <div class="desc">健康检查接口，用于验证服务是否正常运行</div>
            <div class="example">示例: curl http://localhost:8080/api/v1/health</div>
        </div>

        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="path">/api/v1/feed</span>
            <div class="desc">获取猜你喜欢商品列表</div>
            <div class="params">
                <strong>请求参数:</strong><br><br>
                <code>pages</code>: 爬取页数，默认 1，范围 1-10<br>
                <code>machId</code>: 推荐码/机器ID，可选<br><br>
                <strong>示例:</strong><br>
                <code>curl http://localhost:8080/api/v1/feed?pages=3</code><br>
                <code>curl http://localhost:8080/api/v1/feed?pages=2&machId=xxx</code>
            </div>
        </div>

        <div class="endpoint">
            <span class="method post">POST</span>
            <span class="path">/api/v1/feishu/push</span>
            <div class="desc">推送商品数据到飞书多维表格</div>
            <div class="params">
                <strong>请求参数 (JSON):</strong><br><br>
                <code>date</code>: 日期 (必需)<br>
                <code>products</code>: 商品列表 (必需)<br>
                <code>appToken</code>: 飞书应用 token (可选，覆盖默认配置)<br>
                <code>tableToken</code>: 飞书数据表 token (可选，覆盖默认配置)<br><br>
                <strong>示例:</strong><br>
                <code>curl -X POST http://localhost:8080/api/v1/feishu/push \<br>&nbsp;&nbsp;-H "Content-Type: application/json" \<br>&nbsp;&nbsp;-d '{"date":"2024-01-15","products":[...]}'</code>
            </div>
        </div>
    </div>
</body>
</html>`
	c.String(http.StatusOK, html)
}
