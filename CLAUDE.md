# DataRay 开发指南

## 项目概述

DataRay 是一个拖拽式 BI 可视化分析平台的 MVP 版本，核心功能包括数据源管理、数据集管理、拖拽式图表构建和分享功能。

## 技术栈

- **前端**: React 18 + TypeScript + Ant Design + ECharts + Zustand + @dnd-kit
- **后端**: Go + go-zero + bun + PostgreSQL
- **部署**: Docker + docker-compose

## 目录结构

```
.
├── docker-compose.yml          # Docker 编排
├── frontend/                   # 前端项目
│   ├── src/
│   │   ├── api/               # API 调用
│   │   ├── store/             # Zustand 状态管理
│   │   ├── pages/             # 页面组件
│   │   ├── styles/            # 样式文件
│   │   ├── App.tsx           # 主应用组件
│   │   └── main.tsx          # 入口文件
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
└── backend/                   # 后端项目
    ├── cmd/main.go            # 主程序 + HTTP Handlers
    ├── internal/
    │   ├── config/            # 配置加载
    │   ├── database/          # 数据库连接和迁移
    │   └── model/             # 数据模型定义
    ├── etc/config.yaml
    ├── go.mod
    └── go.sum
```

## 后端开发

### 核心文件

- `cmd/main.go`: 包含所有 HTTP Handler 和路由注册
- `internal/model/model.go`: 数据库表结构定义
- `internal/database/database.go`: 数据库初始化和迁移
- `internal/config/config.go`: 配置文件加载

### API 路由

所有路由都在 `cmd/main.go` 的 `registerRoutes` 函数中定义：

```go
func registerRoutes(mux *http.ServeMux, db *bun.DB) {
    mux.HandleFunc("/api/datasources", handleDatasources(db))
    mux.HandleFunc("/api/datasources/", handleDatasource(db))
    // ... 其他路由
}
```

### 添加新 API

1. 在 `cmd/main.go` 中创建 handler 函数
2. 在 `registerRoutes` 中注册路由
3. 更新 `internal/model/model.go` 添加新的数据结构（如需要）

### 数据库模型

模型定义在 `internal/model/model.go`，使用 bun ORM：

```go
type Datasource struct {
    bun.BaseModel `bun:"bi_datasource"`
    ID           int             `bun:"id,pk,autoincrement"`
    Name         string          `bun:"name"`
    // ...
}
```

表结构通过 `database.RunMigrations` 中的 SQL 自动创建。

## 前端开发

### 页面组件

| 文件 | 功能 |
|------|------|
| `pages/Datasource.tsx` | 数据源管理 |
| `pages/Dataset.tsx` | 数据集管理 |
| `pages/Charts.tsx` | 图表列表 |
| `pages/ChartBuilder.tsx` | 拖拽式图表构建器 |
| `pages/Share.tsx` | 分享管理 |
| `pages/ShareView.tsx` | 分享查看 |

### 状态管理

使用 Zustand，定义在 `store/index.ts`：

```typescript
interface AppState {
  datasources: Datasource[]
  datasets: Dataset[]
  charts: Chart[]
  // ...
}
```

### API 调用

所有 API 调用在 `api/index.ts` 中定义，使用 axios：

```typescript
export const datasourcesApi = {
  getAll: () => axios.get('/api/datasources'),
  create: (data) => axios.post('/api/datasources', data),
  // ...
}
```

### 拖拽实现

使用 @dnd-kit，在 `ChartBuilder.tsx` 中实现：

```typescript
import { DndContext, useDraggable, useDroppable } from '@dnd-kit/core'
```

## Docker 开发

### 构建镜像

```bash
docker-compose build
```

### 启动服务

```bash
docker-compose up
```

### 查看日志

```bash
docker-compose logs -f
```

### 停止服务

```bash
docker-compose down
```

## 测试

### 后端单元测试

```bash
cd backend
go test ./...
```

### 前端构建测试

```bash
cd frontend
npm run build
```

## 注意事项

1. **数据库连接**: 后端使用 bun 连接 PostgreSQL，需要确保数据库可访问
2. **跨域问题**: 前端默认连接 localhost:8080，如有变化请修改 `frontend/src/api/index.ts`
3. **SQL 注入**: 当前实现中 SQL 查询直接拼接，MVP 版本暂未做严格过滤
4. **密码安全**: 数据库密码以明文存储，生产环境需要加密处理
