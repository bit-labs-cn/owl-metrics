package repository

import (
	"context"

	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl-metrics/app/model"
	"gorm.io/gorm"
)

type MetricsAppRepositoryInterface interface {
	Save(m *model.MetricsApp) error
	Detail(id any) (*model.MetricsApp, error)
	Delete(id any) error
	Retrieve(page, pageSize int, fn func(*gorm.DB)) (int64, []model.MetricsApp, error)
	FindByAppKey(appKey string) (*model.MetricsApp, error)
	contract.WithContext[MetricsAppRepositoryInterface]
}

type MetricsAppRepository struct {
	db *gorm.DB
	db.BaseRepository[model.MetricsApp]
}

var _ MetricsAppRepositoryInterface = (*MetricsAppRepository)(nil)

// NewMetricsAppRepository 创建 MetricsAppRepository 实例。
func NewMetricsAppRepository(d *gorm.DB) MetricsAppRepositoryInterface {
	return &MetricsAppRepository{db: d, BaseRepository: db.NewBaseRepository[model.MetricsApp](d)}
}

// WithContext 设置上下文。
func (r *MetricsAppRepository) WithContext(ctx context.Context) MetricsAppRepositoryInterface {
	r.db = r.db.WithContext(ctx)
	r.BaseRepository = db.NewBaseRepository[model.MetricsApp](r.db)
	return r
}

// Save 保存。
func (r *MetricsAppRepository) Save(m *model.MetricsApp) error {
	return r.BaseRepository.Save(m)
}

// Detail 获取详情。
func (r *MetricsAppRepository) Detail(id any) (*model.MetricsApp, error) {
	return r.BaseRepository.Detail(id)
}

// Delete 删除。
func (r *MetricsAppRepository) Delete(id any) error {
	return r.BaseRepository.Delete(id)
}

// Retrieve 分页查询。
func (r *MetricsAppRepository) Retrieve(page, pageSize int, fn func(*gorm.DB)) (int64, []model.MetricsApp, error) {
	return r.BaseRepository.Retrieve(page, pageSize, fn)
}

// FindByAppKey 按 AppKey 查询。
func (r *MetricsAppRepository) FindByAppKey(appKey string) (*model.MetricsApp, error) {
	var m model.MetricsApp
	err := r.db.Where("app_key = ?", appKey).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
