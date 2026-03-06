# Alterdo重构为officeworker 个人版本重构方案
要求：一次做一件模块化的事，一边做一边讲解，每次做完在此文档更新记录
## 零、设计总则
所有状态存储在redis，后端应用做到无状态存储
## 一、项目概述
重构到 /officeworker

### 1.1 项目目标
将 AlterDo 重构为个人学习版本，专注于后端架构设计和工程化实践，保留 Agent Server 和 Tauri 客户端，打造一个高质量的 AI Agent 平台。

### 1.2 核心价值体现
- **微服务架构设计** - 完整的 Service+Repository+Handler 分层
- **分布式系统实践** - Redis 分布式锁、限流、熔断
- **高性能优化** - 连接池、缓存策略、并发控制
- **可观测性建设** - 结构化日志、Metrics、Tracing
- **Clean Architecture** - 依赖注入、接口抽象

### 1.3 技术栈

#### 后端 (Go)
- **Web 框架**: Gin (高性能 HTTP 框架)
- **ORM**: GORM (功能强大的 ORM 库)
- **缓存**: Redigo + 分布式锁
- **依赖注入**: Uber FX
- **可观测性**: Prometheus + OpenTelemetry + Zap
- **配置管理**: Viper

#### 保留组件
- **Agent Server**: Bun + Hono + Claude Code SDK (Docker 容器运行)
- **客户端**: Tauri 2 + Vue 3
- **数据存储**: MySQL 8.0 + Redis 7.2

---

## 二、架构设计

### 2.1 整体架构

客户端层包括 Tauri App 和 Web Console，通过 HTTP/SSE 与后端通信。
API Gateway 层处理认证、限流、监控等中间件功能。
Service 层包括认证、会话、Agent、文件、流式通信等核心服务。
Repository 层处理数据访问，包括 MySQL 和 Redis。
底层存储使用 MySQL 和 Redis。
Agent Container Pool 使用 Docker 运行多个 Agent Server 实例。

### 2.2 Clean Architecture 分层

```
app/
├── cmd/                    # 应用入口
│   └── main.go
├── internal/               # 内部代码（不可导出）
│   ├── handler/           # HTTP 处理器层
│   ├── service/           # 业务逻辑层
│   ├── repository/        # 数据访问层
│   ├── domain/            # 领域模型
│   ├── pkg/              # 内部工具包
│   └── config/           # 配置管理
├── models/               # GORM 数据模型
├── api/                  # API 定义
└── docker/               # Docker 相关
```

---

## 三、核心模块设计

### 3.1 认证模块 (简化版)

**设计原则：**
- 去掉复杂的权限系统，简化为 User/Admin 角色
- JWT Token 管理，支持自动刷新
- 使用 GORM 的标签和钩子特性

**API 设计：**
- POST /api/v1/auth/register - 用户注册
- POST /api/v1/auth/login - 用户登录
- POST /api/v1/auth/refresh - Token 刷新
- GET /api/v1/auth/me - 获取当前用户信息

**技术亮点：**
- GORM 标签定义，支持自动迁移和钩子
- JWT 使用 RS256 非对称加密
- Token 刷新机制，使用 Redis 黑名单

### 3.2 会话管理模块

**设计原则：**
- 使用docker创建沙箱，将用户后续向沙箱的请求（如发送消息），直接转发进去，当沙箱内部服务处理
- 沙箱生命周期无需手动管理，当2个小时没有新请求进到容器里，就把容器回收
- 使用 redis 存储 用户和后端容器的关系，当访问容器的路由进来时，需要检测jwt token，再检测用户是否有权向后端容器发消息，
- 使用 Redis 存储会话状态（快速访问）
- 分布式锁防止并发创建冲突

**API 设计：**
- POST /api/v1/sessions - 创建会话
- POST /api/v1/sessions/:id -向指定会话传送消息，带json字段
- GET /api/v1/sessions - 获取会话列表

**技术亮点：**
- Redis Set + Hash 混合存储用户会话
- 容器池预分配策略（预热容器）
- 基于使用率的自动扩缩容
- 使用 Redis Sorted Set 实现会话 TTL

### 3.3 Agent 调度模块

**设计原则：**
- 基于负载的容器选择策略
- SSE 长连接管理
- 断线重连机制

