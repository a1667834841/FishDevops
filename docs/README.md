# 闲鱼数据爬取服务

基于浏览器自动化和MTOP API的闲鱼商品数据爬取服务，支持飞书多维表格数据推送。

## 功能特性

- 🕷️ 浏览器自动化获取Cookie
- 📦 调用闲鱼MTOP API获取"猜你喜欢"商品
- 🌐 RESTful API服务
- 📊 飞书多维表格数据推送
- ⚙️ 灵活的配置管理（YAML + 环境变量）

## 快速开始

### 前置要求

- Go 1.23+
- Chromium浏览器（用于Playwright）

### 安装依赖

```bash
go mod download
go install github.com/playwright-community/playwright-go/cmd/playwright@latest
playwright install --with-deps chromium
```

### 配置

1. 复制配置文件示例：
```bash
cp configs/config.example.yaml config.yaml
```

2. 编辑配置文件或使用环境变量

### 运行

#### 本地开发
```bash
go run main.go
```

#### 流水线/生产环境
```bash
go run cmd/server/main.go
```

#### 编译后运行
```bash
go build -o xianyu_aner main.go
./xianyu_aner
```

## API文档

服务启动后访问 http://localhost:8080

### 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/feed` | 获取猜你喜欢商品 |
| POST | `/api/v1/feishu/push` | 推送到飞书表格 |

### 请求示例

```bash
# 获取猜你喜欢（默认1页）
curl http://localhost:8080/api/v1/feed

# 获取3页数据
curl http://localhost:8080/api/v1/feed?pages=3

# 健康检查
curl http://localhost:8080/api/v1/health
```

## 配置说明

### 配置文件

支持以下配置文件（按优先级）：
- `config.yaml`
- `configs/config.yaml`
- `config.local.yaml`
- `configs/config.local.yaml`

### 环境变量

环境变量会覆盖配置文件中的值：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `SERVER_PORT` | 服务端口 | 8080 |
| `SERVER_MODE` | 运行模式 | release |
| `BROWSER_HEADLESS` | 无头浏览器 | true |
| `FEISHU_ENABLED` | 启用飞书 | false |
| `FEISHU_APP_ID` | 飞书应用ID | - |
| `FEISHU_APP_SECRET` | 飞书密钥 | - |

## 项目结构

```
├── cmd/              # 流水线/生产入口
├── internal/         # 私有应用代码
│   ├── app/         # 应用启动逻辑
│   ├── config/      # 配置管理
│   ├── server/      # HTTP服务器
│   └── model/       # 数据模型
├── pkg/             # 可被外部使用的库
│   ├── mtop/        # 闲鱼MTOP客户端
│   └── feishu/      # 飞书API客户端
├── web/             # 静态资源
└── configs/         # 配置文件示例
```

## 开发指南

### 添加新功能

新增功能时只需修改 `internal/` 下的包，`cmd/server/main.go` 无需改动：

1. 在 `internal/server/handler/` 添加新的handler文件
2. 在 `internal/server/router.go` 注册新路由
3. 在 `internal/model/dto.go` 添加新的DTO（如果需要）

### 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...
```

## 许可证

MIT License
