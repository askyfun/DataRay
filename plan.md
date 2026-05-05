# DataRay 图表查询与可视化语义演进计划

## 今日完成总结

- [x] 将 `plan.md` 从说明文档重构为可执行待办清单。
- [x] 明确本轮改造遵循的原则：
  - [x] Chart 模块负责图表语义。
  - [x] Query 模块负责查询语义。
  - [x] 改代码必须走 TDD。
  - [x] 遇到语义不清先和用户确认。
- [x] 与用户确认过滤条件 `logic` 语义：
  - [x] `logic` 表示“当前条件与前一个条件的连接方式”。
  - [x] 第一个条件忽略 `logic`。
  - [x] 当前阶段只支持线性表达式，不支持括号分组。
- [x] 完成图表查询目标契约草案文档：
  - [x] `docs/api-spec.md`：补充图表查询契约演进草案。
  - [x] `docs/api.md`：补充图表查询兼容协议说明与目标契约草案。
  - [x] `docs/chart-builder-plan.md`：补充契约演进方向。
  - [x] `docs/architecture.md`：补充图表查询与可视化语义分层。
  - [x] `AGENTS.md`：更新文档索引说明。
- [x] 完成过滤条件 `AND/OR` 问题修复（TDD）：
  - [x] 先新增失败测试：`TestBunQueryBuilder_WithFilterLogicOr`。
  - [x] 先新增失败测试：`TestBunQueryBuilder_WithMixedFilterLogic`。
  - [x] 修复 `backend/internal/query/bun_builder.go` 中 WHERE 拼接逻辑。
  - [x] 通过针对性测试回归。
  - [x] 通过 `go test ./internal/query` 回归。

## 当前总体目标

- [ ] 建立统一的图表语义契约，明确 Chart 模块负责视觉语义，Query 模块负责查询语义。
- [ ] 在兼容现有 `dims` / `metrics` 请求的前提下，引入更强的 `ChartSpec` / `QuerySpec` / `QueryAST`。
- [ ] 统一前后端响应结构和文档，支持未来复杂图表扩展。
- [ ] 所有代码改造继续使用 TDD：先写失败测试，再实现，再回归验证。

## 下一次开工优先顺序

### P0：继续修当前确定问题

- [x] 第三步：修复散点图前后端语义不一致，采用 TDD。
  - [x] 先读当前散点图前端渲染逻辑。
  - [x] 先读当前散点图后端 `ScatterProcessor` 逻辑。
  - [x] 明确当前契约到底是”两个 metric”还是”一个 dim + 一个 metric”。→ 确认方案 A：两个 metric。
  - [x] 与用户确认后改代码。
  - [x] 先补失败测试，复现前后端不一致问题。
  - [x] 最小改动修复实现。
  - [x] 跑针对性测试。
  - [x] 跑 `go test ./internal/query` 回归。
  - [x] 同步更新文档。

- [x] 第四步：修复分页与排序规则，采用 TDD。
  - [x] 梳理前端表格排序请求是否真正传给后端。
  - [x] 梳理后端分页时为什么跳过 `ORDER BY`。
  - [x] 先明确目标规则：分页查询是否必须带排序、默认排序字段是什么。
  - [x] 先补失败测试。
  - [x] 最小改动修复实现。
  - [x] 跑针对性测试。
  - [x] 跑 query / service 相关回归测试。
  - [x] 同步更新文档。（行为修复，文档已正确描述 sort 参数，无需更新）

### P1：后端内部模型升级

- [x] 新增内部 `ChartSpec` 类型。
  - [x] 定义基础结构：`chart_type`、`dimension_groups`、`metric_groups`、`style`、`query_options`。
  - [x] 为字段绑定预留附加属性：`label`、`granularity`、`agg`、`alias`、`unit`。
  - [x] 为该结构补单元测试或序列化测试。
- [x] 新增旧请求 → `ChartSpec` 的 adapter。
  - [x] 先补 adapter 测试。
  - [x] 保证旧 `dims` / `metrics` 能映射到默认维度组和指标组。