**API 设计：**
- POST /api/v1/agents/:id/tasks - 提交任务
- GET /api/v1/agents/:id/status - 查询状态
- POST /api/v1/agents/:id/stop - 停止 Agent

**技术亮点：**
- Gin SSE 长连接，支持高并发
- 连接池管理，复用 HTTP 连接
- 使用 Context 实现超时控制和取消
- 优雅关闭，处理正在进行的任务

### 3.4 文件传输模块

**设计原则：**
- 分片上传，支持大文件
- 断点续传
- 并发上传优化

**API 设计：**
- POST /api/v1/files/init - 初始化上传
- POST /api/v1/files/chunk - 上传分片
- POST /api/v1/files/complete - 完成上传
- GET /api/v1/files/:id/info - 获取文件信息
- GET /api/v1/files/:id/download - 下载文件

**技术亮点：**
- 使用 Tauri 本地文件 API（不需要上传到服务器）
- 前端直接扫描本地文件树
- 后端只存储文件元数据
- Agent 容器通过 volume 挂载访问文件

### 3.5 流式通信模块

**设计原则：**
- SSE 推送 Agent 执行进度
- 事件广播（多个会话共享状态）
- 连接心跳检测

**技术亮点：**
- Gin SSE 实现，支持自动重连
- 使用 Redis Pub/Sub 实现多实例事件广播
- 连接超时检测，自动清理僵尸连接
- 限流保护，防止客户端刷屏

---

## 四、中间件设计

### 4.1 认证中间件
验证 JWT Token，提取用户信息并设置到 Context。

### 4.2 限流中间件
使用 Redis 实现分布式限流，采用滑动窗口算法。

### 4.3 Metrics 中间件
记录 HTTP 请求的耗时、方法、路径、状态码等指标。

---

## 五、可观测性设计

### 5.1 结构化日志
使用 Zap 日志库记录结构化日志，包括用户 ID、IP、耗时等信息。

### 5.2 Metrics 指标
使用 Prometheus 记录关键指标：
- QPS (每秒请求数)
- 响应时间 P50/P95/P99
- 错误率
- 容器池使用率
- Redis 命中率

### 5.3 Tracing 链路追踪
使用 OpenTelemetry 进行分布式链路追踪，记录关键操作的属性。

---

## 六、缓存策略

### 6.1 多级缓存

本地缓存 (sync.Map) - 热点数据，低延迟，TTL: 10s
  ↓ Miss
Redis 缓存 - 共享缓存，分布式，TTL: 1h
  ↓ Miss
MySQL 数据库 - 持久化

### 6.2 缓存失效策略

- 主动失效：数据更新时删除缓存
- 被动失效：TTL 过期
- 缓存穿透：布隆过滤器
- 缓存雪崩：随机 TTL

---

## 七、部署架构

### 7.1 单机部署（个人项目）

使用 Docker Compose 部署，包括：
- App (Go) :8080
- MySQL :3306
- Redis :6379

### 7.2 配置管理

使用 Viper + 环境变量，支持环境变量覆盖配置。

---

## 八、实施进度

### 总体进度

```
阶段一: [██████████████████████] 100%
阶段二: [██████░░░░░░░░░░░░░░░] 40%
阶段三: [░░░░░░░░░░░░░░░░░░░░░] 0%
阶段四: [░░░░░░░░░░░░░░░░░░░░░] 0%
阶段五: [░░░░░░░░░░░░░░░░░░░░░] 0%
```

**当前状态**: 阶段二进行中

### 阶段一：基础设施 (5/5)
- [x] 项目结构搭建
- [x] GORM 集成
- [x] Gin 框架配置
- [x] 配置管理（Viper）
- [x] 日志系统（Zap）

### 阶段二：核心模块 (1/5)
- [x] 认证模块（简化版）
- [ ] 会话管理模块
- [ ] Agent 调度模块
- [ ] 文件传输模块
- [ ] 流式通信模块

### 阶段三：中间件与缓存 (0/5)
- [ ] 认证中间件
- [ ] 限流中间件
- [ ] Metrics 中间件
- [ ] 多级缓存策略
- [ ] 分布式锁

### 阶段四：可观测性 (0/4)
- [ ] 结构化日志完善
- [ ] Prometheus 指标
- [ ] OpenTelemetry Tracing
- [ ] Grafana Dashboard

