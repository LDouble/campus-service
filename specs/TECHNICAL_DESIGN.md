# 校园服务助手 - 技术方案

## 1. 技术选型

### 1.1 技术栈总览

| 层级 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 前端框架 | Taro + React + TypeScript | 3.6+ | 跨平台小程序框架 |
| 状态管理 | Zustand | 4.x | 轻量级状态管理 |
| HTTP 客户端 | Taro.request | - | 小程序网络请求 |
| 后端框架 | Golang + Gin | 1.21+ / 1.9+ | 高性能 Web 框架 |
| ORM | GORM | 1.25+ | Go ORM 库 |
| 数据库 | MySQL | 8.0 | 主数据存储 |
| 缓存 | Redis | 7.x | 会话/缓存/限流 |
| 容器 | Docker + Compose | 最新 | 容器化部署 |
| 配置管理 | Viper | 1.x | 配置管理 |
| 日志 | Zap | 1.x | 高性能日志 |
| 认证 | JWT | - | Token 认证 |

### 1.2 选型理由

**前端 - Taro**
- 一套代码多端运行（微信、支付宝、H5）
- React 生态，开发效率高
- TypeScript 支持，类型安全

**后端 - Golang + Gin**
- 高性能，适合高并发场景
- 部署简单，单一二进制文件
- Gin 框架轻量高效

**数据库 - MySQL**
- 成熟稳定，生态完善
- 支持事务，数据一致性有保障
- 运维成本低

**缓存 - Redis**
- 高性能缓存
- 支持会话存储、限流
- 数据结构丰富

---

## 2. 系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      微信小程序                              │
│                    ┌───────────┐                            │
│                    │   Taro    │                            │
│                    │  React    │                            │
│                    │ TypeScript│                            │
│                    └─────┬─────┘                            │
└──────────────────────────┼──────────────────────────────────┘
                           │ HTTPS
┌──────────────────────────▼──────────────────────────────────┐
│                      API Gateway                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Gin Server                                          │   │
│  │  - CORS 中间件                                       │   │
│  │  - 请求日志                                          │   │
│  │  - 限流中间件                                        │   │
│  │  - JWT 认证中间件                                    │   │
│  │  - 错误恢复中间件                                     │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Application Layer                          │
│                                                              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐│
│  │  Auth   │ │ Errand  │ │ Carpool │ │ Market  │ │  Lost  ││
│  │ Handler │ │ Handler │ │ Handler │ │ Handler │ │ Handler││
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └───┬────┘│
│       │           │           │           │           │      │
│  ┌────▼───────────▼───────────▼───────────▼───────────▼────┐│
│  │                    Service Layer                        ││
│  │  - 业务逻辑处理                                          ││
│  │  - 事务管理                                              ││
│  │  - 缓存策略                                              ││
│  └────────────────────────────┬────────────────────────────┘│
└───────────────────────────────┼──────────────────────────────┘
                                │
┌───────────────────────────────▼──────────────────────────────┐
│                       Data Layer                              │
│                                                               │
│  ┌─────────────────────┐       ┌─────────────────────┐      │
│  │       MySQL         │       │       Redis         │      │
│  │  ┌───────────────┐   │       │  ┌───────────────┐  │      │
│  │  │ users         │   │       │  │ session:*     │  │      │
│  │  │ errands       │   │       │  │ token:black   │  │      │
│  │  │ errand_orders │   │       │  │ cache:*       │  │      │
│  │  │ carpools      │   │       │  │ rate:*        │  │      │
│  │  │ carpool_bookings│  │       │  └───────────────┘  │      │
│  │  │ products      │   │       └─────────────────────┘      │
│  │  │ lost_founds   │   │                                      │
│  │  │ messages      │   │                                      │
│  │  │ comments      │   │                                      │
│  │  └───────────────┘   │                                      │
│  └─────────────────────┘                                      │
└───────────────────────────────────────────────────────────────┘
```

### 2.2 请求处理流程

```
Request → Middleware → Handler → Service → Repository → DB
                ↓
            Response
