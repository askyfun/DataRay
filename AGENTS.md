# DataRay 开发规范

**生成时间**: 2026-02-21
**项目**: DataRay - 拖拽式 BI 可视化分析平台

## 项目概述

Monorepo 结构，包含前端 (React/TypeScript) 和后端 (Go)。核心功能：数据源管理、数据集管理、拖拽式图表构建、分享功能。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | React 18 + TypeScript + Ant Design 5.x + ECharts 5.x + Zustand 4.x + @dnd-kit + Vite 6 + Sentry |
| 后端 | Go 1.26 + Gin + bun ORM + PostgreSQL + Sentry |
| 部署 | Docker + docker-compose |

## 目录结构

```
.
├── frontend/                  # 前端项目
│   ├── src/
│   │   ├── api/              # API 调用 (axios)
│   │   ├── store/            # Zustand 状态管理
│   │   ├── pages/            # 页面组件
│   │   └── styles/           # 样式文件
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── backend/                   # 后端项目
│   ├── cmd/main.go           # 入口 + 路由注册
│   ├── internal/
│   │   ├── config/           # 配置加载 (TOML)
│   │   ├── database/         # 数据库连接
│   │   ├── datasource/       # 数据源驱动抽象
│   │   │   ├── driver.go     # Driver 接口
│   │   │   ├── postgresql.go
│   │   │   ├── mysql.go
│   │   │   ├── clickhouse.go
│   │   │   └── starrocks.go
│   │   ├── middleware/       # Gin 中间件
│   │   └── model/            # 数据模型
│   ├── etc/config.toml       # 配置文件
│   └── go.mod
├── Makefile
└── docker-compose.yml
```

## 运行命令

### 前端

```bash
cd frontend && npm install

npm run dev          # 开发模式 (端口 3000)
npm run build        # 构建 (tsc + vite build)
npm run preview      # 预览构建结果
```

### 后端

```bash
cd backend && go mod download

go run ./cmd/main.go -f etc/config.toml    # 开发模式
go build -o bin/server ./cmd/main.go       # 构建

# 运行所有测试
go test ./...

# 运行单个测试
go test -v ./cmd -run TestHandleDatasourcesGET

# 查看测试覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Docker

```bash
make docker-up      # 启动所有服务
make docker-logs    # 查看日志
make docker-down    # 停止服务
make dev            # 本地开发 (使用 air 热重载)
```

## 代码风格

### 前端 (TypeScript)

**TypeScript 配置**:
- `strict: true` - 严格模式
- `noUnusedLocals: true` - 未使用变量报错
- `noUnusedParameters: true` - 未使用参数报错
- 路径别名: `@/*` 指向 `./src/*`

**前后端字段命名规范** (CRITICAL):

前后端 JSON 通信使用 **snake_case** 命名，确保字段完全一致。

```typescript
// ❌ 错误 - 字段名与后端不一致
interface DatasetColumn {
  dataType: string;    // 后端返回 type
  expression: string;  // 后端返回 expr
}

// ✅ 正确 - 与后端保持一致
interface DatasetColumn {
  type: string;
  expr: string;
  role: string;
  name: string;
  comment: string;
  typeConfig: TypeConfig;
}
```

常见对应关系：
| 后端 (Go JSON) | 前端 (TypeScript) | 用途 |
|----------------|-------------------|------|
| `data_type` | `dataType` | 数据源表结构 |
| `table_name` | `tableName` | 表名 |
| `query_sql` | `querySql` | 查询SQL |
| `query_config` | `queryConfig` | 图表查询配置 |
| `created_at` | `createdAt` | 创建时间 |

**导入顺序**:
```typescript
// 1. React/库导入
import React, { useState, useEffect } from 'react';
import { Button, Table, Modal } from 'antd';
import { useNavigate } from 'react-router-dom';

// 2. 项目内部导入
import { datasourcesApi } from '@/api';
import { useStore } from '@/store';

// 3. 类型导入
import type { Datasource, Dataset } from '@/api';
```

**命名规范**:
- 组件: PascalCase (`ChartBuilder.tsx`)
- 工具函数/变量: camelCase
- 常量: UPPER_SNAKE_CASE
- 禁止使用 `any`，使用 `unknown` 代替

### 后端 (Go)

**框架**: 使用 Gin (NOT go-zero)

**包导入顺序** (标准库 → 第三方 → 项目内部):
```go
import (
    "log/slog"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/uptrace/bun"

    "dataray/internal/config"
    "dataray/internal/database"
    "dataray/internal/middleware"
    "dataray/internal/model"
)
```

**命名规范**:
- 函数/变量: camelCase
- 结构体/接口: PascalCase
- Error 变量: `ErrXXX` 或 `XXXError`

**Handler 模式 (Gin)**:
```go
func listDatasources(c *gin.Context) {
    db := c.MustGet("db").(*bun.DB)
    var datasources []model.Datasource
    if err := db.NewSelect().Model(&datasources).Scan(c.Request.Context()); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, datasources)
}
```

**数据库模型** (bun ORM):
```go
type Datasource struct {
    bun.BaseModel `bun:"bi_datasource"`
    ID            int    `bun:"id,pk,autoincrement" json:"id"`
    Name          string `bun:"name" json:"name"`
    Host          string `bun:"host" json:"host"`
    // ...
}
```

**错误处理**:
```go
// 正确
if err != nil {
    slog.Error("操作失败", "error", err)
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}

// 避免
if err != nil {
    // 空块 - 禁止
}
```

## API 路由

| 路径 | 方法 | 功能 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/datasources` | GET/POST | 数据源列表/创建 |
| `/api/datasources/:id` | GET/DELETE | 数据源详情/删除 |
| `/api/datasources/test` | POST | 测试连接 |
| `/api/datasources/:id/tables` | GET | 获取表列表 |
| `/api/datasources/:id/tables/:table/columns` | GET | 获取表字段 |
| `/api/datasets` | GET/POST | 数据集列表/创建 |
| `/api/datasets/:id` | GET/DELETE | 数据集详情/删除 |
| `/api/datasets/:id/columns` | GET | 获取字段列表 |
| `/api/charts` | GET/POST | 图表列表/创建 |
| `/api/charts/:id` | GET/PUT/DELETE | 图表 CRUD |
| `/api/charts/:id/data` | GET | 获取图表数据 |
| `/api/shares` | POST | 创建分享 |
| `/api/shares/:token` | GET | 获取分享信息 |
| `/share/:token` | GET | 访问分享链接 |

## API 规范

**所有 API 开发必须遵循**: `docs/api-spec.md`

### 统一响应格式

所有 API 响应必须使用以下统一格式：

```json
{
  "code": 20000,
  "msg": "success",
  "trace": "xxxxxxx",
  "data": {}
}
```

### 响应封装库

- **后端**: `backend/internal/response/response.go`
- **前端**: `frontend/src/lib/api/client.ts`

### IDL 定义

- **后端**: `backend/internal/idls/*.go`
- **前端**: `frontend/src/idls/*.ts`

## 注意事项

1. **配置文件**: 后端使用 TOML 格式 (`etc/config.toml`)
2. **CORS**: 后端配置 CORS 中间件允许跨域
3. **API 基础URL**: 前端默认连接 `http://localhost:8080`，修改 `frontend/src/lib/api/client.ts`
4. **前端端口**: Vite 默认 3000
5. **热重载**: 后端开发使用 `air` 工具 (`make dev`)

## 已知限制

- 前端无 ESLint/Prettier 配置
- 后端无 golangci-lint 配置
- 缺少端到端测试
