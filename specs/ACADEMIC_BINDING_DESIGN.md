# 教务绑定功能技术设计

## 1. 架构设计

### 1.1 分层架构

```
Handler (HTTP处理)
    ↓
Service (业务逻辑)
    ↓
Repository (数据访问)
    ↓
Model (数据模型)
```

### 1.2 模块依赖

```
academic_binding
├── model/academic_binding.go
├── repository/academic_binding_repository.go
├── service/academic_binding_service.go
├── service/academic_verifier.go (接口)
├── service/mock_verifier.go (模拟实现)
├── handler/academic_binding_handler.go
└── middleware/binding_check.go (权限中间件)
```

## 2. 数据模型

### 2.1 model/academic_binding.go

```go
package model

import "time"

// AcademicBinding 教务绑定
type AcademicBinding struct {
    ID        uint64    `gorm:"primaryKey" json:"id"`
    UserID    uint64    `gorm:"uniqueIndex;not null" json:"user_id"`
    StudentID string    `gorm:"size:20;uniqueIndex;not null" json:"student_id"`
    RealName  string    `gorm:"size:50;not null" json:"real_name"`
    College   string    `gorm:"size:100" json:"college"`
    Major     string    `gorm:"size:100" json:"major"`
    Grade     string    `gorm:"size:10" json:"grade"`
    Status    int8      `gorm:"default:1" json:"status"` // 1-正常 0-已解绑
    BoundAt   time.Time `json:"bound_at"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (AcademicBinding) TableName() string {
    return "academic_bindings"
}

// BindingStatus 绑定状态响应
type BindingStatus struct {
    IsBound   bool      `json:"is_bound"`
    StudentID string    `json:"student_id,omitempty"`
    RealName  string    `json:"real_name,omitempty"`
    College   string    `json:"college,omitempty"`
    Major     string    `json:"major,omitempty"`
    Grade     string    `json:"grade,omitempty"`
    BoundAt   *time.Time `json:"bound_at,omitempty"`
}
```

## 3. 接口设计

### 3.1 AcademicVerifier 接口

```go
// service/academic_verifier.go
package service

// AcademicVerifier 教务验证器接口
type AcademicVerifier interface {
    // Verify 验证教务账号，返回学生信息
    Verify(studentID, password string) (*StudentInfo, error)
}

// StudentInfo 学生信息
type StudentInfo struct {
    StudentID string
    RealName  string
    College   string
    Major     string
    Grade     string
}
```

### 3.2 MockVerifier 模拟实现

```go
// service/mock_verifier.go
package service

import "errors"

// MockVerifier 模拟教务验证器（开发阶段）
type MockVerifier struct{}

func NewMockVerifier() *MockVerifier {
    return &MockVerifier{}
}

func (v *MockVerifier) Verify(studentID, password string) (*StudentInfo, error) {
    // 模拟验证：密码为 "123456" 视为成功
    if password != "123456" {
        return nil, errors.New("教务账号验证失败")
    }
    
    // 返回模拟数据
    return &StudentInfo{
        StudentID: studentID,
        RealName:  "测试用户",
        College:  "计算机学院",
        Major:    "软件工程",
        Grade:     studentID[:4], // 从学号提取年级
    }, nil
}
```

## 4. 业务逻辑

### 4.1 绑定流程

```
1. 用户提交学号 + 密码
2. 检查该学号是否已被绑定
3. 调用 AcademicVerifier.Verify() 验证
4. 验证成功，创建绑定记录
5. 缓存绑定状态到 Redis
6. 返回绑定信息
```

### 4.2 权限检查流程

```
1. 从 JWT 获取 user_id
2. 查询 Redis 缓存
3. 缓存未命中，查询数据库
4. 返回绑定状态
```

## 5. 中间件设计

### 5.1 BindingCheckMiddleware

```go
// middleware/binding_check.go
package middleware

