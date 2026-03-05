# 认证模块使用指南

## API 接口

### 1. 用户注册
```
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user"
    }
  }
}
```

### 2. 用户登录
```
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user"
    }
  }
}
```

### 3. Token 刷新
```
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "testuser",
      "email": "test@example.com",
      "role": "user"
    }
  }
}
```

### 4. 用户登出
```
POST /api/v1/auth/logout
Authorization: Bearer {access_token}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 5. 获取当前用户信息
```
GET /api/v1/auth/me
Authorization: Bearer {access_token}
```

响应示例：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user"
  }
}
```

## 技术特性

### 1. JWT 双 Token 机制
- **Access Token**: 有效期较短（默认 24 小时），用于 API 请求认证
- **Refresh Token**: 有效期较长（默认 7 天），用于刷新 Access Token
- 使用 HS256 算法加密

### 2. Redis 黑名单
- 登出时将 Access Token 加入黑名单
- 黑名单自动过期（与 Token 有效期一致）
- 中间件自动验证 Token 是否在黑名单中

### 3. 密码安全
- 使用 BCrypt 算法加密存储密码
- 密码验证使用 BCrypt CompareHashAndPassword

### 4. 用户状态管理
- 支持用户状态控制（active/inactive）
- 非活跃用户无法登录

### 5. 中间件保护
- RequireAuth: 验证 Token 并注入用户信息到 Context
- RequireRole: 验证用户角色权限

## 测试步骤

### 1. 启动服务
```bash
cd /home/kado_2/workspace/officeworker
cp .env.example .env
# 修改 .env 文件中的数据库和 Redis 配置
go run cmd/main.go
```

### 2. 测试注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

### 3. 测试登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'
```

### 4. 测试获取用户信息
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer {your_access_token}"
```

### 5. 测试登出
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer {your_access_token}"
```

### 6. 测试 Token 刷新
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"{your_refresh_token}"}'
```

## 错误码

- `0`: 成功
- `-1`: 通用错误

## 常见错误

1. **username already exists**: 用户名已被注册
2. **email already exists**: 邮箱已被注册
3. **invalid username or password**: 用户名或密码错误
4. **account is inactive**: 账户已被禁用
5. **invalid token**: Token 无效或已过期
6. **token has been revoked**: Token 已被撤销（登出）
7. **user not authenticated**: 用户未认证
8. **insufficient permissions**: 权限不足
