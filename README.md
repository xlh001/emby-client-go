# Emby 客户端管理系统

基于 Golang 开发的 Emby 服务器和设备管理系统，提供多服务器连接、设备管理和设备-服务器关联功能。

## 功能特性

### 📱 设备管理
- 设备添加、编辑、删除
- 设备状态监控（活跃/非活跃）
- 设备信息记录（平台、IP地址、MAC地址等）
- 自动生成设备唯一标识符

### 🖥️ 服务器管理
- 多 Emby 服务器连接管理
- 服务器连接状态检测
- 服务器信息获取和显示
- 支持认证和匿名连接

### 🔗 设备-服务器关联
- 灵活的设备-服务器关联配置
- 优先级设置
- 关联关系启用/禁用控制
- 从服务器同步设备信息

### 🌐 Web界面
- 现代化响应式管理界面
- 实时数据更新
- 用户友好的操作界面
- 移动端适配

### 🔐 安全认证
- JWT Token 认证
- 用户权限管理
- 安全密码存储

## 技术栈

- **后端**: Go + Gin Web框架
- **数据库**: SQLite (开发) / PostgreSQL (生产)
- **ORM**: GORM
- **前端**: Bootstrap 5 + Vanilla JavaScript
- **认证**: JWT Token

## 快速开始

### 1. 环境要求
- Go 1.21 或更高版本
- Git

### 2. 克隆项目
```bash
git clone https://github.com/yourusername/emby-client-go.git
cd emby-client-go
```

### 3. 安装依赖
```bash
go mod tidy
```

### 4. 初始化数据库
```bash
go run cmd/init.go
```

### 5. 启动服务
```bash
go run cmd/main.go
```

### 6. 访问系统
打开浏览器访问 `http://localhost:8080`

默认管理员账户：
- 用户名: `admin`
- 密码: `admin123`

## 配置说明

系统通过 `config.yaml` 文件进行配置：

```yaml
server:
  port: "8080"           # 服务端口

database:
  path: "./data/emby.db" # 数据库文件路径

jwt:
  secret: "your-secret-key-change-in-production" # JWT密钥（生产环境请修改）
  expire_time: 24                                           # Token过期时间（小时）
```

## API 文档

### 认证接口
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/change-password` - 修改密码

### 设备接口
- `GET /api/devices` - 获取设备列表
- `POST /api/devices` - 添加设备
- `PUT /api/devices/:id` - 更新设备
- `DELETE /api/devices/:id` - 删除设备
- `GET /api/devices/:id/servers` - 获取设备关联的服务器
- `POST /api/devices/:id/servers/:serverId` - 设备关联服务器

### 服务器接口
- `GET /api/servers` - 获取服务器列表
- `POST /api/servers` - 添加服务器
- `PUT /api/servers/:id` - 更新服务器
- `DELETE /api/servers/:id` - 删除服务器
- `POST /api/servers/:id/test` - 测试服务器连接
- `POST /api/servers/:id/sync-devices` - 同步设备

## 项目结构

```
emby-client-go/
├── cmd/                  # 应用程序入口
│   ├── main.go          # 主程序
│   └── init.go          # 初始化程序
├── internal/            # 内部包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── handlers/        # HTTP处理器
│   ├── models/          # 数据模型
│   └── services/        # 业务逻辑服务
├── web/                 # Web资源
│   ├── static/          # 静态文件
│   │   ├── css/
│   │   └── js/
│   └── templates/       # HTML模板
├── config.yaml          # 配置文件
├── go.mod               # Go模块文件
└── README.md           # 项目说明
```

## 开发指南

### 添加新功能
1. 在 `internal/models/` 中定义数据模型
2. 在 `internal/services/` 中实现业务逻辑
3. 在 `internal/handlers/` 中添加API处理器
4. 更新路由注册
5. 添加前端界面（如果需要）

### 数据库迁移
系统使用 GORM 的 AutoMigrate 功能自动管理数据库结构。修改模型后，重启服务即可自动迁移。

## 部署

### 生产环境部署
1. 修改 `config.yaml` 中的配置
2. 使用 `go build` 编译可执行文件
3. 配置反向代理（Nginx）
4. 启动服务

### Docker 部署
建议使用 Docker 容器化部署：

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o emby-manager cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/emby-manager .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ["./emby-manager"]
```

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如果您遇到任何问题或有功能建议，请：
1. 查看现有的 [Issues](../../issues)
2. 创建新的 Issue 描述问题或建议
3. 或者直接联系项目维护者

## 更新日志

### v1.0.0 (2024-XX-XX)
- 初始版本发布
- 基础设备管理功能
- 服务器连接和管理
- 设备-服务器关联
- Web管理界面
- 用户认证系统