package v1

import (
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl-metrics/app/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*AppHandle)(nil)

type AppHandle struct {
	svc *service.MetricsAppService
}

// NewAppHandle 创建 AppHandle 实例。
func NewAppHandle(svc *service.MetricsAppService) *AppHandle {
	return &AppHandle{svc: svc}
}

// ModuleName 返回模块名称。
func (h *AppHandle) ModuleName() (string, string) {
	return "metrics-app", "指标应用"
}

// Retrieve GET /api/v1/metrics/admin/apps
func (h *AppHandle) Retrieve(ctx *gin.Context) {
	var req router.PageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数校验失败")
		return
	}
	count, list, err := h.svc.Retrieve(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// Create POST /api/v1/metrics/admin/apps
func (h *AppHandle) Create(ctx *gin.Context) {
	var req service.CreateMetricsAppReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	item, err := h.svc.Create(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, item)
}

// Detail GET /api/v1/metrics/admin/apps/:id
func (h *AppHandle) Detail(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	item, err := h.svc.Detail(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, item)
}

// Update PUT /api/v1/metrics/admin/apps/:id
func (h *AppHandle) Update(ctx *gin.Context) {
	var req service.UpdateMetricsAppReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := h.svc.Update(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// Delete DELETE /api/v1/metrics/admin/apps/:id
func (h *AppHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if err := h.svc.Delete(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// RegenerateSecret PUT /api/v1/metrics/admin/apps/:id/regenerate-secret
func (h *AppHandle) RegenerateSecret(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	secret, err := h.svc.RegenerateSecret(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, gin.H{"appSecret": secret})
}