```

---

## 3. 项目结构

### 3.1 后端目录结构

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # 入口文件
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── handler/
│   │   ├── auth.go              # 认证处理器
│   │   ├── errand.go            # 跑腿处理器
│   │   ├── carpool.go           # 拼车处理器
│   │   ├── market.go            # 二手处理器
│   │   └── lostfound.go         # 失物招领处理器
│   ├── middleware/
│   │   ├── auth.go              # JWT 认证中间件
│   │   ├── cors.go              # CORS 中间件
│   │   ├── logger.go            # 日志中间件
│   │   ├── recovery.go          # 错误恢复中间件
│   │   └── ratelimit.go         # 限流中间件
│   ├── model/
│   │   ├── user.go              # 用户模型
│   │   ├── errand.go            # 跑腿模型
│   │   ├── carpool.go           # 拼车模型
│   │   ├── product.go           # 商品模型
│   │   └── lostfound.go         # 失物招领模型
│   ├── repository/
│   │   ├── user.go              # 用户数据访问
│   │   ├── errand.go            # 跑腿数据访问
│   │   ├── carpool.go           # 拼车数据访问
│   │   ├── product.go           # 商品数据访问
│   │   └── lostfound.go         # 失物招领数据访问
│   ├── service/
│   │   ├── auth.go              # 认证服务
│   │   ├── errand.go            # 跑腿服务
│   │   ├── carpool.go           # 拼车服务
│   │   ├── market.go            # 二手服务
│   │   └── lostfound.go         # 失物招领服务
│   └── pkg/
│       ├── response/
│       │   └── response.go      # 统一响应
│       ├── jwt/
│       │   └── jwt.go           # JWT 工具
│       ├── wechat/
│       │   └── wechat.go        # 微信 API
│       └── utils/
│           └── utils.go         # 通用工具
├── go.mod
├── go.sum
└── Dockerfile
```

### 3.2 前端目录结构

```
frontend/
├── src/
│   ├── app.config.ts            # 应用配置
│   ├── app.tsx                  # 入口文件
│   ├── index.html               # H5 入口
│   ├── pages/
│   │   ├── index/               # 首页
│   │   │   ├── index.tsx
│   │   │   └── index.module.scss
│   │   ├── login/               # 登录页
│   │   ├── errand/              # 跑腿
│   │   │   ├── index/           # 列表页
│   │   │   ├── detail/          # 详情页
│   │   │   └── publish/         # 发布页
│   │   ├── carpool/             # 拼车
│   │   │   ├── index/
│   │   │   ├── detail/
│   │   │   └── publish/
│   │   ├── market/              # 二手
│   │   │   ├── index/
│   │   │   ├── detail/
│   │   │   └── publish/
│   │   ├── lostfound/           # 失物招领
│   │   │   ├── index/
│   │   │   ├── detail/
│   │   │   └── publish/
│   │   └── profile/             # 个人中心
│   │       ├── index/
│   │       └── edit/
│   ├── components/              # 公共组件
│   │   ├── Navbar/              # 导航栏
│   │   ├── Tabbar/              # 底部标签栏
│   │   ├── Card/                # 卡片组件
│   │   ├── Empty/               # 空状态
│   │   ├── Loading/             # 加载状态
│   │   └── Modal/               # 弹窗
│   ├── services/                # API 服务
│   │   ├── request.ts           # 请求封装
│   │   ├── auth.ts              # 认证 API
│   │   ├── errand.ts            # 跑腿 API
│   │   ├── carpool.ts           # 拼车 API
│   │   ├── market.ts            # 二手 API
│   │   └── lostfound.ts         # 失物招领 API
│   ├── stores/                  # 状态管理
│   │   ├── user.ts              # 用户状态
│   │   └── common.ts            # 公共状态
│   ├── utils/                   # 工具函数
│   │   ├── storage.ts           # 本地存储
│   │   ├── format.ts            # 格式化
│   │   └── validate.ts          # 校验
│   └── types/                   # 类型定义
│       └── index.ts
├── config/
│   ├── dev.ts                   # 开发配置
│   ├── prod.ts                  # 生产配置
│   └── index.ts                 # 配置入口
├── package.json
└── project.config.json          # 小程序配置
```

---

## 4. 数据库设计

### 4.1 ER 图

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    users    │       │   errands   │       │errand_orders│
│─────────────│       │─────────────│       │─────────────│
│ id          │──┐    │ id          │──┐    │ id          │
│ openid      │  │    │ user_id     │◄─┘    │ errand_id   │
│ nickname    │  │    │ title       │       │ user_id     │
│ avatar      │  │    │ type        │       │ status      │
│ phone       │  │    │ description │       │ created_at  │
│ status      │  │    │ reward      │       └─────────────┘
│ created_at  │  │    │ location    │
└─────────────┘  │    │ status      │
                 │    │ created_at  │
                 │    └─────────────┘
                 │
                 │    ┌─────────────┐       ┌─────────────┐
                 │    │  carpools   │       │carpool_books│
                 │    │─────────────│       │─────────────│
                 └───►│ id          │──┐    │ id          │
                      │ user_id     │◄─┘    │ carpool_id  │
                      │ origin      │       │ user_id     │
                      │ destination │       │ seats       │
                      │ departure   │       │ status      │
                      │ seats       │       │ created_at  │
                      │ price       │       └─────────────┘
                      │ status      │
                      │ created_at  │
                      └─────────────┘

┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│  products   │       │  favorites  │       │  messages   │
│─────────────│       │─────────────│       │─────────────│
│ id          │──┐    │ id          │       │ id          │
│ user_id     │◄─┘    │ user_id     │       │ from_id     │
│ title       │       │ product_id  │       │ to_id       │
│ description │       │ created_at  │       │ content     │
│ price       │       └─────────────┘       │ type        │
│ images      │                             │ is_read     │
│ category    │                             │ created_at  │
│ status      │                             └─────────────┘
│ created_at  │
└─────────────┘

┌─────────────┐       ┌─────────────┐
│ lost_founds │       │  comments   │
│─────────────│       │─────────────│
│ id          │       │ id          │
│ user_id     │       │ user_id     │
│ type        │       │ target_type │
│ title       │       │ target_id   │
│ description │       │ content     │
│ images      │       │ rating      │
│ location    │       │ created_at  │
│ status      │       └─────────────┘
│ created_at  │
└─────────────┘
```

### 4.2 表结构详细设计

详见 `docs/DATABASE_DESIGN.md`

---

## 5. API 设计规范

### 5.1 RESTful 规范

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/resources | 获取列表 |
| GET | /api/v1/resources/:id | 获取详情 |
| POST | /api/v1/resources | 创建资源 |
| PUT | /api/v1/resources/:id | 更新资源 |
| DELETE | /api/v1/resources/:id | 删除资源 |

### 5.2 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1709900000
}
```

### 5.3 错误码定义

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未授权 |
| 1003 | 禁止访问 |
| 1004 | 资源不存在 |
| 2001 | 用户不存在 |
| 2002 | 登录失败 |
| 3001 | 任务不存在 |
| 3002 | 任务已接单 |
| 4001 | 商品不存在 |
| 4002 | 商品已售出 |

详见 `specs/API_SPEC.md`

---

## 6. 安全设计

### 6.1 认证流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  小程序      │     │   后端API   │     │   微信API   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       │  1. wx.login()    │                   │
       │  获取 code         │                   │
       │                   │                   │
       │  2. POST /auth/login                  │
       │  { code }         │                   │
       │                   │                   │
       │                   │  3. code换openid  │
       │                   │ ─────────────────>│
       │                   │                   │
       │                   │  4. 返回 openid   │
       │                   │ <─────────────────│
       │                   │                   │
       │                   │  5. 查询/创建用户 │
       │                   │  生成 JWT Token   │
       │                   │                   │
       │  6. 返回 token    │                   │
       │ <─────────────────│                   │
       │                   │                   │
       │  7. 存储 token    │                   │
       │  后续请求携带      │                   │
       │                   │                   │
```

### 6.2 安全措施

| 措施 | 说明 |
|------|------|
| HTTPS | 全站 HTTPS 加密传输 |
| JWT | Token 认证，设置合理过期时间 |
| 限流 | 接口限流，防止恶意请求 |
| 参数校验 | 严格校验所有输入参数 |
| SQL 注入防护 | 使用 ORM 参数化查询 |
| XSS 防护 | 输出转义，Content-Security-Policy |
| 敏感信息加密 | 手机号等敏感信息加密存储 |

---

## 7. 部署方案

### 7.1 Docker Compose 配置

```yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: campus_service
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  backend:
    build: ./backend
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis

volumes:
  mysql_data:
  redis_data:
```

### 7.2 环境变量

```env
# 服务配置
APP_PORT=8080
APP_ENV=development

# 数据库配置
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=campus_service

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT 配置
JWT_SECRET=your_jwt_secret
JWT_EXPIRE=86400

# 微信配置
WECHAT_APPID=your_appid
WECHAT_SECRET=your_secret
```

---

## 8. 开发规范

### 8.1 Git 分支策略

```
main (生产)
  └── develop (开发)
        ├── feature/auth (功能分支)
        ├── feature/errand
        ├── feature/carpool
        └── feature/market
```

### 8.2 Commit 规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式
refactor: 重构
test: 测试
chore: 构建/工具
```

### 8.3 代码规范

- Go: 遵循 Go 官方代码规范
- TypeScript: ESLint + Prettier
- 注释: 关键逻辑必须注释
- 测试: 核心功能需要单元测试