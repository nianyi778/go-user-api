// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"
	"time"

	"github.com/example/go-user-api/internal/model"
	"github.com/example/go-user-api/internal/service"
	"github.com/example/go-user-api/pkg/errors"
	"github.com/example/go-user-api/pkg/logger"
	"github.com/example/go-user-api/pkg/response"
	"github.com/gin-gonic/gin"
)

// RiskReportUsageHandler 风险报告使用记录处理器
// 处理所有使用记录相关的 HTTP 请求
type RiskReportUsageHandler struct {
	service service.RiskReportUsageService
	log     logger.Logger
}

// NewRiskReportUsageHandler 创建风险报告使用记录处理器实例
func NewRiskReportUsageHandler(service service.RiskReportUsageService, log logger.Logger) *RiskReportUsageHandler {
	return &RiskReportUsageHandler{
		service: service,
		log:     log.With(logger.String("handler", "risk_report_usage")),
	}
}

// Create 创建使用记录
// @Summary 上报使用记录
// @Description 上报单次查询的使用记录
// @Tags 风险报告
// @Accept json
// @Produce json
// @Param request body model.CreateRiskReportUsageRequest true "使用记录信息"
// @Success 201 {object} response.Response{data=model.RiskReportUsageResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/risk-report/usage [post]
func (h *RiskReportUsageHandler) Create(c *gin.Context) {
	var req model.CreateRiskReportUsageRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("使用记录参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层创建记录
	usage, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"success":   true,
		"message":   "记录已保存",
		"record_id": usage.ID,
	}
	response.Created(c, resp)
}

// BatchCreate 批量创建使用记录
// @Summary 批量上报使用记录
// @Description 批量上报多次查询的使用记录
// @Tags 风险报告
// @Accept json
// @Produce json
// @Param request body model.BatchCreateRiskReportUsageRequest true "批量使用记录信息"
// @Success 200 {object} response.Response{data=model.BatchCreateRiskReportUsageResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/risk-report/usage/batch [post]
func (h *RiskReportUsageHandler) BatchCreate(c *gin.Context) {
	var req model.BatchCreateRiskReportUsageRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Debug("批量使用记录参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层批量创建
	result, err := h.service.BatchCreate(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	resp := map[string]interface{}{
		"success":       true,
		"message":       "批量创建完成",
		"success_count": result.SuccessCount,
		"failure_count": result.FailureCount,
		"record_ids":    result.RecordIDs,
		"errors":        result.Errors,
	}
	response.Success(c, resp)
}

// GetByID 获取使用记录详情
// @Summary 获取使用记录详情
// @Description 根据 ID 获取使用记录详情
// @Tags 风险报告
// @Produce json
// @Param id path string true "记录 ID"
// @Success 200 {object} response.Response{data=model.RiskReportUsageResponse} "查询成功"
// @Failure 404 {object} response.Response "记录不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/risk-report/usage/{id} [get]
func (h *RiskReportUsageHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// 调用服务层获取记录
	usage, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, usage.ToResponse())
}

// List 获取使用记录列表
// @Summary 获取使用记录列表
// @Description 根据条件查询使用记录列表
// @Tags 风险报告
// @Produce json
// @Param user_id query string false "用户 ID"
// @Param ticker query string false "股票代码"
// @Param start_time query string false "开始时间（RFC3339 格式）"
// @Param end_time query string false "结束时间（RFC3339 格式）"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.PageData} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/risk-report/usage [get]
func (h *RiskReportUsageHandler) List(c *gin.Context) {
	var req model.RiskReportUsageListRequest

	// 绑定并验证请求参数
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Debug("使用记录列表参数验证失败", logger.Err(err))
		h.handleValidationError(c, err)
		return
	}

	// 调用服务层获取列表
	usages, total, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 转换为响应格式
	usageResponses := make([]interface{}, len(usages))
	for i, usage := range usages {
		usageResponses[i] = usage.ToResponse()
	}

	// 返回分页响应
	response.SuccessWithPagination(c, usageResponses, req.Page, req.PageSize, total)
}

// GetUserStats 获取用户统计信息
// @Summary 获取用户统计信息
// @Description 获取指定用户的使用统计信息
// @Tags 风险报告
// @Produce json
// @Param user_id path string true "用户 ID"
// @Param start_time query string false "开始时间（RFC3339 格式）"
// @Param end_time query string false "结束时间（RFC3339 格式）"
// @Success 200 {object} response.Response{data=map[string]interface{}} "查询成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/risk-report/usage/stats/{user_id} [get]
func (h *RiskReportUsageHandler) GetUserStats(c *gin.Context) {
	userID := c.Param("user_id")

	// 解析时间参数
	var startTime, endTime time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	// 调用服务层获取统计信息
	stats, err := h.service.GetUserStats(c.Request.Context(), userID, startTime, endTime)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 返回成功响应
	response.Success(c, stats)
}

// handleValidationError 处理验证错误
func (h *RiskReportUsageHandler) handleValidationError(c *gin.Context, err error) {
	response.BadRequest(c, "参数验证失败: "+err.Error())
}

// handleError 处理错误
func (h *RiskReportUsageHandler) handleError(c *gin.Context, err error) {
	// 如果是自定义错误，使用错误中的状态码
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.HTTPStatus, response.Response{
			Code:    appErr.Code,
			Message: appErr.Message,
			Data:    nil,
		})
		return
	}

	// 默认返回 500 错误
	response.Error(c, http.StatusInternalServerError, response.CodeInternalError, "服务器内部错误")
}
