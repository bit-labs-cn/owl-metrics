package service

import (
	"context"
	"time"

	"bit-labs.cn/owl-metrics/app/model"
	"bit-labs.cn/owl-metrics/app/repository"
)

type RollupService struct {
	eventRepo  repository.MetricsEventRepositoryInterface
	rollupRepo repository.MetricsRollupRepositoryInterface
}

// NewRollupService 创建 RollupService 实例。
func NewRollupService(
	eventRepo repository.MetricsEventRepositoryInterface,
	rollupRepo repository.MetricsRollupRepositoryInterface,
) *RollupService {
	return &RollupService{eventRepo: eventRepo, rollupRepo: rollupRepo}
}

// RollupDay 聚合指定日期的事件到 rollup 表。
func (s *RollupService) RollupDay(ctx context.Context, appID uint, day string) error {
	eventRepo := s.eventRepo.WithContext(ctx)
	rollupRepo := s.rollupRepo.WithContext(ctx)

	byDay, err := eventRepo.AggregateByDay(appID, day, day)
	if err != nil {
		return err
	}
	uniques, err := eventRepo.AggregateUniquesByDay(appID, day, day, "")
	if err != nil {
		return err
	}
	uvMap := map[string]int64{}
	for _, u := range uniques {
		uvMap[u.Day+"|"+u.Name] = u.Uniques
	}

	rows := make([]model.MetricsRollupDaily, 0, len(byDay)*4)
	for _, agg := range byDay {
		key := agg.Day + "|" + agg.Name
		rows = append(rows, model.MetricsRollupDaily{
			AppID:    appID,
			Day:      agg.Day,
			Name:     agg.Name,
			DimKey:   model.RollupDimTotal,
			DimValue: "_",
			Count:    agg.Count,
			Uniques:  uvMap[key],
		})
	}

	for _, dimKey := range []string{model.RollupDimPlatform, model.RollupDimOS, model.RollupDimURL} {
		for _, name := range distinctEventNames(byDay) {
			top, err := eventRepo.AggregateTopN(appID, day, day, name, dimKey, 100)
			if err != nil {
				return err
			}
			for _, t := range top {
				if t.DimValue == "" {
					continue
				}
				rows = append(rows, model.MetricsRollupDaily{
					AppID:    appID,
					Day:      day,
					Name:     name,
					DimKey:   dimKey,
					DimValue: t.DimValue,
					Count:    t.Count,
				})
			}
		}
	}
	return rollupRepo.UpsertBatch(rows)
}

// RollupYesterdayAllApps 聚合所有应用昨日数据。
func (s *RollupService) RollupYesterdayAllApps(ctx context.Context, appRepo repository.MetricsAppRepositoryInterface) error {
	day := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	_, apps, err := appRepo.WithContext(ctx).Retrieve(1, 1000, nil)
	if err != nil {
		return err
	}
	for i := range apps {
		if err := s.RollupDay(ctx, apps[i].ID, day); err != nil {
			return err
		}
	}
	return nil
}

func distinctEventNames(aggs []repository.DayEventAgg) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, a := range aggs {
		if _, ok := seen[a.Name]; ok {
			continue
		}
		seen[a.Name] = struct{}{}
		out = append(out, a.Name)
	}
	return out
}
