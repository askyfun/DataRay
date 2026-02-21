# Backend 优化任务清单

## 优先级 P0 (安全 - 必须修复)

- [x] 1. SQL 注入防护 - buildSQL() 和所有 SQL 拼接使用参数化查询
- [x] 2. 修复弱 Token 生成 - 使用 crypto/rand 替代可预测算法
- [x] 3. 外部数据源连接添加 timeout - 防止连接泄漏

## 优先级 P1 (稳定性)

- [x] 4. 添加 Graceful Shutdown - 优雅关闭服务器
- [x] 5. 请求参数验证 - 统一验证中间件/函数

## 优先级 P2 (可维护性)

- [x] 6. 提取重复代码 - datasource lookup 逻辑抽取为独立函数
- [ ] 7. Handler 拆分 - 从 main.go 拆分到 handler/ 目录 (取消 - 需要更大范围重构)

## 优先级 P3 (性能)

- [x] 8. 添加分页 - list 接口支持 limit/offset
- [x] 9. 数据库索引 - 为 token, datasource_id 等添加索引

## 优先级 P4 (代码质量)

- [x] 10. 添加 Request ID 追踪
- [ ] 11. 结构化日志 (取消 - 需要引入新依赖)

---

## 已完成优化摘要

### 1. SQL 注入防护
- 添加 `isValidIdentifier()` 函数验证表名和列名

### 2. Token 生成
- 使用 `crypto/rand` 生成安全 token

### 3. 数据源连接超时
- 所有外部数据源连接添加 30s 超时

### 4. Graceful Shutdown
- 使用 `http.Server` + 信号处理
- 优雅关闭服务器

### 5. 请求参数验证
- 添加 `ValidationError` 结构
- 添加 `bindAndValidate()` 和 `validateStruct()` 函数
- 添加 `respondValidationError()` 响应函数

### 6. 分页支持
- List 接口支持 `?limit=100&offset=0`
- 默认 limit=100, 最大 1000

### 7. 数据库索引
- `idx_dataset_datasource_id`
- `idx_chart_dataset_id`
- `idx_share_token`

### 8. Request ID 追踪
- 添加 `requestIDMiddleware()`
- 每个请求自动生成/传递 X-Request-ID
- 便于日志追踪和问题排查