- [x] 新增 `QuerySpec` 类型。
  - [x] 定义结构化维度表达式。
  - [x] 定义结构化指标表达式。
  - [x] 定义过滤、排序、分页、limit 能力。
- [x] 新增 `ChartSpec` → `QuerySpec` 的转换。
  - [x] 先补转换测试。
  - [x] 保证基础图表在兼容模式下生成 SQL 不变。

### P2：增强 QueryAST

- [ ] 将维度从字符串升级为结构化表达式。
  - [ ] 字段名。
  - [ ] 展示名。
  - [ ] 时间粒度。
  - [ ] 分桶。
  - [ ] 格式化信息。
- [ ] 将指标升级为结构化表达式。
  - [ ] 聚合。
  - [ ] 别名。
  - [ ] 单位。
  - [ ] 数值格式。
  - [ ] 是否累计 / 占比。
- [ ] 为日期维度增加统一时间粒度能力。
  - [ ] PostgreSQL 方言测试。
  - [ ] MySQL 方言测试。
  - [ ] ClickHouse 方言测试。
- [ ] 增强过滤能力。
  - [ ] 过滤组结构设计。
  - [ ] 时间范围过滤。
  - [ ] 空值过滤。
  - [ ] 字段类型感知。
- [ ] 引入 `QueryPlanner`。
  - [ ] 定义 `QuerySpec -> QueryAST` 规划流程。
  - [ ] 为基础图表补规划测试。

### P3：前端 ChartBuilder 升级

- [ ] 前端新增 `ChartDefinition` registry。
  - [ ] 折线图定义。
  - [ ] 透视表定义。
  - [ ] 双轴图定义。
  - [ ] 饼图定义。
- [ ] ChartBuilder 根据图表定义动态渲染维度组和指标组。
- [ ] 支持维度字段属性配置。
  - [ ] 日期精度。
  - [ ] 展示名。
- [ ] 支持指标属性配置。
  - [ ] 聚合。
  - [ ] 别名。
  - [ ] 单位。
  - [ ] 格式化。
- [ ] 支持图表样式配置。
  - [ ] 颜色。
  - [ ] 线型。
  - [ ] 表格行高。
- [ ] 支持图表特有查询配置。
  - [ ] 饼图低比例合并为“其他”。

### P4：统一响应结构落地

- [ ] 后端返回统一响应 Envelope，并兼容旧的 `select_sql` / `count_sql` 顶层字段。
  - [ ] 定义 `shape`。
  - [ ] 定义 `meta`。
  - [ ] 定义 `fields`。
  - [ ] 定义 `sql`。
- [ ] 前端 Store 按 `shape` 解析响应。
- [ ] SQL 查看弹窗改为从统一 `sql` 字段读取。

## 文档同步待办

- [x] 更新 `docs/api.md`。
- [x] 更新 `docs/api-spec.md`。
- [x] 更新 `docs/architecture.md`。
- [x] 更新 `docs/chart-builder-plan.md`。
- [x] 更新 `AGENTS.md` 中与图表构建、查询契约相关的说明。
- [ ] 后续每做完一个代码阶段，再同步更新对应文档和示例。

## 测试与验收规则

- [x] 每一个 bug 修复都先补失败测试。
- [ ] 每一个契约升级都补前后端类型/单元测试。
- [ ] 后端关键阶段提交前运行目标测试，必要时运行 `go test -race ./...`。
- [ ] 前端改造阶段运行相关 Vitest 测试。
- [ ] 每完成一个阶段，立即把对应任务标记为 `[x]`。

## 今天停在这里时的状态

- [x] 文档契约草案已完成。
- [x] `AND/OR` 过滤问题已完成并验证。
- [x] 散点图语义一致性已完成（方案 A：两个 metric，前端已修复）。
- [x] 分页与排序规则已完成（后端 ORDER BY 不再被分页跳过，前端 sort 参数已接入请求链路）。
- [x] `ChartSpec` / `QuerySpec` 类型定义与 adapter 已完成（`chart_spec.go`）。
- [ ] `ChartSpec` → `QuerySpec` 转换尚未接入 service 层（目前仅类型和 adapter 就绪）。
