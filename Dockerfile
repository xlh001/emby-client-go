# 构建阶段
FROM node:20-alpine AS builder

# 构建前端
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# 构建后端
WORKDIR /app/backend
RUN apk add --no-cache go gcc musl-dev sqlite-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o emby-manager cmd/server/main.go

# 运行阶段
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite-libs
WORKDIR /root/
COPY --from=builder /app/backend/emby-manager .
COPY --from=builder /app/backend/configs ./configs
COPY --from=builder /app/frontend/dist/frontend/browser ./frontend/dist
RUN mkdir -p ./data
EXPOSE 8080
ENV SERVER_HOST=0.0.0.0 SERVER_PORT=8080 DATABASE_TYPE=sqlite DATABASE_DATABASE=./data/emby_manager.db
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["./emby-manager"]
