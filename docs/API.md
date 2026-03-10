# 校园服务助手 API 文档

## 概述

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: Bearer Token (JWT)
- **数据格式**: JSON
- **编码**: UTF-8

## 通用响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "错误信息",
  "data": null
}
```

### 分页响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 认证

### 微信登录
```
POST /auth/wechat-login
```

**请求体**:
```json
{
  "code": "微信登录code"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "jwt_token",
    "user": {
      "id": 1,
      "nickname": "用户昵称",
      "avatar": "头像URL",
      "phone": "手机号"
    }
  }
}
```

### 获取当前用户信息
```
GET /users/me
Authorization: Bearer <token>
```

### 更新用户信息
```
PUT /users/me
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "nickname": "新昵称",
  "avatar": "新头像URL",
  "phone": "手机号"
}
```

---

## 跑腿服务

### 发布跑腿任务
```
POST /errands
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "title": "取快递",
  "description": "帮忙取一下菜鸟驿站的快递",
  "type": "pickup",
  "pickup_location": "菜鸟驿站",
  "delivery_location": "宿舍楼A栋",
  "reward": 5.0,
  "deadline": "2026-03-10T18:00:00Z"
}
```

**任务类型**:
- `pickup` - 取快递
- `delivery` - 送东西
- `buy` - 代买
- `other` - 其他

### 获取跑腿任务列表
```
GET /errands?page=1&page_size=10&type=pickup&status=open&keyword=快递
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码，默认1 |
| page_size | int | 每页数量，默认10 |
| type | string | 任务类型 |
| status | string | 状态: open, accepted, completed, cancelled |
| keyword | string | 搜索关键词 |
| sort_by | string | 排序: reward, reward_desc, time, time_desc |

### 获取跑腿任务详情
```
GET /errands/:id
```

### 更新跑腿任务
```
PUT /errands/:id
Authorization: Bearer <token>
```

### 删除跑腿任务
```
DELETE /errands/:id
Authorization: Bearer <token>
```

### 接单
```
POST /errands/:id/accept
Authorization: Bearer <token>
```

### 取消任务
```
POST /errands/:id/cancel
Authorization: Bearer <token>
```

### 完成任务
```
POST /errands/:id/complete
Authorization: Bearer <token>
```

### 我发布的任务
```
GET /errands/my/published
Authorization: Bearer <token>
```

### 我接的任务
```
GET /errands/my/accepted
Authorization: Bearer <token>
```

---

## 拼车服务

### 发布拼车行程
```
POST /carpools
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "origin": "学校北门",
  "destination": "火车站",
  "departure_time": "2026-03-10T15:00:00Z",
  "total_seats": 4,
  "available_seats": 3,
  "price": 15.0,
  "vehicle_info": "白色大众",
  "remark": "准时出发"
}
```

### 获取拼车列表
```
GET /carpools?page=1&page_size=10&status=open
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| origin | string | 出发地关键词 |
| destination | string | 目的地关键词 |
| status | string | 状态: open, full, departed, cancelled |
| date | string | 日期筛选 |

### 获取拼车详情
```
GET /carpools/:id
```

### 更新拼车行程
```
PUT /carpools/:id
Authorization: Bearer <token>
```

### 删除拼车行程
```
DELETE /carpools/:id
Authorization: Bearer <token>
```

### 预约座位
```
POST /carpools/:id/book
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "seats": 1,
  "remark": "备注"
}
```

### 确认出发
```
POST /carpools/:id/depart
Authorization: Bearer <token>
```

### 取消行程
```
POST /carpools/:id/cancel
Authorization: Bearer <token>
```

### 我发布的行程
```
GET /carpools/my/published
Authorization: Bearer <token>
```

### 我的预约
```
GET /carpools/my/bookings
Authorization: Bearer <token>
```

### 取消预约
```
POST /carpool-bookings/:id/cancel
Authorization: Bearer <token>
```

---

## 二手交易

### 发布商品
```
POST /market/items
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "title": "二手自行车",
  "description": "九成新，骑了半年",
  "category": "sports",
  "price": 200.0,
  "original_price": 500.0,
  "condition": "like_new",
  "images": ["url1", "url2"],
  "location": "宿舍楼下交易",
  "contact": "微信: xxx"
}
```

**商品分类**:
- `electronics` - 电子产品
- `books` - 书籍
- `clothes` - 服饰
- `sports` - 运动用品
- `daily` - 日用品
- `others` - 其他

**商品成色**:
- `new` - 全新
- `like_new` - 几乎全新
- `good` - 良好
- `fair` - 一般

### 获取商品列表
```
GET /market/items?page=1&page_size=10&category=electronics&min_price=0&max_price=1000
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| category | string | 分类 |
| status | string | 状态: on_sale, sold, off_shelf |
| keyword | string | 搜索关键词 |
| min_price | float | 最低价格 |
| max_price | float | 最高价格 |
| sort_by | string | 排序: price, price_desc, time, time_desc |

### 获取商品详情
```
GET /market/items/:id
```

### 更新商品
```
PUT /market/items/:id
Authorization: Bearer <token>
```

### 删除商品
```
DELETE /market/items/:id
Authorization: Bearer <token>
```

