// Package errors 提供应用程序统一的错误处理机制
//
// 本包定义了自定义错误类型 AppError，包含错误码、HTTP 状态码和错误消息，
// 用于在整个应用程序中保持一致的错误处理和响应格式。
//
// 使用示例：
//
//	// 返回预定义错误
//	return errors.ErrUserNotFound
//
//	// 创建自定义错误
//	return errors.New(errors.CodeBadRequest, http.StatusBadRequest, "无效的参数")
//
//	// 包装原始错误
//	return errors.Wrap(err, errors.CodeDatabaseError, "数据库操作失败")
package errors

import (
	"fmt"
	"net/http"
)

// 错误码定义
// 错误码格式说明：
// - 1xxxx: 认证相关错误
// - 2xxxx: 用户相关错误
// - 3xxxx: 数据验证错误
// - 4xxxx: 资源相关错误
// - 5xxxx: 服务器内部错误
const (
	// 通用错误码
	CodeSuccess       = 0     // 成功
	CodeUnknown       = 10000 // 未知错误
	CodeBadRequest    = 10001 // 请求参数错误
	CodeUnauthorized  = 10002 // 未授权
	CodeForbidden     = 10003 // 禁止访问
	CodeNotFound      = 10004 // 资源不存在
	CodeConflict      = 10005 // 资源冲突
	CodeInternalError = 10006 // 服务器内部错误
	CodeValidation    = 10007 // 数据验证失败
	CodeTooManyReqs   = 10008 // 请求过于频繁

	// 认证相关错误码 (1xxxx)
	CodeInvalidToken      = 11001 // 无效的令牌
	CodeTokenExpired      = 11002 // 令牌已过期
	CodeInvalidPassword   = 11003 // 密码错误
	CodeInvalidCredential = 11004 // 无效的凭证
	CodeTokenMalformed    = 11005 // 令牌格式错误
	CodeTokenNotFound     = 11006 // 令牌不存在

	// 用户相关错误码 (2xxxx)
	CodeUserNotFound      = 20001 // 用户不存在
	CodeUserAlreadyExists = 20002 // 用户已存在
	CodeUserDisabled      = 20003 // 用户已禁用
	CodeEmailAlreadyUsed  = 20004 // 邮箱已被使用
	CodeUsernameExists    = 20005 // 用户名已存在
	CodePasswordTooWeak   = 20006 // 密码强度不足

	// 数据验证错误码 (3xxxx)
	CodeInvalidEmail    = 30001 // 无效的邮箱格式
	CodeInvalidUsername = 30002 // 无效的用户名格式
	CodeInvalidPhone    = 30003 // 无效的手机号格式
	CodeFieldRequired   = 30004 // 必填字段缺失
	CodeFieldTooLong    = 30005 // 字段长度超限
	CodeFieldTooShort   = 30006 // 字段长度不足

	// 资源相关错误码 (4xxxx)
	CodeResourceNotFound = 40001 // 资源不存在
	CodeResourceExists   = 40002 // 资源已存在
	CodeResourceLocked   = 40003 // 资源已锁定

	// 数据库相关错误码 (5xxxx)
	CodeDatabaseError   = 50001 // 数据库错误
	CodeDatabaseTimeout = 50002 // 数据库超时
	CodeDuplicateEntry  = 50003 // 重复条目
)

// AppError 应用程序错误类型
// 实现了 error 接口，可以像普通错误一样使用
type AppError struct {
	// Code 业务错误码
	Code int `json:"code"`
	// HTTPStatus HTTP 状态码
	HTTPStatus int `json:"-"`
	// Message 错误消息（面向用户）
	Message string `json:"message"`
	// Detail 错误详情（面向开发者，可选）
	Detail string `json:"detail,omitempty"`
	// Err 原始错误（用于错误链）
	Err error `json:"-"`
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现错误解包，支持 errors.Is 和 errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail 添加错误详情
func (e *AppError) WithDetail(detail string) *AppError {
	newErr := *e
	newErr.Detail = detail
	return &newErr
}

// WithError 包装原始错误
func (e *AppError) WithError(err error) *AppError {
	newErr := *e
	newErr.Err = err
	return &newErr
}

// New 创建新的应用错误
func New(code int, httpStatus int, message string) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: httpStatus,
		Message:    message,
	}
}

// Wrap 包装已有错误为应用错误
func Wrap(err error, code int, message string) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: http.StatusInternalServerError,
		Message:    message,
		Err:        err,
	}
}

// WrapWithStatus 包装已有错误并指定 HTTP 状态码
func WrapWithStatus(err error, code int, httpStatus int, message string) *AppError {
	return &AppError{
		Code:       code,
		HTTPStatus: httpStatus,
		Message:    message,
		Err:        err,
	}
}

// IsAppError 检查错误是否为 AppError 类型
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 将错误转换为 AppError 类型
// 如果错误不是 AppError 类型，返回 nil
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// FromError 从普通错误创建 AppError
// 如果已经是 AppError，直接返回
// 否则包装为内部服务器错误
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return ErrInternalServer.WithError(err)
}

// ============================================================
// 预定义错误实例
// ============================================================

