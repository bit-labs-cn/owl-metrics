package service

import (
	"context"

	"bit-labs.cn/owl-metrics/app/model"
	"bit-labs.cn/owl-metrics/app/repository"
)

// MetricsStore 指标存储抽象（预留 ClickHouse 实现）。
type MetricsStore interface {
	Ingest(ctx context.Context, events []model.MetricsEvent) error
}

// SQLMetricsStore 基于 PostgreSQL/SQL 的 MetricsStore 实现。
type SQLMetricsStore struct {
	eventRepo repository.MetricsEventRepositoryInterface
}

var _ MetricsStore = (*SQLMetricsStore)(nil)

// NewSQLMetricsStore 创建 SQLMetricsStore。
func NewSQLMetricsStore(eventRepo repository.MetricsEventRepositoryInterface) *SQLMetricsStore {
	return &SQLMetricsStore{eventRepo: eventRepo}
}

// Ingest 批量写入事件明细。
func (s *SQLMetricsStore) Ingest(ctx context.Context, events []model.MetricsEvent) error {
	return s.eventRepo.WithContext(ctx).BatchCreate(events)
}
