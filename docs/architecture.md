# DataRay 架构文档

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
│   │   ├── pages/           # 页面组件
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

## 关键约束

1. **配置文件**: 后端使用 TOML 格式 (`etc/config.toml`)
2. **CORS**: 后端配置 CORS 中间件允许跨域
3. **API 基础URL**: 前端默认连接 `http://localhost:8080`，修改 `frontend/src/lib/api/client.ts`
4. **前端端口**: Vite 默认 3000
5. **热重载**: 后端开发使用 `air` 工具 (`make dev`)
