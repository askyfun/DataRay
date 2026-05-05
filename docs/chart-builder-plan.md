# Chart Builder 增强实现计划

## 一、架构设计

### 1.1 数据流

```
前端配置
  ├── dims: [dim1, dim2...]
  ├── metrics: [metric1, metric2...]
  ├── filters: [filter1, filter2...]
  └── chartType: string

  ↓ POST /api/charts/query

后端处理
  ├── QueryBuilder: 生成基础 SQL
  ├── 执行查询
  ├── 根据 chartType 二次加工
  └── 返回处理后的数据

  ↓

前端渲染
  └── ECharts / Table 组件展示
```

### 1.2 API 设计

**请求**: `POST /api/charts/query`

```typescript
interface ChartQueryRequest {
  dataset_id: number;
  chart_type: string;
  dims: string[];           // 维度字段列表
  metrics: MetricConfig[];   // 指标配置
  filters: FilterConfig[];  // 过滤条件
  pagination?: Pagination;   // 分页配置
  sort?: SortConfig;        // 排序配置
}

interface MetricConfig {
  field: string;
  agg: 'sum' | 'avg' | 'count' | 'max' | 'min';
  alias?: string;
}

interface FilterConfig {
  field: string;
  op: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'in' | 'between';
  value: any;
  valueEnd?: any;
  logic: 'and' | 'or';
}

interface Pagination {
  page: number;
  pageSize: number;
}

interface SortConfig {
  field: string;
  order: 'asc' | 'desc';
}
```

**响应**: 根据 chartType 返回不同结构

```typescript
// Table 响应
interface TableResponse {
  columns: string[];
  data: Record<string, any>[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}

// Pie 响应
interface PieResponse {
  data: Array<{
    name: string;
    value: number;
    percentage: number;  // 计算后的百分比
  }>;
  // 长尾合并后
  other?: {
    name: string;
    value: number;
    percentage: number;
  };
}

// Bar/Line/Area 响应
interface AxisResponse {
  xAxis: string[];
  series: Array<{
    name: string;
    data: number[];
  }>;
}

// Scatter 响应
interface ScatterResponse {
  data: Array<[number, number]>;  // [x, y]
}
```

---

## 一点五、契约演进方向

### 1.5.1 当前模型的限制

当前模型以 `dims` / `metrics` 作为统一输入，适合基础图表，但存在以下限制：

- 无法表达日期维度精度，如按小时、天、周、月聚合。
- 无法表达多组维度，例如透视表的“行/列”。
- 无法表达多组指标，例如双轴图的“主轴/次轴”。
- 无法承载图表样式配置和图表特有查询配置。
- 前后端对图表语义和响应结构缺少统一契约。

### 1.5.2 目标模型

后续演进目标是由 Chart 模块维护结构化 `ChartSpec`，再转换为 Query 模块使用的 `QuerySpec`。

#### ChartDefinition

ChartDefinition 用于声明一种图表需要哪些输入：

- 图表类型
- 维度组定义
- 指标组定义
- 样式 schema
- 查询配置 schema
- 响应 shape
- 字段数量和类型约束

#### ChartSpec

ChartSpec 用于表达图表实例绑定的真实字段和配置：

- `chart_type`
- `dimension_groups`
- `metric_groups`
- `style`
- `query_options`

字段绑定支持附加属性，例如：

- 维度：`field`、`label`、`granularity`
- 指标：`field`、`label`、`agg`、`alias`、`unit`

#### QuerySpec

QuerySpec 只表达查询语义，不直接表达“折线图”“饼图”等视觉类型，主要包括：

- 结构化维度表达式
- 结构化指标表达式
- 过滤组
- 排序
- 分页
- 时间粒度
- 分桶
- limit / topN

### 1.5.3 首批图表示例

