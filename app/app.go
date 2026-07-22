package app

import (
	"context"
	"log"

	"bit-labs.cn/owl-metrics/app/database"
	v1 "bit-labs.cn/owl-metrics/app/handle/v1"
	"bit-labs.cn/owl-metrics/app/repository"
	"bit-labs.cn/owl-metrics/app/route"
	"bit-labs.cn/owl-metrics/app/service"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/router"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var _ owl.SubApp = (*SubAppMetrics)(nil)

type SubAppMetrics struct {
	app foundation.Application
}

// Name 获取应用名称。
func (s *SubAppMetrics) Name() string { return "metrics" }

// Bootstrap 引导应用启动。
func (s *SubAppMetrics) Bootstrap() {}

func (s *SubAppMetrics) RegisterMigrate() []any { return database.Models() }

func (s *SubAppMetrics) BeforeMigrate(db *gorm.DB) error { return nil }

func (s *SubAppMetrics) AfterMigrate(db *gorm.DB) error {
	return s.app.Invoke(func(rollup *service.RollupService, appRepo repository.MetricsAppRepositoryInterface) {
		go func() {
			if err := rollup.RollupYesterdayAllApps(context.Background(), appRepo); err != nil {
				log.Printf("[metrics] startup rollup: %v", err)
			}
		}()
	})
}

// ServiceProviders 获取服务提供者列表。
func (s *SubAppMetrics) ServiceProviders() []foundation.ServiceProvider {
	return nil
}

// Menu 获取菜单配置。
func (s *SubAppMetrics) RegisterMenus() []*router.Menu {
	return route.InitMenu()
}

// Commands 获取命令列表。
func (s *SubAppMetrics) RegisterCommands() []*cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics-rollup",
		Short: "聚合指标事件到每日 rollup 表",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.app.Invoke(func(rollup *service.RollupService, appRepo repository.MetricsAppRepositoryInterface) error {
				return rollup.RollupYesterdayAllApps(cmd.Context(), appRepo)
			})
		},
	}
	return []*cobra.Command{cmd}
}

// RegisterRouters 注册路由。
func (s *SubAppMetrics) RegisterRouters() {
	route.InitApi(s.app, s.Name())
}

// Binds 获取绑定列表。
func (s *SubAppMetrics) Binds() []any {
	return []any{
		repository.NewMetricsAppRepository,
		repository.NewMetricsEventRepository,
		repository.NewMetricsRollupRepository,

		NewSQLMetricsStoreProvider,
		service.NewMetricsAppService,
		service.NewIngestService,
		service.NewQueryService,
		service.NewRollupService,

		v1.NewCollectHandle,
		v1.NewAppHandle,
		v1.NewQueryHandle,
	}
}

// NewSQLMetricsStoreProvider 提供 MetricsStore 实现。
func NewSQLMetricsStoreProvider(eventRepo repository.MetricsEventRepositoryInterface) service.MetricsStore {
	return service.NewSQLMetricsStore(eventRepo)
}