### 标记已售
```
POST /market/items/:id/sold
Authorization: Bearer <token>
```

### 下架商品
```
POST /market/items/:id/off-shelf
Authorization: Bearer <token>
```

### 重新上架
```
POST /market/items/:id/on-shelf
Authorization: Bearer <token>
```

### 收藏商品
```
POST /market/items/:id/favorite
Authorization: Bearer <token>
```

### 取消收藏
```
DELETE /market/items/:id/favorite
Authorization: Bearer <token>
```

### 我的商品
```
GET /market/my/items
Authorization: Bearer <token>
```

### 我的收藏
```
GET /market/my/favorites
Authorization: Bearer <token>
```

---

## 失物招领

### 发布失物招领
```
POST /lost-found
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "type": "lost",
  "title": "丢失校园卡",
  "description": "在图书馆丢失一张校园卡",
  "category": "cards",
  "images": ["url1"],
  "location": "图书馆",
  "lost_time": "2026-03-10T10:00:00Z",
  "contact": "电话: 138xxxx"
}
```

**类型**:
- `lost` - 寻物启事
- `found` - 失物招领

**分类**:
- `electronics` - 电子产品
- `cards` - 证件卡片
- `books` - 书籍文具
- `clothes` - 服饰配件
- `bags` - 包袋箱包
- `others` - 其他

### 获取失物招领列表
```
GET /lost-found?page=1&page_size=10&type=lost&category=cards
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| page_size | int | 每页数量 |
| type | string | 类型: lost, found |
| category | string | 分类 |
| status | string | 状态: open, closed, resolved |
| keyword | string | 搜索关键词 |

### 获取失物招领详情
```
GET /lost-found/:id
```

### 更新失物招领
```
PUT /lost-found/:id
Authorization: Bearer <token>
```

### 删除失物招领
```
DELETE /lost-found/:id
Authorization: Bearer <token>
```

### 关闭失物招领
```
POST /lost-found/:id/close
Authorization: Bearer <token>
```

### 标记已解决
```
POST /lost-found/:id/resolve
Authorization: Bearer <token>
```

### 提交认领申请
```
POST /lost-found/:id/claim
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "message": "这是我的卡，卡号后四位是1234",
  "proof": ["证明图片URL"]
}
```

### 获取认领列表
```
GET /lost-found/:id/claims
Authorization: Bearer <token>
```

### 通过认领
```
POST /lost-found/claims/:claim_id/approve
Authorization: Bearer <token>
```

### 拒绝认领
```
POST /lost-found/claims/:claim_id/reject
Authorization: Bearer <token>
```

### 我的发布
```
GET /lost-found/my/published
Authorization: Bearer <token>
```

### 我的认领
```
GET /lost-found/my/claims
Authorization: Bearer <token>
```

---

## 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/Token无效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 数据模型

### User 用户
```json
{
  "id": 1,
  "nickname": "用户昵称",
  "avatar": "头像URL",
  "phone": "手机号",
  "created_at": "2026-03-10T10:00:00Z"
}
```

### Errand 跑腿任务
```json
{
  "id": 1,
  "user_id": 1,
  "title": "取快递",
  "description": "描述",
  "type": "pickup",
  "pickup_location": "取件地点",
  "delivery_location": "送达地点",
  "reward": 5.0,
  "deadline": "2026-03-10T18:00:00Z",
  "status": "open",
  "view_count": 10,
  "created_at": "2026-03-10T10:00:00Z",
  "user": { ... }
}
```

### Carpool 拼车行程
```json
{
  "id": 1,
  "user_id": 1,
  "origin": "出发地",
  "destination": "目的地",
  "departure_time": "2026-03-10T15:00:00Z",
  "total_seats": 4,
  "available_seats": 3,
  "price": 15.0,
  "vehicle_info": "车辆信息",
  "status": "open",
  "created_at": "2026-03-10T10:00:00Z",
  "user": { ... }
}
```

### MarketItem 二手商品
```json
{
  "id": 1,
  "user_id": 1,
  "title": "商品名称",
  "description": "描述",
  "category": "electronics",
  "price": 200.0,
  "original_price": 500.0,
  "condition": "like_new",
  "images": ["url1", "url2"],
  "location": "交易地点",
  "contact": "联系方式",
  "status": "on_sale",
  "view_count": 20,
  "created_at": "2026-03-10T10:00:00Z",
  "user": { ... }
}
```

### LostFound 失物招领
```json
{
  "id": 1,
  "user_id": 1,
  "type": "lost",
  "title": "标题",
  "description": "描述",
  "category": "cards",
  "images": ["url1"],
  "location": "地点",
  "lost_time": "2026-03-10T10:00:00Z",
  "contact": "联系方式",
  "status": "open",
  "view_count": 30,
  "created_at": "2026-03-10T10:00:00Z",
  "user": { ... }
}
```

---

## 开发环境

### 启动服务
```bash
cd /root/projects/campus-service
docker-compose up -d
```

### 服务端口
| 服务 | 端口 |
|------|------|
| Backend | 8080 |
| MySQL | 3306 |
| Redis | 6379 |

### 健康检查
```
GET /health
```

---

*文档版本: 1.0.0*
*最后更新: 2026-03-10*