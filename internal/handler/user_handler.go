// Package handler 提供 HTTP 请求处理器
//
// 本包实现了 HTTP 请求的处理逻辑，是 MVC 架构中的控制器层。
// Handler 负责：
// - 解析和验证请求参数
// - 调用 Service 层处理业务逻辑
// - 格式化并返回响应
//
// Handler 不应该包含业务逻辑，所有业务逻辑应该在 Service 层实现。
package handler

import (
	"net/http"

	"github.com/example/go-user-api/internal/middleware"
	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/service"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
// 处理所有用户相关的 HTTP 请求
type UserHandler struct {
	userService service.UserService
	log         logger.Logger
}

// NewUserHandler 创建用户处理器实例
// 参数：
//   - userService: 用户服务实例
//   - log: 日志记录器
func NewUserHandler(userService service.UserService, log logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log.With(logger.String("handler", "user")),
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "注册信息"
// @Success 201 {object} response.Response{data=model.UserResponse} "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "用户名或邮箱已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("注册参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层注册用户
	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Created(c, user.ToResponse())
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名/邮箱和密码登录，获取访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=model.LoginResponse} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 403 {object} response.Response "用户已被禁用"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("登录参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 获取客户端 IP
	clientIP := c.ClientIP()

	// 调用服务层登录
	resp, err := h.userService.Login(c.Request.Context(), &req, clientIP)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, resp)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body model.RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} response.Response{data=model.RefreshTokenResponse} "刷新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "令牌无效或已过期"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req model.RefreshTokenRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("刷新令牌参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层刷新令牌
	resp, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, resp)
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前用户
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=model.UserResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户 ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "")
		return
	}

	// 调用服务层获取用户
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, user.ToResponse())
}

// GetUser 获取用户信息
// @Summary 获取用户详情
// @Description 根据用户 ID 获取用户详细信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户 ID"
// @Success 200 {object} response.Response{data=model.UserResponse} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// 获取用户 ID 参数
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户 ID 不能为空")
		return
	}

	// 调用服务层获取用户
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, user.ToResponse())
}

// UpdateCurrentUser 更新当前用户信息
// @Summary 更新当前用户
// @Description 更新当前登录用户的信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.UserResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	// 从上下文获取用户 ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "")
		return
	}

	var req model.UpdateUserRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("更新用户参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层更新用户
	user, err := h.userService.Update(c.Request.Context(), userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, user.ToResponse())
}

// UpdateUser 更新用户信息（管理员）
// @Summary 更新用户（管理员）
// @Description 管理员更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户 ID"
// @Param request body model.AdminUpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.UserResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// 获取用户 ID 参数
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户 ID 不能为空")
		return
	}

	var req model.UpdateUserRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("更新用户参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层更新用户
	user, err := h.userService.Update(c.Request.Context(), userID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, user.ToResponse())
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response{data=model.MessageResponse} "修改成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "原密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 从上下文获取用户 ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "")
		return
	}

	var req model.ChangePasswordRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("修改密码参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层修改密码
	if err := h.userService.UpdatePassword(c.Request.Context(), userID, &req); err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, model.MessageResponse{Message: "密码修改成功"})
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户（软删除）
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id path string true "用户 ID"
// @Success 204 "删除成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 获取用户 ID 参数
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "用户 ID 不能为空")
		return
	}

	// 不允许删除自己
	currentUserID := middleware.GetUserID(c)
	if userID == currentUserID {
		response.BadRequest(c, "不能删除自己的账号")
		return
	}

	// 调用服务层删除用户
	if err := h.userService.Delete(c.Request.Context(), userID); err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.NoContent(c)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持搜索和过滤
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param username query string false "用户名（模糊搜索）"
// @Param email query string false "邮箱（模糊搜索）"
// @Param status query int false "状态：0-禁用，1-正常，2-未激活"
// @Param role query string false "角色：user, admin"
// @Param sort_by query string false "排序字段：created_at, updated_at, username, email"
// @Param sort_order query string false "排序方向：asc, desc"
// @Success 200 {object} response.Response{data=response.PageData} "获取成功"
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "无权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req model.UserListRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Debug("用户列表参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层获取用户列表
	users, total, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应格式
	userResponses := model.UsersToResponse(users)

	// 返回分页响应
	response.SuccessWithPagination(c, userResponses, req.GetDefaultPage(), req.GetDefaultPageSize(20, 100), total)
}

// handleError 处理错误响应
// 根据错误类型返回相应的 HTTP 响应
func (h *UserHandler) handleError(c *gin.Context, err error) {
	// 检查是否是应用错误
	if appErr := errors.AsAppError(err); appErr != nil {
		response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	// 未知错误，记录日志并返回内部错误
	h.log.Error("处理请求时发生未知错误", logger.Err(err))
	response.InternalError(c, "")
}

// handleValidationError 处理验证错误
// 解析验证错误并返回格式化的错误响应
func (h *UserHandler) handleValidationError(c *gin.Context, err error) {
	// 这里可以根据需要解析 validator 的错误，提取字段级别的错误信息
	// 为简化示例，这里直接返回通用的错误消息
	response.BadRequest(c, "请求参数验证失败: "+err.Error())
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务是否正常运行
// @Tags 系统
// @Produce json
// @Success 200 {object} response.Response{data=model.HealthResponse} "服务正常"
// @Router /health [get]
func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:  "healthy",
		Version: "v1.0.0",
	})
}
