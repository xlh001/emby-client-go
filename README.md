# Emby Manager

一个功能完整的Emby服务器管理平台，支持多服务器管理、媒体库浏览和播放控制。

## 功能特性

- 🏠 **多服务器管理**: 添加、编辑、删除多个Emby服务器
- 👥 **用户管理**: JWT认证、权限控制
- 📚 **媒体库浏览**: 跨服务器浏览和搜索
- 🎮 **播放控制**: 远程播放控制和进度同步
- 🐳 **容器化部署**: 支持Docker/K8s/裸机部署

## 技术栈

- **后端**: Go + Gin + GORM + JWT
- **前端**: Vue3 + TypeScript + Element Plus + Vite
- **数据库**: SQLite (默认) / PostgreSQL / MySQL
- **部署**: Docker + Kubernetes

## 快速开始

### 开发环境

```bash
# 启动开发环境
docker-compose up -d

# 后端开发
cd backend && go mod tidy && go run cmd/server/main.go

# 前端开发
cd frontend && npm install && npm run dev
```

### 生产部署

```bash
# Docker部署
docker-compose -f docker-compose.prod.yml up -d

# Kubernetes部署
kubectl apply -f k8s/
```

## 项目结构

```
emby-manager/
├── backend/                 # Go后端服务
├── frontend/               # Vue3前端应用
├── k8s/                    # Kubernetes配置
├── docs/                   # 项目文档
├── docker-compose.yml      # 开发环境
└── docker-compose.prod.yml # 生产环境
```

## API文档

服务启动后访问: http://localhost:8080/swagger/index.html

## 许可证

MIT License