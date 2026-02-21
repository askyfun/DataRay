# DataRay API 规范

## 1. 统一响应格式

所有 API 响应必须使用以下统一格式：

```json
{
  "code": 20000,
  "msg": "success",
  "trace": "xxxxxxx",
  "data": {}
}
```

### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | int | 是 | 状态码，详见状态码表 |
| msg | string | 是 | 响应信息，通常不需要显示给用户 |
| trace | string | 是 | 链路追踪 ID，从请求头 X-Request-ID 获取 |
| data | object/array | 是 | 主要数据，必须是对象格式，方便后续扩展 |

### 状态码表

| 状态码 | 说明 |
|--------|------|
| 20000 | 成功 |
| 20100 | 请求参数错误 |
| 20200 | 认证/授权错误 |
| 20300 | 资源不存在 |
| 20400 | 业务逻辑错误 |
| 20500 | 第三方服务错误 |
| 50000 | 服务端内部错误 |

## 2. 请求头要求

### 请求头

| 请求头 | 必填 | 说明 |
|--------|------|------|
| X-Request-ID | 否 | 链路追踪 ID，由客户端生成，响应会原样返回 |

### 响应头

| 响应头 | 说明 |
|--------|------|
| X-Request-ID | 链路追踪 ID，原样返回请求中的值 |

## 3. 后端实现

### 目录结构

```
backend/
├── internal/
│   ├── response/          # 统一响应封装
│   │   └── response.go
│   └── idls/             # 接口协议定义
│       ├── datasource.go
│       ├── dataset.go
│       ├── chart.go
│       └── share.go
```

### 响应封装示例

```go
// 成功响应
response.Success(c, data)

// 失败响应
response.Error(c, 20100, "invalid parameter")
```

## 4. 前端实现

### 目录结构

```
frontend/src/
├── lib/
│   └── api/              # 统一响应处理
│       └── client.ts
└── idls/                # 接口协议定义
    ├── datasource.ts
    ├── dataset.ts
    ├── chart.ts
    └── share.ts
```

### 响应处理示例

```typescript
// 统一处理响应
const result = await apiClient.get<ResponseData<Datasource[]>>('/api/datasources');
if (result.code === 20000) {
  // 成功处理
}
```

## 5. IDL 定义规范

### 命名规范

- 文件名使用领域名称，如 `datasource.go`、`dataset.go`
- 同一模块的接口协议放在同一个文件中
- 类型定义使用 PascalCase

### 文件结构

```go
// 请求类型定义
type CreateDatasourceRequest struct {
    Name string `json:"name"`
    Type string `json:"type"`
    // ...
}

// 响应类型定义
type DatasourceResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    // ...
}

// 列表响应（带分页）
type DatasourceListResponse struct {
    Items  []DatasourceResponse `json:"items"`
    Total  int                  `json:"total"`
    Page   int                  `json:"page"`
    Size   int                  `json:"size"`
}
```

## 6. 错误处理

### 前端错误处理

```typescript
// 响应拦截器自动处理
apiClient.interceptors.response.use(
  (response) => {
    const { code, msg, data } = response.data;
    if (code !== 20000) {
      // 显示错误信息
      message.error(msg);
      return Promise.reject(new Error(msg));
    }
    return response;
  }
);
```

### 状态码与错误提示映射

| 状态码 | 用户提示 |
|--------|----------|
| 20100 | 请求参数有误，请检查输入 |
| 20200 | 登录状态已失效，请重新登录 |
| 20300 | 请求的资源不存在 |
| 20400 | 操作失败，请稍后重试 |
| 20500 | 服务暂时不可用 |
| 50000 | 服务器内部错误 |
