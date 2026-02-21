# DataRay 开发规范

**生成时间**: 2026-02-21  
**项目**: DataRay - 拖拽式 BI 可视化分析平台

## 项目概览

Monorepo 结构，包含前端 (React/TypeScript) 和后端 (Go)。核心功能：数据源管理、数据集管理、拖拽式图表构建、分享功能。

## 关键约束 (CRITICAL)

- **字段命名**: 前后端 JSON 通信使用 snake_case (如 `table_name`, not `tableName`)
- **禁止 any**: TypeScript 使用 `unknown` 代替
- **错误处理**: Go handler 禁止空 catch 块

## 开发资源

| 文档 | 说明 |
|------|------|
| [docs/setup.md](docs/setup.md) | 环境搭建、运行命令 |
| [docs/architecture.md](docs/architecture.md) | 目录结构、技术栈 |
| [docs/coding-style.md](docs/coding-style.md) | 代码风格指南 |
| [docs/api-spec.md](docs/api-spec.md) | API 规范 |
| [docs/todo.md](docs/todo.md) | 开发任务清单 |

## 已知限制

- 前端无 ESLint/Prettier 配置
- 后端无 golangci-lint 配置
- 缺少端到端测试
