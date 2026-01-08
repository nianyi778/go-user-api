// Package response 提供统一的 HTTP 响应格式
//
// 本包定义了标准化的 JSON 响应结构，确保 API 返回格式的一致性。
// 所有 API 响应都应该使用本包提供的函数来构建响应。
//
// 响应格式示例：
//
// 成功响应：
//
//	{
//	    "code": 0,
//	    "message": "success",
//	    "data": { ... }
//	}
//
// 错误响应：
//
//	{
//	    "code": 10001,
//	    "message": "用户名已存在",
//	    "data": null
//	}
//
// 分页响应：
//
//	{
//	    "code": 0,
//	    "message": "success",
//	    "data": {
//	        "list": [...],
//	        "pagination": {
//	            "page": 1,
//	            "page_size": 20,
//	            "total": 100,
//	            "total_pages": 5
//	        }
//	    }
//	}
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	// Code 业务状态码，0 表示成功，非 0 表示失败
	Code int `json:"code"`
	// Message 响应消息，成功时为 "success"，失败时为错误描述
	Message string `json:"message"`
	// Data 响应数据，可以是任意类型
	Data interface{} `json:"data"`
}

// Pagination 分页信息
type Pagination struct {
	// Page 当前页码（从 1 开始）
	Page int `json:"page"`
	// PageSize 每页数量
	PageSize int `json:"page_size"`
	// Total 总记录数
	Total int64 `json:"total"`
	// TotalPages 总页数
	TotalPages int `json:"total_pages"`
}

// PageData 分页数据响应
type PageData struct {
	// List 数据列表
	List interface{} `json:"list"`
	// Pagination 分页信息
	Pagination Pagination `json:"pagination"`
}

// 常用业务状态码定义
const (
	// CodeSuccess 成功
	CodeSuccess = 0
	// CodeBadRequest 请求参数错误
	CodeBadRequest = 10001
	// CodeUnauthorized 未授权/未登录
	CodeUnauthorized = 10002
	// CodeForbidden 禁止访问
	CodeForbidden = 10003
	// CodeNotFound 资源不存在
	CodeNotFound = 10004
	// CodeConflict 资源冲突
	CodeConflict = 10005
	// CodeInternalError 服务器内部错误
	CodeInternalError = 10006
	// CodeValidationError 数据验证错误
	CodeValidationError = 10007
	// CodeTooManyRequests 请求过于频繁
	CodeTooManyRequests = 10008
)

// 常用消息定义
const (
	MsgSuccess           = "success"
	MsgBadRequest        = "请求参数错误"
	MsgUnauthorized      = "请先登录"
	MsgForbidden         = "没有权限访问"
	MsgNotFound          = "资源不存在"
	MsgConflict          = "资源已存在"
	MsgInternalError     = "服务器内部错误"
	MsgValidationError   = "数据验证失败"
	MsgTooManyRequests   = "请求过于频繁，请稍后再试"
	MsgInvalidToken      = "无效的令牌"
	MsgTokenExpired      = "令牌已过期"
	MsgUserNotFound      = "用户不存在"
	MsgWrongPassword     = "密码错误"
	MsgUserAlreadyExists = "用户已存在"
	MsgEmailAlreadyUsed  = "邮箱已被使用"
)

// JSON 发送 JSON 响应
func JSON(c *gin.Context, httpCode int, code int, message string, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Success 发送成功响应
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, CodeSuccess, MsgSuccess, data)
}

// SuccessWithMessage 发送带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, CodeSuccess, message, data)
}

// Created 发送创建成功响应
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, CodeSuccess, MsgSuccess, data)
}

// NoContent 发送无内容响应（用于删除操作）
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SuccessWithPagination 发送分页数据响应
func SuccessWithPagination(c *gin.Context, list interface{}, page, pageSize int, total int64) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	data := PageData{
		List: list,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
	Success(c, data)
}

// Error 发送错误响应
func Error(c *gin.Context, httpCode int, code int, message string) {
	JSON(c, httpCode, code, message, nil)
}

// ErrorWithData 发送带数据的错误响应（用于验证错误时返回详细信息）
func ErrorWithData(c *gin.Context, httpCode int, code int, message string, data interface{}) {
	JSON(c, httpCode, code, message, data)
}

// BadRequest 发送请求参数错误响应
func BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = MsgBadRequest
	}
	Error(c, http.StatusBadRequest, CodeBadRequest, message)
}

// BadRequestWithData 发送带数据的请求参数错误响应
func BadRequestWithData(c *gin.Context, message string, data interface{}) {
	if message == "" {
		message = MsgBadRequest
	}
	ErrorWithData(c, http.StatusBadRequest, CodeBadRequest, message, data)
}

// Unauthorized 发送未授权响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = MsgUnauthorized
	}
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden 发送禁止访问响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = MsgForbidden
	}
	Error(c, http.StatusForbidden, CodeForbidden, message)
}

// NotFound 发送资源不存在响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = MsgNotFound
	}
	Error(c, http.StatusNotFound, CodeNotFound, message)
}

// Conflict 发送资源冲突响应
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = MsgConflict
	}
	Error(c, http.StatusConflict, CodeConflict, message)
}

// ValidationError 发送验证错误响应
func ValidationError(c *gin.Context, message string, errors interface{}) {
	if message == "" {
		message = MsgValidationError
	}
	ErrorWithData(c, http.StatusBadRequest, CodeValidationError, message, errors)
}

// InternalError 发送服务器内部错误响应
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = MsgInternalError
	}
	Error(c, http.StatusInternalServerError, CodeInternalError, message)
}

// TooManyRequests 发送请求过于频繁响应
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = MsgTooManyRequests
	}
	Error(c, http.StatusTooManyRequests, CodeTooManyRequests, message)
}

// Abort 中止请求并发送错误响应
// 用于中间件中终止请求处理
func Abort(c *gin.Context, httpCode int, code int, message string) {
	c.Abort()
	JSON(c, httpCode, code, message, nil)
}

// AbortWithUnauthorized 中止请求并发送未授权响应
func AbortWithUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = MsgUnauthorized
	}
	Abort(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

// AbortWithForbidden 中止请求并发送禁止访问响应
func AbortWithForbidden(c *gin.Context, message string) {
	if message == "" {
		message = MsgForbidden
	}
	Abort(c, http.StatusForbidden, CodeForbidden, message)
}

// AbortWithTooManyRequests 中止请求并发送请求过于频繁响应
func AbortWithTooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = MsgTooManyRequests
	}
	Abort(c, http.StatusTooManyRequests, CodeTooManyRequests, message)
}