// RequireBinding 要求已绑定教务
func RequireBinding() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := GetUserID(c)
        
        // 检查绑定状态
        isBound := checkBindingStatus(userID)
        
        if !isBound {
            c.JSON(403, gin.H{
                "code": 40001,
                "msg":  "请先绑定教务账号",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// OptionalBinding 可选绑定（用于脱敏控制）
func OptionalBinding() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := GetUserID(c)
        isBound := checkBindingStatus(userID)
        c.Set("is_bound", isBound)
        c.Next()
    }
}
```

## 6. 脱敏工具

### 6.1 utils/mask.go

```go
package utils

// MaskPhone 手机号脱敏
func MaskPhone(phone string) string {
    if len(phone) != 11 {
        return phone
    }
    return phone[:3] + "****" + phone[7:]
}

// MaskWechat 微信号脱敏
func MaskWechat(wechat string) string {
    if len(wechat) <= 4 {
        return wechat[:2] + "****"
    }
    return wechat[:2] + "****" + wechat[len(wechat)-2:]
}

// MaskQQ QQ号脱敏
func MaskQQ(qq string) string {
    if len(qq) <= 4 {
        return qq[:2] + "****"
    }
    return qq[:2] + "****" + qq[len(qq)-2:]
}

// MaskName 姓名脱敏
func MaskName(name string) string {
    runes := []rune(name)
    if len(runes) <= 1 {
        return name
    }
    return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}

// MaskAddress 地址脱敏
func MaskAddress(address string) string {
    // 只保留省市区，后面打码
    // 简单实现：截取前15字符 + ****
    if len(address) <= 15 {
        return address[:len(address)/2] + "****"
    }
    return address[:15] + "****"
}
```

## 7. 路由设计

### 7.1 新增路由

```go
// 教务绑定相关
academic := api.Group("/academic")
{
    academic.POST("/bind", handler.BindAcademic)
    academic.GET("/status", handler.GetBindingStatus)
    academic.DELETE("/unbind", handler.UnbindAcademic)
}
```

### 7.2 权限控制

```go
// 需要绑定的操作
errands := api.Group("/errands")
errands.Use(middleware.RequireBinding()) // 发布、接单需要绑定
{
    errands.POST("", handler.CreateErrand)      // 发布
    errands.POST("/:id/accept", handler.AcceptErrand) // 接单
    // ... 其他需要绑定的操作
}

// 可选绑定的操作（脱敏控制）
errands.GET("", handler.ListErrands) // 列表：可选绑定
errands.GET("/:id", handler.GetErrand) // 详情：可选绑定
```

## 8. 缓存设计

### 8.1 Redis Key 设计

```
academic:binding:{user_id}  -> JSON(BindingStatus)  TTL: 24h
```

### 8.2 缓存策略

- 绑定成功：写入缓存
- 解绑：删除缓存
- 查询：先查缓存，未命中查数据库

## 9. 数据库迁移

```sql
CREATE TABLE academic_bindings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    student_id VARCHAR(20) NOT NULL COMMENT '学号',
    real_name VARCHAR(50) NOT NULL COMMENT '真实姓名',
    college VARCHAR(100) COMMENT '学院',
    major VARCHAR(100) COMMENT '专业',
    grade VARCHAR(10) COMMENT '年级',
    status TINYINT DEFAULT 1 COMMENT '状态：1-正常 0-已解绑',
    bound_at DATETIME NOT NULL COMMENT '绑定时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_user_id (user_id),
    UNIQUE KEY idx_student_id (student_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='教务绑定表';
```

## 10. 实现步骤

### 阶段一：基础功能
1. 创建数据模型 `model/academic_binding.go`
2. 创建数据访问层 `repository/academic_binding_repository.go`
3. 创建验证器接口和模拟实现
4. 创建业务逻辑层 `service/academic_binding_service.go`
5. 创建 HTTP 处理层 `handler/academic_binding_handler.go`
6. 注册路由和数据库迁移

### 阶段二：权限控制
1. 创建绑定检查中间件 `middleware/binding_check.go`
2. 创建脱敏工具 `utils/mask.go`
3. 修改现有路由，添加权限控制
4. 修改现有接口，添加脱敏逻辑

### 阶段三：缓存优化
1. 实现绑定状态 Redis 缓存
2. 添加缓存失效逻辑

---

*创建时间: 2026-03-10*