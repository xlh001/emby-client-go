# Emby Manager 开发环境

## 快速启动

### 前置要求
- Docker
- Docker Compose
- Git

### 启动步骤

1. 克隆项目
```bash
git clone <repository-url>
cd emby-manager
```

2. 启动开发环境
```bash
docker-compose up -d
```

3. 访问应用
- 前端: http://localhost:3000
- 后端API: http://localhost:8080
- API文档: http://localhost:8080/swagger/index.html

### 数据库连接信息
- 类型: PostgreSQL
- 主机: localhost:5432
- 数据库: emby_manager
- 用户名: emby_user
- 密码: emby_password

### 本地开发

#### 后端开发
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

#### 前端开发
```bash
cd frontend
npm install
npm run dev
```

### 环境变量配置

创建 `.env` 文件来配置环境变量：

```env
# 数据库配置
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=emby_user
DB_PASSWORD=emby_password
DB_DATABASE=emby_manager

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT配置
JWT_SECRET=your_secret_key_here_change_in_production
JWT_EXPIRE_TIME=86400

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=debug
```

### 常用命令

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止所有服务
docker-compose down

# 重置数据库
docker-compose down -v
docker-compose up -d postgres
```

### 项目结构

```
emby-manager/
├── backend/                 # Go后端服务
│   ├── cmd/server/         # 服务入口
│   ├── internal/           # 内部包
│   │   ├── config/         # 配置管理
│   │   ├── handlers/       # HTTP处理器
│   │   ├── middleware/     # 中间件
│   │   ├── models/         # 数据模型
│   │   ├── services/       # 业务逻辑
│   │   └── utils/          # 工具函数
│   └── pkg/                # 公共包
├── frontend/               # Vue3前端应用
│   ├── src/
│   │   ├── components/     # 组件
│   │   ├── views/          # 页面
│   │   ├── stores/         # 状态管理
│   │   ├── services/       # API服务
│   │   ├── types/          # TypeScript类型
│   │   └── utils/          # 工具函数
├── docker-compose.yml      # 开发环境配置
└── README.md              # 项目文档
```

### 开发注意事项

1. 代码修改会自动热重载
2. 数据库数据会持久化存储
3. API文档自动更新
4. 日志输出到控制台和文件

### 故障排除

1. **端口冲突**: 确保端口 3000, 8080, 5432, 6379 未被占用
2. **数据库连接失败**: 检查PostgreSQL服务状态
3. **前端构建失败**: 删除 node_modules 目录重新安装依赖