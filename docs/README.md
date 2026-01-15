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

## 使用方式

### 两种使用模式

本项目提供两种使用模式：
- **API 服务模式** - 启动 HTTP 服务，通过 API 接口调用
- **命令行模式** - 直接运行爬虫工具，适合一次性数据采集

---

### 模式一：API 服务

#### 1. 环境准备

```bash
# 克隆项目
git clone <repository-url>
cd xianyu_aner

# 安装 Go 依赖
go mod download

# 安装 Playwright 浏览器
go install github.com/playwright-community/playwright-go/cmd/playwright@latest
playwright install --with-deps chromium
```

#### 2. 配置服务

```bash
# 复制配置文件
cp configs/config.example.yaml config.yaml

# 编辑配置文件，填入必要信息
# 主要是飞书相关配置（如果需要推送到飞书）
```

#### 3. 启动服务

```bash
# 方式一：直接运行
go run main.go

# 方式二：编译后运行
go build -o xianyu_aner main.go
./xianyu_aner

# 方式三：使用生产入口
go run cmd/server/main.go
```

#### 4. 使用 API

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 获取猜你喜欢商品（默认1页，每页20条）
curl http://localhost:8080/api/v1/feed

# 获取多页数据
curl http://localhost:8080/api/v1/feed?pages=3

# 推送到飞书多维表格（需要先配置飞书参数）
curl -X POST http://localhost:8080/api/v1/feishu/push
```

#### 5. 高级用法

**使用环境变量配置：**

```bash
# 设置环境变量覆盖配置文件
export SERVER_PORT=9000
export BROWSER_HEADLESS=false
export FEISHU_ENABLED=true

# 启动服务
go run main.go
```

**Docker 部署（如果支持）：**

```bash
# 构建镜像
docker build -t xianyu_aner .

# 运行容器
docker run -p 8080:8080 \
  -e FEISHU_APP_ID=your_app_id \
  -e FEISHU_APP_SECRET=your_app_secret \
  xianyu_aner
```

### 飞书多维表格配置

如果需要使用飞书推送功能，需要完成以下配置：

1. **创建飞书应用**
   - 访问 [飞书开放平台](https://open.feishu.cn/)
   - 创建应用并获取 App ID 和 App Secret

2. **获取权限**
   - 开启以下权限：`bitable:app`、`bitable:app:readonly`

3. **配置多维表格**
   - 创建多维表格并记录 app_token 和 table_id
   - 在配置文件中填入相应信息

---

### 模式二：命令行爬虫

独立的爬虫工具，适合一次性数据采集和批处理任务。

#### 基本用法

```bash
# 编译爬虫工具
go build -o crawl cmd/crawl/main.go

# 运行爬虫（使用默认参数）
./crawl

# 或直接使用 go run
go run cmd/crawl/main.go
```

#### 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-configs` | string | - | 配置文件路径（可选） |
| `-pages` | int | 10 | 爬取页数 |
| `-min-want` | int | 1 | 最低想要人数过滤 |
| `-days` | int | 14 | 发布时间范围（天数） |
| `-output` | string | feed_result.json | 输出文件路径 |
| `-push-feishu` | bool | false | 是否推送到飞书 |
| `-headless` | bool | true | 是否使用无头浏览器 |
| `-version` | bool | false | 显示版本信息 |

#### 使用示例

```bash
# 基础爬取（默认10页）
go run cmd/crawl/main.go

# 爬取20页数据
go run cmd/crawl/main.go -pages=20

# 过滤想要人数>=10的商品，近7天发布
go run cmd/crawl/main.go -min-want=10 -days=7

# 爬取并推送到飞书
go run cmd/crawl/main.go -pages=5 -push-feishu

# 使用有头浏览器（可以看到登录过程）
go run cmd/crawl/main.go -headless=false

# 自定义输出文件
go run cmd/crawl/main.go -output=data.json

# 组合使用
go run cmd/crawl/main.go -pages=30 -min-want=5 -days=30 -push-feishu -output=result.json

# 查看版本
go run cmd/crawl/main.go -version
```

#### 执行流程

爬虫工具会按以下步骤执行：

```
========================================
  闲鱼数据爬取工具
========================================

[步骤 1/4] 获取登录 Cookie (无头模式: true)...
成功获取 Token: xxxxx...

[步骤 2/4] 爬取猜你喜欢数据 (页数: 10)...
爬取完成！获取到 200 条数据，耗时 15.23 秒

[步骤 3/4] 保存数据到文件: feed_result.json
成功保存 200 条数据

[步骤 4/4] 推送到飞书多维表格...
推送成功！创建记录数: 200

========================================
  任务完成统计
========================================
爬取商品数: 200
总耗时: 45.67 秒
========================================
```

#### 输出格式

爬取的数据会以 JSON 格式保存，包含以下字段：

```json
[
  {
    "item_id": "123456789",
    "title": "商品标题",
    "price": "¥99.00",
    "want_count": 25,
    "publish_time": "2024-01-15 10:30:00",
    "location": "上海",
    "image_url": "https://...",
    "tags": ["包邮", "正品"],
    ...
  }
]
```

## 更新日志

### [版本 0.x.x]

#### 最近更新
- feat(feishu): 完善商品详情接口支持并丰富推送数据字段
- fix(feishu): 优化字段创建逻辑并添加详细日志输出
- fix(mtop): 修复签名生成及卡片解析逻辑
- feat(feishu): 优化多维表格商品去重逻辑，精确查询商品ID记录

#### 已知问题
- 无

#### 计划中的功能
- [ ] 添加 Webhook 通知功能

---

### 历史版本

#### v0.1.0
- 初始版本发布
- 支持浏览器自动化获取 Cookie
- 实现闲鱼 MTOP API 调用
- 支持飞书多维表格数据推送
- 提供 RESTful API 接口

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
