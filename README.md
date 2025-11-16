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
- Angular 20 + TypeScript
- Angular Material
- RxJS
- Angular Router

### 部署
- Docker + Docker Compose
- Nginx (可选)
- PostgreSQL (可选)
- Redis (可选)

## 快速开始

### 方式一：Docker 部署（推荐）

1. **拉取镜像**
```bash
docker pull xlh001/emby-client-go:latest
```

2. **启动容器**
```bash
docker run -d \
  --name emby-client-go \
  -p 8080:8080 \
  -v $(pwd)/data:/root/data \
  -v $(pwd)/configs:/root/configs \
  xlh001/emby-client-go:latest
```

3. **查看初始密码**
```bash
docker logs emby-client-go | grep "默认管理员"
```

4. **访问应用**
- 应用: http://localhost:8080
- 默认用户名: `admin`
- 密码: 查看日志获取

### 方式二：Docker Compose 部署

1. **下载配置文件**
```bash
wget https://raw.githubusercontent.com/xlh001/emby-client-go/main/docker-compose.yml
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

3. **查看初始密码**
```bash
docker-compose logs emby-client-go | grep "默认管理员"
```

### 方式三：二进制部署

1. **下载二进制文件**

从 [Releases](https://github.com/xlh001/emby-client-go/releases) 下载对应平台的二进制文件：
- Linux: `emby-client-go-linux-amd64` 或 `emby-client-go-linux-arm64`
- Windows: `emby-client-go-windows-amd64.exe`
- macOS: `emby-client-go-darwin-amd64` 或 `emby-client-go-darwin-arm64`

2. **创建配置文件**
```bash
mkdir -p configs data
wget -O configs/config.yaml https://raw.githubusercontent.com/xlh001/emby-client-go/main/backend/configs/config.yaml
```

3. **运行应用**
```bash
# Linux/macOS
chmod +x emby-client-go-*
./emby-client-go-linux-amd64

# Windows
emby-client-go-windows-amd64.exe
```

4. **查看初始密码**

启动日志中会显示：`默认管理员已创建 - 用户名: admin, 密码: xxx`

### 开发环境

```bash
docker-compose -f docker-compose.dev.yml up
```

**访问地址：**
- 前端: http://localhost:4200
- 后端: http://localhost:8080
- API文档: http://localhost:8080/swagger/index.html

## 配置说明

### 配置文件

配置文件位置：`configs/config.yaml`

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"  # debug 或 release

database:
  type: "sqlite"  # sqlite, postgres, mysql
  database: "./data/emby_manager.db"

jwt:
  secret: "your-secret-key"  # 请修改为安全的密钥
  expire_time: 86400
```

### 环境变量

支持通过环境变量覆盖配置（前缀 `EMBY_`）：

```bash
EMBY_SERVER_PORT=8080
EMBY_DATABASE_TYPE=sqlite
EMBY_JWT_SECRET=your-secret-key
```

### 数据库选择

- **SQLite**: 默认，适合小规模部署，无需额外配置
- **PostgreSQL**: 推荐生产环境使用，高并发支持
- **MySQL**: 可选支持

### 首次登录

1. 启动应用后，查看日志获取初始密码
2. 使用 `admin` 和初始密码登录
3. **立即修改密码**以确保安全

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
├── frontend/               # Angular 前端
│   └── src/
│       └── app/           # 应用模块
│           ├── components/    # 可复用组件
│           ├── services/      # API 服务
│           └── models/        # 数据模型
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
npm start
```

## 许可证

MIT License