### 阶段五：测试与优化 (0/4)
- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能优化
- [ ] 文档完善

---

## 九、技术亮点总结

### 9.1 架构设计
- ✅ Clean Architecture 分层
- ✅ 依赖注入（Uber FX）
- ✅ 接口抽象与多态

### 9.2 性能优化
- ✅ Gin 高性能 HTTP 框架
- ✅ 连接池复用
- ✅ 多级缓存策略
- ✅ 并发控制

### 9.3 分布式实践
- ✅ Redis 分布式锁
- ✅ 限流与熔断
- ✅ 容器池管理
- ✅ 事件广播

### 9.4 可观测性
- ✅ 结构化日志
- ✅ Metrics 监控
- ✅ 链路追踪
- ✅ 健康检查

### 9.5 工程化
- ✅ ORM 数据模型（GORM）
- ✅ 配置管理
- ✅ 优雅关闭
- ✅ Docker 化部署

---

## 十、后续扩展方向

### 10.1 短期扩展
- WebSocket 支持双向通信
- Webhook 集成
- 任务队列（使用 Temporal）

### 10.2 长期规划
- 服务网格（Istio）
- 云原生部署（Kubernetes）
- 多租户支持
- AI 模型热更新

---

## 十一、参考资料

- [Gin Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [Uber FX](https://github.com/uber-go/fx)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)

---

## 十二、实施记录

### 阶段一完成记录 (2026-03-02)

**完成内容**：
1. 项目结构搭建 - Clean Architecture 分层目录结构
2. Gin 框架配置 - Server 封装、RouterGroup、中间件链（CORS/Logger/Recovery）
3. GORM 集成 - MySQL 连接池、AutoMigrate、数据模型
4. 配置管理（Viper）- .env 文件加载、默认值设置、环境变量覆盖
5. 日志系统（Zap）- 统一日志接口、JSON/Console 格式支持、级别配置

**技术亮点**：
- Clean Architecture 分层清晰
- GORM 自动迁移保持模型同步
- 连接池优化（空闲/最大连接、生命周期）
- 中间件链式处理
- 结构化日志支持 JSON 输出

**文件清单**：
- `cmd/main.go` - 应用入口
- `internal/config/config.go` - 配置管理
- `internal/pkg/gin/` - Gin 服务器和路由
- `internal/pkg/middleware/` - 中间件
- `internal/pkg/logger/logger.go` - 日志封装
- `internal/repository/database.go` - 数据库连接
- `models/*.go` - 数据模型

---

### 阶段二-认证模块完成记录 (2026-03-03)

**完成内容**：
1. JWT 工具 - Token 生成、验证、刷新（HS256 加密）
2. Redis 封装 - 客户端连接、Token 黑名单
3. 用户仓储层 - CRUD 操作、唯一性检查
4. 认证服务层 - 注册、登录、刷新、登出、获取用户信息
5. 认证处理器 - HTTP 接口实现
6. 认证中间件 - Token 验证、黑名单检查
7. 认证路由 - `/api/v1/auth/*` 路由注册

**API 接口**：
- POST /api/v1/auth/register - 用户注册
- POST /api/v1/auth/login - 用户登录
- POST /api/v1/auth/refresh - Token 刷新
- POST /api/v1/auth/logout - 用户登出（需认证）
- GET /api/v1/auth/me - 获取当前用户信息（需认证）

**技术亮点**：
- JWT 双 Token 机制（Access Token + Refresh Token）
- Redis 黑名单实现 Token 注销
- BCrypt 密码加密存储
- 统一响应格式（code/message/data）
- 中间件自动注入用户信息到 Context

**文件清单**：
- `internal/pkg/jwt/jwt.go` - JWT 管理器
- `internal/pkg/redis/client.go` - Redis 客户端
- `internal/pkg/redis/blacklist.go` - Token 黑名单
- `internal/repository/user_repository.go` - 用户仓储
- `internal/service/auth_service.go` - 认证服务
- `internal/handler/auth_handler.go` - 认证处理器
- `internal/pkg/middleware/auth.go` - 认证中间件
- `internal/pkg/gin/auth_router.go` - 认证路由

---

### 阶段二-会话管理模块完成记录 (2026-03-03)

**完成内容**：
1. 会话仓储层 - CRUD 操作、用户会话查询、状态更新
2. Redis 分布式锁 - Lock/Unlock/Extend、Lua 脚本保证原子性
3. 容器池管理 - 容器获取/释放/销毁、预热、统计信息
4. 会话服务层 - 创建/获取/更新/删除会话、激活/停用会话
5. 会话处理器 - HTTP 接口实现
6. 会话路由 - `/api/v1/sessions/*` 路由注册

**API 接口**：
- POST /api/v1/sessions - 创建会话（需认证）
- GET /api/v1/sessions - 获取当前用户的会话列表（需认证）
- GET /api/v1/sessions/active - 获取当前用户的活跃会话（需认证）
- GET /api/v1/sessions/:id - 获取会话详情（需认证）
- PUT /api/v1/sessions/:id - 更新会话信息（需认证）
- DELETE /api/v1/sessions/:id - 删除会话（需认证）
- POST /api/v1/sessions/:id/activate - 激活会话（需认证）
- POST /api/v1/sessions/:id/deactivate - 停用会话（需认证）

**技术亮点**：
- Redis Set + Hash 混合存储会话状态
- 分布式锁防止并发创建冲突
- 容器池管理（获取/释放/销毁/预热）
- 自动容器池预热（启动时创建空闲容器）
- Redis 缓存会话信息减少数据库查询
- 容器池使用率监控（active/idle/usage/capacity）

**Docker 集成**：
- Docker 客户端封装 - 容器创建/启动/停止/删除
- 容器生命周期管理 - 按需创建、自动销毁、优雅关闭
- 卷挂载 - 容器通过 volume 挂载访问文件系统
- 镜像管理 - 支持镜像拉取、使用指定镜像
- 网络配置 - 支持 Docker 网络、容器互联
- 容器状态监控 - 查询容器运行状态、获取日志

**文件清单**：
- `internal/repository/session_repository.go` - 会话仓储
- `internal/pkg/redis/distributed_lock.go` - 分布式锁
- `internal/pkg/docker/client.go` - Docker 客户端管理
- `internal/service/container_pool.go` - 容器池管理（集成 Docker）
- `internal/service/session_service.go` - 会话服务
- `internal/handler/session_handler.go` - 会话处理器
- `internal/pkg/gin/session_router.go` - 会话路由

---

## 十三、Docker 配置说明

### 13.1 环境变量配置

在 `.env` 文件中添加以下 Docker 相关配置：

```bash
# Docker
DOCKER_IMAGE=agent-server:latest
DOCKER_NETWORK=officeworker-net
DOCKER_FILE_BASE_PATH=/tmp/officeworker/files
DOCKER_POOL_MIN_SIZE=2
DOCKER_POOL_MAX_SIZE=10
```

### 13.2 配置说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `DOCKER_IMAGE` | Agent Server 镜像名称 | `agent-server:latest` |
| `DOCKER_NETWORK` | Docker 网络名称 | `officeworker-net` |
| `DOCKER_FILE_BASE_PATH` | 文件存储路径（挂载到容器） | `/tmp/officeworker/files` |
| `DOCKER_POOL_MIN_SIZE` | 容器池最小容器数 | `2` |
| `DOCKER_POOL_MAX_SIZE` | 容器池最大容器数 | `10` |

### 13.3 前置条件

1. **安装 Docker**：确保系统已安装 Docker Engine
   ```bash
   docker --version
   ```

2. **构建 Agent Server 镜像**：
   ```bash
   cd agent-server
   docker build -t agent-server:latest .
   ```

3. **创建 Docker 网络**：
   ```bash
   docker network create officeworker-net
   ```

4. **创建文件存储目录**：
   ```bash
   mkdir -p /tmp/officeworker/files
   chmod 777 /tmp/officeworker/files
   ```

### 13.4 容器生命周期

| 阶段 | 操作 | 说明 |
|------|------|------|
| **预热** | Warmup | 启动时创建 `POOL_MIN_SIZE` 个空闲容器 |
| **获取** | Acquire | 从池中获取空闲容器或创建新容器 |
| **释放** | Release | 容器标记为空闲，可被复用 |
| **销毁** | Destroy | 停止并删除 Docker 容器 |
| **清理** | Cleanup | 定期清理超时空闲容器 |

---

**文档版本**: v2.4
**创建日期**: 2026-03-02
**最后更新**: 2026-03-03
**作者**: Kado
**状态**: 阶段二进行中
