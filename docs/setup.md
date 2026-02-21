# DataRay 环境搭建

## 前端

```bash
cd frontend && npm install

npm run dev          # 开发模式 (端口 3000)
npm run build        # 构建 (tsc + vite build)
npm run preview      # 预览构建结果
```

## 后端

```bash
cd backend && go mod download

go run ./cmd/main.go -f etc/config.toml    # 开发模式
go build -o bin/server ./cmd/main.go       # 构建

# 测试
go test ./...
go test -v ./cmd -run TestHandleDatasourcesGET

# 覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Docker

```bash
make docker-up      # 启动所有服务
make docker-logs    # 查看日志
make docker-down    # 停止服务
make dev            # 本地开发 (使用 air 热重载)
```
