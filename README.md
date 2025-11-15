# Emby 管理平台

一个功能完整的 Emby 服务器管理平台，提供用户认证、服务器管理、媒体浏览、播放控制等功能。

## 功能特性

### ✅ 已实现功能

- **用户认证系统**
  - JWT 令牌认证
  - 用户注册/登录/登出
  - 账户锁定保护
  - 令牌自动刷新

- **服务器管理**
  - 多服务器连接管理
  - 服务器状态监控
  - 设备同步
  - 连接测试

- **媒体库集成**
  - 媒体库同步
  - 跨服务器搜索
  - 媒体浏览
  - 分页加载

- **播放控制**
  - 远程播放控制（播放/暂停/停止/跳转）
  - 活动会话管理
  - 播放历史记录

- **实时通信**
  - WebSocket 连接管理
  - 实时状态同步

## 技术栈

### 后端
- Go 1.24
- Gin Web Framework
- GORM (支持 SQLite/PostgreSQL/MySQL)
- JWT 认证
- WebSocket

### 前端
- Vue 3 + TypeScript
- Vite
- Element Plus
- Pinia
- Vue Router

### 部署
- Docker + Docker Compose
- Nginx (可选)
- PostgreSQL (可选)
- Redis (可选)

## 快速开始

### 开发环境

1. **启动开发环境**
```bash
docker-compose -f docker-compose.dev.yml up
```

2. **访问应用**
- 前端: http://localhost:5173
- 后端: http://localhost:8080
- API文档: http://localhost:8080/swagger/index.html

### 生产环境

1. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，设置必要的配置
```

2. **启动服务**
```bash
# 基础部署（SQLite）
docker-compose up -d

# 使用 PostgreSQL
docker-compose --profile postgres up -d

# 使用 Nginx 反向代理
docker-compose --profile nginx up -d
```

3. **访问应用**
- 应用: http://localhost:8080
- Nginx: http://localhost (如果启用)

## 配置说明

### 环境变量

参考 `.env.example` 文件配置以下内容：

- **服务器配置**: 端口、模式、超时
- **数据库配置**: 类型、连接信息
- **JWT配置**: 密钥、过期时间
- **日志配置**: 级别、格式

### 数据库选择

- **SQLite**: 默认，适合小规模部署
- **PostgreSQL**: 推荐生产环境使用
- **MySQL**: 可选支持

## API 文档

启动服务后访问 Swagger 文档：
```
http://localhost:8080/swagger/index.html
```

## 项目结构

```
.
├── backend/                 # Go 后端
│   ├── cmd/server/         # 主程序
│   ├── internal/           # 内部包
│   │   ├── handlers/      # HTTP 处理器
│   │   ├── services/      # 业务逻辑
│   │   ├── models/        # 数据模型
│   │   └── middleware/    # 中间件
│   └── pkg/               # 公共包
│       ├── emby/          # Emby 客户端
│       └── websocket/     # WebSocket 管理
│
├── frontend/               # Vue 前端
│   └── src/
│       ├── views/         # 页面组件
│       ├── components/    # 可复用组件
│       ├── services/      # API 服务
│       ├── stores/        # 状态管理
│       └── router/        # 路由配置
│
├── docker-compose.yml      # 生产环境编排
├── docker-compose.dev.yml  # 开发环境编排
├── Dockerfile             # 生产镜像
└── Dockerfile.dev         # 开发镜像
```

## 开发指南

### 后端开发

```bash
cd backend
go mod download
go run cmd/server/main.go
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

## 许可证

MIT License
