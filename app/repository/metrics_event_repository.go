package repository

import (
	"context"

	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl-metrics/app/model"
	"gorm.io/gorm"
)

type MetricsEventRepositoryInterface interface {
	BatchCreate(events []model.MetricsEvent) error
	Retrieve(page, pageSize int, fn func(*gorm.DB)) (int64, []model.MetricsEvent, error)
	AggregateByDay(appID uint, fromDay, toDay string) ([]DayEventAgg, error)
	AggregateTopN(appID uint, fromDay, toDay, name, dimKey string, limit int) ([]DimAgg, error)
	AggregateTotals(appID uint, fromDay, toDay string) ([]NameAgg, error)
	AggregateUniquesByDay(appID uint, fromDay, toDay, name string) ([]DayUniqueAgg, error)
	contract.WithContext[MetricsEventRepositoryInterface]
}

type DayEventAgg struct {
	Day   string
	Name  string
	Count int64
}

type DayUniqueAgg struct {
	Day     string
	Name    string
	Uniques int64
}

type NameAgg struct {
	Name  string
	Count int64
}

type DimAgg struct {
	DimValue string
	Count    int64
}

type MetricsEventRepository struct {
	db *gorm.DB
}

var _ MetricsEventRepositoryInterface = (*MetricsEventRepository)(nil)

// NewMetricsEventRepository 创建 MetricsEventRepository 实例。
func NewMetricsEventRepository(d *gorm.DB) MetricsEventRepositoryInterface {
	return &MetricsEventRepository{db: d}
}

// WithContext 设置上下文。
func (r *MetricsEventRepository) WithContext(ctx context.Context) MetricsEventRepositoryInterface {
	return &MetricsEventRepository{db: r.db.WithContext(ctx)}
}

// BatchCreate 批量写入事件。
func (r *MetricsEventRepository) BatchCreate(events []model.MetricsEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.CreateInBatches(events, 200).Error
}

// Retrieve 分页查询明细。
func (r *MetricsEventRepository) Retrieve(page, pageSize int, fn func(*gorm.DB)) (int64, []model.MetricsEvent, error) {
	q := r.db.Model(&model.MetricsEvent{})
	if fn != nil {
		fn(q)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.MetricsEvent
	err := q.Order("server_ts desc, id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return total, list, err
}

// AggregateByDay 按天+事件名聚合计数。
func (r *MetricsEventRepository) AggregateByDay(appID uint, fromDay, toDay string) ([]DayEventAgg, error) {
	var rows []DayEventAgg
	err := r.db.Model(&model.MetricsEvent{}).
		Select("day, name, COUNT(*) as count").
		Where("app_id = ? AND day >= ? AND day <= ?", appID, fromDay, toDay).
		Group("day, name").
		Order("day asc").
		Scan(&rows).Error
	return rows, err
}

// AggregateUniquesByDay 按天+事件名聚合 UV（anon_id 去重）。
func (r *MetricsEventRepository) AggregateUniquesByDay(appID uint, fromDay, toDay, name string) ([]DayUniqueAgg, error) {
	var rows []DayUniqueAgg
	q := r.db.Model(&model.MetricsEvent{}).
		Select("day, name, COUNT(DISTINCT anon_id) as uniques").
		Where("app_id = ? AND day >= ? AND day <= ? AND anon_id <> ''", appID, fromDay, toDay)
	if name != "" {
		q = q.Where("name = ?", name)
	}
	err := q.Group("day, name").Order("day asc").Scan(&rows).Error
	return rows, err
}

// AggregateTotals 按事件名汇总计数。
func (r *MetricsEventRepository) AggregateTotals(appID uint, fromDay, toDay string) ([]NameAgg, error) {
	var rows []NameAgg
	err := r.db.Model(&model.MetricsEvent{}).
		Select("name, COUNT(*) as count").
		Where("app_id = ? AND day >= ? AND day <= ?", appID, fromDay, toDay).
		Group("name").
		Order("count desc").
		Scan(&rows).Error
	return rows, err
}

// AggregateTopN 按维度 TopN（platform/os/url 等从列或 props 解析，此处用列字段）。
func (r *MetricsEventRepository) AggregateTopN(appID uint, fromDay, toDay, name, dimKey string, limit int) ([]DimAgg, error) {
	col := "platform"
	switch dimKey {
	case "platform":
		col = "platform"
	case "os":
		col = "os"
	case "url":
		col = "url"
	default:
		col = "platform"
	}
	if limit <= 0 {
		limit = 10
	}
	var rows []DimAgg
	q := r.db.Model(&model.MetricsEvent{}).
		Select(col+" as dim_value, COUNT(*) as count").
		Where("app_id = ? AND day >= ? AND day <= ?", appID, fromDay, toDay)
	if name != "" {
		q = q.Where("name = ?", name)
	}
	err := q.Group(col).Order("count desc").Limit(limit).Scan(&rows).Error
	return rows, err
}
