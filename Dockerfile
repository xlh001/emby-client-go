# 后端构建阶段
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# 复制依赖文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY/backend/ ./

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o emby-manager cmd/server/main.go

# 前端构建阶段
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# 复制依赖文件
COPY frontend/package*.json ./

# 安装依赖
RUN npm ci --only=production

# 复制源代码
COPY/frontend/ ./

# 构建前端
RUN npm run build

# 最终运行阶段
FROM alpine:latest

# 安装必要的依赖
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# 从构建阶段复制文件
COPY --from=backend-builder /app/emby-manager .
COPY --from=backend-builder /app/configs ./configs
COPY --from=frontend-builder /app/dist ./frontend/dist

# 创建数据目录
RUN mkdir -p ./data

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080
ENV DATABASE_TYPE=sqlite
ENV DATABASE_DATABASE=./data/emby_manager.db

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动命令
CMD ["./emby-manager"]