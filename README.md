# 校园服务助手

校园生活服务平台小程序，提供跑腿、拼车、二手交易、失物招领等服务。

## 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 前端 | Taro + React + TypeScript | 3.6+ |
| 后端 | Golang + Gin + GORM | 1.21+ |
| 数据库 | MySQL | 8.0 |
| 缓存 | Redis | 7.x |
| 容器 | Docker + Compose | 最新 |

## 项目结构

```
campus-service/
├── backend/                # 后端服务
│   ├── cmd/server/         # 入口
│   ├── internal/
│   │   ├── config/         # 配置
│   │   ├── handler/        # HTTP处理
│   │   ├── middleware/     # 中间件
│   │   ├── model/          # 数据模型
│   │   ├── repository/     # 数据访问
│   │   ├── service/        # 业务逻辑
│   │   └── pkg/            # 工具包
│   ├── config.yaml         # 配置文件
│   └── Dockerfile
├── frontend/               # 前端小程序
│   ├── src/
│   │   ├── pages/          # 页面
│   │   ├── components/     # 组件
│   │   ├── services/       # API服务
│   │   ├── stores/         # 状态管理
│   │   └── utils/          # 工具
│   └── config/             # Taro配置
├── docker/                 # Docker配置
│   └── mysql/init.sql      # 数据库初始化
├── specs/                  # 规格文档
│   ├── PRD.md              # 产品需求文档
│   └── TECHNICAL_DESIGN.md # 技术方案
├── docker-compose.yml
├── .env.example
└── README.md
```

## 快速开始

### 环境要求

- Docker & Docker Compose
- Node.js 18+ (前端开发)
- Go 1.21+ (后端开发)

### 1. 克隆项目

```bash
git clone <repository-url>
cd campus-service
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置微信小程序 AppID 和 AppSecret
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend
```

### 4. 访问服务

- 后端 API: http://localhost:8080
- 健康检查: http://localhost:8080/health

### 5. 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm run dev:weapp

# 构建
npm run build:weapp
```

## API 接口

### 认证模块

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/auth/login | 微信登录 |
| GET | /api/v1/auth/profile | 获取用户信息 |
| PUT | /api/v1/auth/profile | 更新用户信息 |
| POST | /api/v1/auth/token/refresh | 刷新Token |

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误码

| 错误码 | 描述 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器错误 |

## 开发指南

### 后端开发

```bash
cd backend

# 安装依赖
go mod download

# 运行
go run cmd/server/main.go

# 构建
go build -o campus-server cmd/server/main.go
```

### 数据库迁移

数据库表会在服务启动时自动创建。

### 配置说明

编辑 `backend/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 3306
  user: root
  password: root123
  name: campus_service

redis:
  host: localhost
  port: 6379

wechat:
  appid: your_appid
  secret: your_secret

jwt:
  secret: your_jwt_secret
  access_token_expiry: 2h
  refresh_token_expiry: 168h
```

## 功能模块

- [x] 用户鉴权 (微信登录)
- [ ] 跑腿服务
- [ ] 拼车服务
- [ ] 二手交易
- [ ] 失物招领

## License

MIT