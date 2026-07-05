package repository

import (
	"context"

	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl-metrics/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MetricsRollupRepositoryInterface interface {
	UpsertBatch(rows []model.MetricsRollupDaily) error
	AggregateByDay(appID uint, fromDay, toDay, name string) ([]model.MetricsRollupDaily, error)
	SumTotals(appID uint, fromDay, toDay string) ([]RollupNameTotal, error)
	TopN(appID uint, fromDay, toDay, name, dimKey string, limit int) ([]model.MetricsRollupDaily, error)
	contract.WithContext[MetricsRollupRepositoryInterface]
}

type RollupNameTotal struct {
	Name    string
	Count   int64
	Uniques int64
}

type MetricsRollupRepository struct {
	db *gorm.DB
}

var _ MetricsRollupRepositoryInterface = (*MetricsRollupRepository)(nil)

// NewMetricsRollupRepository 创建 MetricsRollupRepository 实例。
func NewMetricsRollupRepository(d *gorm.DB) MetricsRollupRepositoryInterface {
	return &MetricsRollupRepository{db: d}
}

// WithContext 设置上下文。
func (r *MetricsRollupRepository) WithContext(ctx context.Context) MetricsRollupRepositoryInterface {
	return &MetricsRollupRepository{db: r.db.WithContext(ctx)}
}

// UpsertBatch 批量 upsert 汇总行。
func (r *MetricsRollupRepository) UpsertBatch(rows []model.MetricsRollupDaily) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "app_id"},
			{Name: "day"},
			{Name: "name"},
			{Name: "dim_key"},
			{Name: "dim_value"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"count", "uniques", "updated_at"}),
	}).CreateInBatches(rows, 200).Error
}

// AggregateByDay 读取 rollup 按天序列。
func (r *MetricsRollupRepository) AggregateByDay(appID uint, fromDay, toDay, name string) ([]model.MetricsRollupDaily, error) {
	var rows []model.MetricsRollupDaily
	q := r.db.Where("app_id = ? AND day >= ? AND day <= ? AND dim_key = ?", appID, fromDay, toDay, model.RollupDimTotal)
	if name != "" {
		q = q.Where("name = ?", name)
	}
	err := q.Order("day asc").Find(&rows).Error
	return rows, err
}

// SumTotals 汇总区间总计。
func (r *MetricsRollupRepository) SumTotals(appID uint, fromDay, toDay string) ([]RollupNameTotal, error) {
	var rows []RollupNameTotal
	err := r.db.Model(&model.MetricsRollupDaily{}).
		Select("name, SUM(count) as count, SUM(uniques) as uniques").
		Where("app_id = ? AND day >= ? AND day <= ? AND dim_key = ?", appID, fromDay, toDay, model.RollupDimTotal).
		Group("name").
		Order("count desc").
		Scan(&rows).Error
	return rows, err
}

// TopN 读取维度 TopN。
func (r *MetricsRollupRepository) TopN(appID uint, fromDay, toDay, name, dimKey string, limit int) ([]model.MetricsRollupDaily, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []model.MetricsRollupDaily
	err := r.db.Where("app_id = ? AND day >= ? AND day <= ? AND name = ? AND dim_key = ?",
		appID, fromDay, toDay, name, dimKey).
		Order("count desc").
		Limit(limit * 32).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	// 合并跨天同 dim_value
	merged := map[string]*model.MetricsRollupDaily{}
	for i := range rows {
		key := rows[i].DimValue
		if cur, ok := merged[key]; ok {
			cur.Count += rows[i].Count
			cur.Uniques += rows[i].Uniques
		} else {
			copyRow := rows[i]
			merged[key] = &copyRow
		}
	}
	out := make([]model.MetricsRollupDaily, 0, len(merged))
	for _, v := range merged {
		out = append(out, *v)
	}
	// sort desc
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Count > out[i].Count {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