// 通用错误
var (
	// ErrBadRequest 请求参数错误
	ErrBadRequest = &AppError{
		Code:       CodeBadRequest,
		HTTPStatus: http.StatusBadRequest,
		Message:    "请求参数错误",
	}

	// ErrUnauthorized 未授权
	ErrUnauthorized = &AppError{
		Code:       CodeUnauthorized,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "未授权，请先登录",
	}

	// ErrForbidden 禁止访问
	ErrForbidden = &AppError{
		Code:       CodeForbidden,
		HTTPStatus: http.StatusForbidden,
		Message:    "无权限访问该资源",
	}

	// ErrNotFound 资源不存在
	ErrNotFound = &AppError{
		Code:       CodeNotFound,
		HTTPStatus: http.StatusNotFound,
		Message:    "请求的资源不存在",
	}

	// ErrConflict 资源冲突
	ErrConflict = &AppError{
		Code:       CodeConflict,
		HTTPStatus: http.StatusConflict,
		Message:    "资源冲突",
	}

	// ErrInternalServer 服务器内部错误
	ErrInternalServer = &AppError{
		Code:       CodeInternalError,
		HTTPStatus: http.StatusInternalServerError,
		Message:    "服务器内部错误",
	}

	// ErrValidation 数据验证失败
	ErrValidation = &AppError{
		Code:       CodeValidation,
		HTTPStatus: http.StatusBadRequest,
		Message:    "数据验证失败",
	}

	// ErrTooManyRequests 请求过于频繁
	ErrTooManyRequests = &AppError{
		Code:       CodeTooManyReqs,
		HTTPStatus: http.StatusTooManyRequests,
		Message:    "请求过于频繁，请稍后再试",
	}
)

// 认证相关错误
var (
	// ErrInvalidToken 无效的令牌
	ErrInvalidToken = &AppError{
		Code:       CodeInvalidToken,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "无效的访问令牌",
	}

	// ErrTokenExpired 令牌已过期
	ErrTokenExpired = &AppError{
		Code:       CodeTokenExpired,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "访问令牌已过期",
	}

	// ErrInvalidPassword 密码错误
	ErrInvalidPassword = &AppError{
		Code:       CodeInvalidPassword,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "密码错误",
	}

	// ErrInvalidCredential 无效的凭证
	ErrInvalidCredential = &AppError{
		Code:       CodeInvalidCredential,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "用户名或密码错误",
	}

	// ErrTokenMalformed 令牌格式错误
	ErrTokenMalformed = &AppError{
		Code:       CodeTokenMalformed,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "令牌格式错误",
	}

	// ErrTokenNotFound 令牌不存在
	ErrTokenNotFound = &AppError{
		Code:       CodeTokenNotFound,
		HTTPStatus: http.StatusUnauthorized,
		Message:    "请提供访问令牌",
	}
)

// 用户相关错误
var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = &AppError{
		Code:       CodeUserNotFound,
		HTTPStatus: http.StatusNotFound,
		Message:    "用户不存在",
	}

	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = &AppError{
		Code:       CodeUserAlreadyExists,
		HTTPStatus: http.StatusConflict,
		Message:    "用户已存在",
	}

	// ErrUserDisabled 用户已禁用
	ErrUserDisabled = &AppError{
		Code:       CodeUserDisabled,
		HTTPStatus: http.StatusForbidden,
		Message:    "用户已被禁用",
	}

	// ErrEmailAlreadyUsed 邮箱已被使用
	ErrEmailAlreadyUsed = &AppError{
		Code:       CodeEmailAlreadyUsed,
		HTTPStatus: http.StatusConflict,
		Message:    "该邮箱已被注册",
	}

	// ErrUsernameExists 用户名已存在
	ErrUsernameExists = &AppError{
		Code:       CodeUsernameExists,
		HTTPStatus: http.StatusConflict,
		Message:    "该用户名已被使用",
	}

	// ErrPasswordTooWeak 密码强度不足
	ErrPasswordTooWeak = &AppError{
		Code:       CodePasswordTooWeak,
		HTTPStatus: http.StatusBadRequest,
		Message:    "密码强度不足，请使用更复杂的密码",
	}
)

// 数据验证相关错误
var (
	// ErrInvalidEmail 无效的邮箱格式
	ErrInvalidEmail = &AppError{
		Code:       CodeInvalidEmail,
		HTTPStatus: http.StatusBadRequest,
		Message:    "无效的邮箱格式",
	}

	// ErrInvalidUsername 无效的用户名格式
	ErrInvalidUsername = &AppError{
		Code:       CodeInvalidUsername,
		HTTPStatus: http.StatusBadRequest,
		Message:    "无效的用户名格式",
	}

	// ErrInvalidPhone 无效的手机号格式
	ErrInvalidPhone = &AppError{
		Code:       CodeInvalidPhone,
		HTTPStatus: http.StatusBadRequest,
		Message:    "无效的手机号格式",
	}

	// ErrFieldRequired 必填字段缺失
	ErrFieldRequired = &AppError{
		Code:       CodeFieldRequired,
		HTTPStatus: http.StatusBadRequest,
		Message:    "必填字段缺失",
	}
)

// 资源相关错误
var (
	// ErrResourceNotFound 资源不存在
	ErrResourceNotFound = &AppError{
		Code:       CodeResourceNotFound,
		HTTPStatus: http.StatusNotFound,
		Message:    "请求的资源不存在",
	}
)

// 数据库相关错误
var (
	// ErrDatabaseError 数据库错误
	ErrDatabaseError = &AppError{
		Code:       CodeDatabaseError,
		HTTPStatus: http.StatusInternalServerError,
		Message:    "数据库操作失败",
	}

	// ErrDuplicateEntry 重复条目
	ErrDuplicateEntry = &AppError{
		Code:       CodeDuplicateEntry,
		HTTPStatus: http.StatusConflict,
		Message:    "数据已存在",
	}
)
