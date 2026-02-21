package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 状态码定义
const (
	// 2xx 成功
	CodeSuccess = 20000

	// 4xx 客户端错误
	CodeBadRequest    = 20100 // 请求参数错误
	CodeUnauthorized  = 20200 // 认证/授权错误
	CodeNotFound      = 20300 // 资源不存在
	CodeBusinessError = 20400 // 业务逻辑错误
	CodeThirdPartyErr = 20500 // 第三方服务错误

	// 5xx 服务端错误
	CodeInternalError = 50000 // 服务端内部错误
)

// Response 统一响应结构
type Response struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Trace string      `json:"trace"`
	Data  interface{} `json:"data"`
}

// 获取请求追踪 ID
func getTraceID(c *gin.Context) string {
	traceID := c.GetHeader("X-Request-ID")
	if traceID == "" {
		traceID = c.GetString("request_id")
	}
	return traceID
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:  CodeSuccess,
		Msg:   "success",
		Trace: getTraceID(c),
		Data:  data,
	})
}

// SuccessWithMsg 带自定义消息的成功响应
func SuccessWithMsg(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:  CodeSuccess,
		Msg:   msg,
		Trace: getTraceID(c),
		Data:  data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:  code,
		Msg:   msg,
		Trace: getTraceID(c),
		Data:  nil,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:  code,
		Msg:   msg,
		Trace: getTraceID(c),
		Data:  data,
	})
}

// BadRequest 请求参数错误
func BadRequest(c *gin.Context, msg string) {
	Error(c, CodeBadRequest, msg)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, msg string) {
	Error(c, CodeUnauthorized, msg)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, msg string) {
	Error(c, CodeNotFound, msg)
}

// BusinessError 业务逻辑错误
func BusinessError(c *gin.Context, msg string) {
	Error(c, CodeBusinessError, msg)
}

// ThirdPartyError 第三方服务错误
func ThirdPartyError(c *gin.Context, msg string) {
	Error(c, CodeThirdPartyErr, msg)
}

// InternalError 服务端内部错误
func InternalError(c *gin.Context, msg string) {
	Error(c, CodeInternalError, msg)
}

// PageResult 分页结果
type PageResult struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// SuccessWithPage 带分页的成功响应
func SuccessWithPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	Success(c, PageResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
