package v1

import (
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl-metrics/app/service"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*QueryHandle)(nil)

type QueryHandle struct {
	svc *service.QueryService
}

// NewQueryHandle 创建 QueryHandle 实例。
func NewQueryHandle(svc *service.QueryService) *QueryHandle {
	return &QueryHandle{svc: svc}
}

// ModuleName 返回模块名称。
func (h *QueryHandle) ModuleName() (string, string) {
	return "metrics-query", "指标查询"
}

// Overview GET /api/v1/metrics/admin/overview
func (h *QueryHandle) Overview(ctx *gin.Context) {
	var req service.QueryRangeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	data, err := h.svc.Overview(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, data)
}

// Timeseries GET /api/v1/metrics/admin/timeseries
func (h *QueryHandle) Timeseries(ctx *gin.Context) {
	var req service.QueryRangeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	data, err := h.svc.Timeseries(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, data)
}

// TopN GET /api/v1/metrics/admin/topn
func (h *QueryHandle) TopN(ctx *gin.Context) {
	var req service.QueryRangeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	data, err := h.svc.TopN(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, data)
}

// Events GET /api/v1/metrics/admin/events
func (h *QueryHandle) Events(ctx *gin.Context) {
	var req service.QueryRangeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	var pageReq router.PageReq
	if err := ctx.ShouldBindQuery(&pageReq); err != nil {
		router.BadRequest(ctx, "参数校验失败")
		return
	}
	count, list, err := h.svc.ListEvents(ctx.Request.Context(), &req, &pageReq)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), pageReq.Page, pageReq.PageSize, list)
}
