# DataRay 开发规范

**生成时间**: 2026-02-26  
**项目**: DataRay - 拖拽式 BI 可视化分析平台

## 项目概览

Monorepo 结构，包含前端 (React/TypeScript) 和后端 (Go)。核心功能：数据源管理、数据集管理、拖拽式图表构建、分享功能。

## 关键约束 (CRITICAL)

- **字段命名**: 前后端 JSON 通信使用 snake_case (如 `table_name`, not `tableName`)
- **禁止 any**: TypeScript 使用 `unknown` 代替
- **单元测试**: 后端新增功能或修复 bug 后必须补充对应的单元测试；大型重构后必须重跑单元测试确保功能无损
- **错误处理**: Go handler 禁止空 catch 块

## Sisyphus 工作流

### 意图识别与路由

| 表面形式 | 真实意图 | 路由方式 |
|----------|----------|----------|
| "explain X", "how does Y work" | 研究/理解 | explore → 回答 |
| "implement X", "add Y", "create Z" | 实现（显式） | plan → delegate |
| "look into X", "check Y" | 调查 | explore → 报告 |
| "I'm seeing error X" | 修复 | 诊断 → 最小化修复 |
| "refactor", "improve" | 开放性变更 | 评估 → 建议方案 |

### Agent 选择

| Agent | 用途 | 使用场景 |
|-------|------|----------|
| `explore` | 代码库上下文搜索 | 搜索现有模式、查找文件 |
| `librarian` | 外部文档/参考 | 不熟悉的库、官方文档 |
| `oracle` | 架构评审、复杂调试 | 2+ 次修复失败、架构决策 |
| `plan` | 任务规划 | 2+ 步任务、范围不清 |
| `momus` | 计划评审 | 工作计划质量审查 |

### Category + Skills 委托

| Category | 用途 | 典型场景 |
|----------|------|----------|
| `visual-engineering` | 前端 UI/UX | React 组件、样式、动画 |
| `ultrabrain` | 复杂逻辑 | 深度算法、复杂业务逻辑 |
| `deep` | 深度问题解决 | 需要深入理解的问题 |
| `artistry` | 创造性方法 | 非传统方案 |
| `quick` | 简单修改 | 拼写错误、小修小补 |

**Skills (项目级优先)**:
- `frontend-philosophy` - UI 哲学
- `code-review` - 代码审查
- `code-philosophy` - 数据流哲学
- `plan-protocol` - 计划管理
- `plan-review` - 计划评审
- `golang-pro` - Go 开发
- `web-design-guidelines` - Web 界面指南

### 执行规则

**TODO 管理**:
1. 非平凡任务前创建 TODO
2. 每步标记 `in_progress`
3. 完成后立即标记 `completed`
4. 范围变更时更新 TODO

**并行执行**: 独立任务并行运行，explore/librarian 使用 `run_in_background=true`

**验证保证**:
- 构建: 退出码 0
- 测试: 全部通过
- 手动验证: 功能正常

### 必用协议

**100% 确定要求**:
- 不确定时先探索，不要猜测
- 有歧义必须问用户
- 不理解代码不要盲目修改

**PLAN AGENT 调用条件**:
- 任务有 2+ 步
- 范围不清晰
- 需要架构决策

### 零容忍失败

**禁止**:
- 部分实现 ("简化版本")
- 未经授权的范围变更
- 删除失败测试
- 类型压制 (`as any`, `@ts-ignore`)
- 空 catch 块

## 关键约束

**代码可测试性**:
- 设计代码必须具备可测试性
- 开发完代码必须增加单元测试
- 确保代码可以测试通过
- 禁止提交无法测试的代码

**文档同步更新**:
- 大型业务逻辑调整必须更新相关 .md 文档
- 架构调整必须更新相关 .md 文档
- 文档范围包括: docs/*.md, README.md, AGENTS.md 等
- 保持代码与文档状态一致

---

## 开发资源

| 文档 | 说明 |
|------|------|
| [docs/setup.md](docs/setup.md) | 环境搭建、运行命令 |
| [docs/architecture.md](docs/architecture.md) | 目录结构、技术栈 |
| [docs/coding-style.md](docs/coding-style.md) | 代码风格指南 |
|| [docs/api.md](docs/api.md) | API 接口文档 (统一规范) |
| [docs/todo.md](docs/todo.md) | 开发任务清单 |

## 已知限制

- 前端无 ESLint/Prettier 配置
- 后端无 golangci-lint 配置
- 缺少端到端测试