- 折线图：`x_axis` 维度组 + `values` 指标组
- 透视表：`rows` / `columns` 维度组 + `values` 指标组
- 双轴图：`primary_axis` / `secondary_axis` 维度组 + `primary_values` / `secondary_values` 指标组
- 饼图：`category` 维度组 + `value` 指标组，并支持 `merge_other_below_ratio`

### 1.5.4 响应统一目标

响应将逐步统一为：

- `shape`：响应形态
- `data`：图表数据主体
- `meta`：分页、粒度、“其他”合并等元信息
- `fields`：字段元数据
- `sql`：调试 SQL

兼容期内，旧的响应结构仍然保留。

---


## 二、后端实现

### 2.1 新增文件结构

```
backend/internal/
├── query/
│   ├── builder.go      # QueryBuilder 核心
│   ├── executor.go     # 查询执行器
│   ├── processor.go    # 图表数据处理器
│   └── types.go        # 类型定义
```

### 2.2 QueryBuilder 核心逻辑

```go
// 基础 SQL 生成
WITH t1 AS (
  SELECT
    dim1, dim2, ...,
    SUM(metric1) as metric1,
    AVG(metric2) as metric2,
    COUNT(*) as metric3
  FROM source_table
  WHERE filter1 AND filter2
  GROUP BY dim1, dim2
)
SELECT * FROM t1
```

### 2.3 图表处理器

| ChartType | 处理器 | 功能 |
|-----------|-------|------|
| table | TableProcessor | 分页、total count |
| pie | PieProcessor | 百分比计算、长尾合并 |
| bar | AxisProcessor | 透传数据 |
| line | AxisProcessor | 透传数据 |
| area | AxisProcessor | 透传数据 |
| scatter | ScatterProcessor | 透传数据 |

### 2.4 Pie 特殊处理

```go
func (p *PieProcessor) Process(rows []map[string]interface{}) (*PieResponse, error) {
  // 1. 计算 total
  // 2. 计算每个维度的百分比
  // 3. 长尾合并: 保留 Top N，其余合并为 "Other"
  //    - 默认保留 Top 10
  //    - 可配置 threshold 百分比
}
```

### 2.5 Table 分页处理

```go
func (p *TableProcessor) Process(rows []map[string]interface{}, pagination Pagination) (*TableResponse, error) {
  // 1. 获取 total count (需要单独查询 COUNT)
  // 2. 分页 slice
  // 3. 返回分页元数据
}
```

---

## 三、前端实现

### 3.1 API 调用 (api/index.ts)

```typescript
interface ChartQueryRequest {
  dataset_id: number;
  chart_type: string;
  dims: string[];
  metrics: Array<{ field: string; agg: string; alias?: string }>;
  filters: FilterConfig[];
  pagination?: { page: number; pageSize: number };
  sort?: { field: string; order: 'asc' | 'desc' };
}
```

### 3.2 Store 更新

- 新增 `executeChartQuery` action
- 处理响应，根据 chartType 解析不同结构

### 3.3 ChartBuilder 更新

- 支持配置分页参数
- Pie 图表显示百分比开关
- 长尾合并阈值配置

### 3.3.1 当前落地状态（2026-05-05）

- 已落地 `ChartDefinition` registry 第一阶段，位置：`frontend/src/components/ChartBuilder/chartDefinitions.ts`
- `ChartBuilder` 已根据图表定义动态渲染查询配置行标签与空态文案，不再只写死“维度/指标”两行
- 当前已覆盖：表格、柱状图、折线图、面积图、饼图、散点图、透视表的基础定义
- 切换图表类型时，前端会按定义补齐最小维度组/指标组数量，便于后续扩展多组语义
- 当前仍保持旧请求协议兼容：查询请求继续扁平化为 `dims` / `metrics` / `filters` / `sort` / `pagination`
- 多 group 基础绑定已接通：不同槽位的添加/删除/重排可作用于对应 group，页面级回归测试入口已补
- 字段与样式配置的最小模型已落地：维度 `label`、指标 `agg/alias/unit/format`、颜色、平滑曲线、表格行尺寸均可保存并恢复
- 饼图 `query_options` 已形成最小前后端闭环：前端可配置“低比例并入其他”，后端 `PieProcessor` 已消费该阈值
- 当前限制：维度 `granularity`、更完整的配置 UI、以及后端对 `unit/format/style` 等配置的真正消费仍属于后续阶段

