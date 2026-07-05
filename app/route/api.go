package route

import (
	v1 "bit-labs.cn/owl-metrics/app/handle/v1"
	"time"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/router"
	middleware2 "bit-labs.cn/owl/provider/router/middleware"
	"github.com/gin-gonic/gin"
)

var (
	appMenu      *router.Menu
	dashboardMenu *router.Menu
)

// InitMenu 初始化菜单。
func InitMenu() []*router.Menu {
	return []*router.Menu{
		{
			Path: "/metrics",
			Name: "Metrics",
			Meta: router.Meta{
				Title:      "指标分析",
				Icon:       "ep:data-line",
				ShowParent: true,
				ShowLink:   true,
			},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{dashboardMenu, appMenu},
		},
	}
}

// InitApi 初始化 API 路由。
func InitApi(app foundation.Application, appName string) {
	err := app.Invoke(func(
		engine *gin.Engine,
		collect *v1.CollectHandle,
		appH *v1.AppHandle,
		query *v1.QueryHandle,
	) {
		gv1 := engine.Group("/api/v1")

		collectRoute := router.NewRouteInfoBuilder(appName, collect, gv1, router.MenuOption{})
		collectRoute.Use(middleware2.RateLimiter(time.Second*1, 20))
		collectRoute.Post("/metrics/collect", router.AccessPublic, collect.Collect).Name("指标采集").WithoutOperateLog().Build()
		collectRoute.Get("/metrics/pixel.gif", router.AccessPublic, collect.Pixel).Name("指标像素采集").WithoutOperateLog().Build()

		adminApp := router.NewRouteInfoBuilder(appName, appH, gv1, router.MenuOption{
			Path:          "/metrics/admin/apps/index",
			ComponentName: "MetricsApps",
			Icon:          "ep:collection",
		})
		adminApp.Get("/metrics/admin/apps", router.AccessAuthorized, appH.Retrieve).Name("指标应用列表").Build()
		adminApp.Post("/metrics/admin/apps", router.AccessAuthorized, appH.Create).Name("创建指标应用").Build()
		adminApp.Get("/metrics/admin/apps/:id", router.AccessAuthorized, appH.Detail).Name("指标应用详情").Build()
		adminApp.Put("/metrics/admin/apps/:id", router.AccessAuthorized, appH.Update).Name("更新指标应用").Build()
		adminApp.Delete("/metrics/admin/apps/:id", router.AccessAuthorized, appH.Delete).Name("删除指标应用").Build()
		adminApp.Put("/metrics/admin/apps/:id/regenerate-secret", router.AccessAuthorized, appH.RegenerateSecret).Name("重置AppSecret").Build()
		appMenu = adminApp.GetMenu()

		adminQuery := router.NewRouteInfoBuilder(appName, query, gv1, router.MenuOption{
			Path:          "/metrics/admin/dashboard/index",
			ComponentName: "MetricsDashboard",
			Icon:          "ep:trend-charts",
		})
		adminQuery.Get("/metrics/admin/overview", router.AccessAuthorized, query.Overview).Name("指标概览").Build()
		adminQuery.Get("/metrics/admin/timeseries", router.AccessAuthorized, query.Timeseries).Name("指标趋势").Build()
		adminQuery.Get("/metrics/admin/topn", router.AccessAuthorized, query.TopN).Name("指标TopN").Build()
		adminQuery.Get("/metrics/admin/events", router.AccessAuthorized, query.Events).Name("指标事件").Build()
		dashboardMenu = adminQuery.GetMenu()
	})
	owl.PanicIf(err)
}
