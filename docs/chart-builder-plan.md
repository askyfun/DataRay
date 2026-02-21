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