### 3.4 TableChart 组件

- 使用 Ant Design Table
- 支持分页组件
- 显示 total count

---

## 四、可视化类型详细设计

### 4.1 Table (表格)

**后端**:
- 查询时使用 `COUNT(*)` 获取总数
- SQL 添加 `LIMIT pageSize OFFSET (page-1)*pageSize`

**前端**:
- Ant Design Table 组件
- 分页器显示 total

### 4.2 Pie (饼图)

**后端**:
- 聚合计算后，对第一维度分组
- 计算每个分组的百分比
- 长尾合并: 行数 > 20 时，将后面的行合并为 "Other"

**前端**:
- 显示 percentage
- 支持切换显示模式 (数值/百分比)

### 4.3 Bar/Line/Area (柱状/折线/面积图)

**后端**:
- 基础查询，第一维度作为 X 轴
- 多个指标作为 series

**前端**:
- ECharts 标准配置

### 4.4 Scatter (散点图)

**后端**:
- 支持 2 个指标作为 X/Y 坐标

**前端**:
- ECharts scatter 配置

### 4.5 Pivot (透视表)

**后端**:
- 行列转换 (pivot)
- 多维度组合

**前端**:
- 使用 Ant Design Table 或专用透视表组件

---

## 五、SQL 生成规则

### 5.1 基础模板

```sql
WITH base AS (
  SELECT
    {dims},
    {metric_selects}
  FROM ({base_query}) AS t
  {where_clause}
  {group_by_clause}
)
SELECT * FROM base
{order_by_clause}
{pagination_clause}
```

### 5.2 指标聚合映射

| 前端 agg | SQL 聚合 |
|---------|---------|
| sum | SUM(field) |
| avg | AVG(field) |
| count | COUNT(*) |
| max | MAX(field) |
| min | MIN(field) |

### 5.3 过滤条件映射

| 前端 op | SQL |
|--------|-----|
| eq | = |
| neq | <> |
| gt | > |
| gte | >= |
| lt | < |
| lte | <= |
| like | LIKE |
| in | IN |
| between | BETWEEN AND |

---

## 六、实施步骤

### Phase 1: 后端核心 (Day 1)
- [ ] 创建 query 包
- [ ] 实现 QueryBuilder
- [ ] 实现基础 SQL 生成

### Phase 2: 图表处理器 (Day 2)
- [ ] TableProcessor (分页)
- [ ] PieProcessor (百分比 + 长尾)
- [ ] AxisProcessor (通用)

### Phase 3: API 集成 (Day 3)
- [ ] 注册 `/api/charts/query` 路由
- [ ] Handler 实现
- [ ] 单元测试

### Phase 4: 前端集成 (Day 4)
- [ ] API 调用更新
- [ ] Store action
- [ ] ChartBuilder 适配

### Phase 5: 可视化增强 (Day 5)
- [ ] Table 分页组件
- [ ] Pie 百分比显示
- [ ] 长尾合并配置

---

## 七、配置项

### 7.1 Pie 图表配置

```typescript
interface PieChartConfig {
  threshold: number;      // 长尾阈值，默认 20
  showPercentage: boolean; // 显示百分比，默认 true
  otherLabel: string;     // 其他标签，默认 "Other"
}
```

### 7.2 Table 图表配置

```typescript
interface TableChartConfig {
  pageSize: number;       // 每页行数，默认 10
  showPagination: boolean; // 显示分页，默认 true
  showTotal: boolean;      // 显示总数，默认 true
}
```
