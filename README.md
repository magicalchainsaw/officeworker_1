# OfficeWorker

一个基于 Go 的 AI Agent 平台，采用 Clean Architecture 分层架构。

## 技术栈

- **Web 框架**: Gin
- **ORM**: GORM
- **缓存**: Redis
- **依赖注入**: Uber FX
- **可观测性**: Prometheus + OpenTelemetry + Zap
- **配置管理**: Viper

## 项目结构

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

## 快速开始

```bash
# 安装依赖
go mod tidy

# 运行服务
go run cmd/main.go

# 使用 Docker Compose
docker-compose up -d
```

## 开发进度

参见 [Alterdo迁移到officeworker方案.md](../Alterdo迁移到officeworker方案.md)
