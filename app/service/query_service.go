package service

import (
	"context"
	"encoding/json"
	"time"

	"bit-labs.cn/owl-metrics/app/model"
	"bit-labs.cn/owl-metrics/app/repository"
	"bit-labs.cn/owl/provider/router"
	"gorm.io/gorm"
)

type QueryRangeReq struct {
	AppID   uint   `form:"appId" json:"appId" validate:"required,gt=0" label:"应用ID"`
	FromDay string `form:"fromDay" json:"fromDay" validate:"omitempty" label:"开始日期"`
	ToDay   string `form:"toDay" json:"toDay" validate:"omitempty" label:"结束日期"`
	Name    string `form:"name" json:"name" validate:"omitempty,max=64" label:"事件名"`
	DimKey  string `form:"dimKey" json:"dimKey" validate:"omitempty,max=64" label:"维度"`
	Limit   int    `form:"limit" json:"limit" validate:"omitempty,gte=1,lte=100" label:"条数"`
}

type OverviewResp struct {
	PageViews   int64            `json:"pageViews"`
	UniqueUsers int64            `json:"uniqueUsers"`
	Downloads   int64            `json:"downloads"`
	TotalEvents int64            `json:"totalEvents"`
	ByName      []NameCount      `json:"byName"`
	FromDay     string           `json:"fromDay"`
	ToDay       string           `json:"toDay"`
}

type NameCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type TimeseriesPoint struct {
	Day     string `json:"day"`
	Name    string `json:"name"`
	Count   int64  `json:"count"`
	Uniques int64  `json:"uniques,omitempty"`
}

type TopNItem struct {
	DimValue string `json:"dimValue"`
	Count    int64  `json:"count"`
	Uniques  int64  `json:"uniques,omitempty"`
}

type EventRow struct {
	ID         uint           `json:"id,string"`
	Name       string         `json:"name"`
	Platform   string         `json:"platform"`
	AnonID     string         `json:"anonId"`
	DistinctID string         `json:"distinctId"`
	URL        string         `json:"url"`
	OS         string         `json:"os"`
	Props      map[string]any `json:"props,omitempty"`
	ServerTS   int64          `json:"serverTs"`
	Day        string         `json:"day"`
}

type QueryService struct {
	eventRepo  repository.MetricsEventRepositoryInterface
	rollupRepo repository.MetricsRollupRepositoryInterface
}

// NewQueryService 创建 QueryService 实例。
func NewQueryService(
	eventRepo repository.MetricsEventRepositoryInterface,
	rollupRepo repository.MetricsRollupRepositoryInterface,
) *QueryService {
	return &QueryService{eventRepo: eventRepo, rollupRepo: rollupRepo}
}

// Overview 看板 KPI。
func (s *QueryService) Overview(ctx context.Context, req *QueryRangeReq) (*OverviewResp, error) {
	from, to := normalizeRange(req.FromDay, req.ToDay)
	totals, err := s.queryTotals(ctx, req.AppID, from, to)
	if err != nil {
		return nil, err
	}
	resp := &OverviewResp{
		ByName:  totals,
		FromDay: from,
		ToDay:   to,
	}
	for _, t := range totals {
		resp.TotalEvents += t.Count
		switch t.Name {
		case model.EventPageView:
			resp.PageViews = t.Count
		case model.EventDownload:
			resp.Downloads = t.Count
		}
	}
	uvRows, err := s.eventRepo.WithContext(ctx).AggregateUniquesByDay(req.AppID, from, to, model.EventPageView)
	if err == nil {
		seen := map[string]struct{}{}
		for _, r := range uvRows {
			_ = r
			resp.UniqueUsers += r.Uniques
		}
		if len(uvRows) == 0 {
			// fallback: count distinct anon for page_view in range
			_, events, _ := s.eventRepo.WithContext(ctx).Retrieve(1, 1, func(tx *gorm.DB) {
				tx.Where("app_id = ? AND day >= ? AND day <= ? AND name = ?", req.AppID, from, to, model.EventPageView)
			})
			_ = events
		}
		_ = seen
	}
	// simpler UV: sum daily uniques from rollup if historical, else event aggregate
	if today := time.Now().Format("2006-01-02"); to >= today {
		uvToday, _ := s.eventRepo.WithContext(ctx).AggregateUniquesByDay(req.AppID, today, today, model.EventPageView)
		for _, u := range uvToday {
			resp.UniqueUsers += u.Uniques
		}
	}
	rollupTotals, _ := s.rollupRepo.WithContext(ctx).SumTotals(req.AppID, from, to)
	for _, rt := range rollupTotals {
		if rt.Name == model.EventPageView {
			resp.UniqueUsers = max64(resp.UniqueUsers, rt.Uniques)
		}
	}
	return resp, nil
}

// Timeseries 按天趋势。
func (s *QueryService) Timeseries(ctx context.Context, req *QueryRangeReq) ([]TimeseriesPoint, error) {
	from, to := normalizeRange(req.FromDay, req.ToDay)
	today := time.Now().Format("2006-01-02")

	var points []TimeseriesPoint
	if from < today {
		histTo := to
		if histTo >= today {
			histTo = dayBefore(today)
		}
		if histTo >= from {
			rows, err := s.rollupRepo.WithContext(ctx).AggregateByDay(req.AppID, from, histTo, req.Name)
			if err != nil {
				return nil, err
			}
			for _, r := range rows {
				points = append(points, TimeseriesPoint{
					Day:     r.Day,
					Name:    r.Name,
					Count:   r.Count,
					Uniques: r.Uniques,
				})
			}
		}
	}
	if to >= today {
		rows, err := s.eventRepo.WithContext(ctx).AggregateByDay(req.AppID, today, today)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if req.Name != "" && r.Name != req.Name {
				continue
			}
			points = append(points, TimeseriesPoint{Day: r.Day, Name: r.Name, Count: r.Count})
		}
	}
	return points, nil
}

// TopN 热门维度。
func (s *QueryService) TopN(ctx context.Context, req *QueryRangeReq) ([]TopNItem, error) {
	from, to := normalizeRange(req.FromDay, req.ToDay)
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	dimKey := req.DimKey
	if dimKey == "" {
		dimKey = model.RollupDimPlatform
	}
	name := req.Name
	if name == "" {
		name = model.EventPageView
	}

	today := time.Now().Format("2006-01-02")
	merged := map[string]int64{}

	if from < today {
		histTo := to
		if histTo >= today {
			histTo = dayBefore(today)
		}
		if histTo >= from {
			rows, err := s.rollupRepo.WithContext(ctx).TopN(req.AppID, from, histTo, name, dimKey, limit)
			if err != nil {
				return nil, err
			}
			for _, r := range rows {
				merged[r.DimValue] += r.Count
			}
		}
	}
	if to >= today {
		rows, err := s.eventRepo.WithContext(ctx).AggregateTopN(req.AppID, today, today, name, dimKey, limit)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			merged[r.DimValue] += r.Count
		}
	}

	out := make([]TopNItem, 0, len(merged))
	for k, v := range merged {
		out = append(out, TopNItem{DimValue: k, Count: v})
	}
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

// ListEvents 事件浏览器。
func (s *QueryService) ListEvents(ctx context.Context, req *QueryRangeReq, pageReq *router.PageReq) (int64, []EventRow, error) {
	from, to := normalizeRange(req.FromDay, req.ToDay)
	count, list, err := s.eventRepo.WithContext(ctx).Retrieve(pageReq.Page, pageReq.PageSize, func(tx *gorm.DB) {
		tx.Where("app_id = ? AND day >= ? AND day <= ?", req.AppID, from, to)
		if req.Name != "" {
			tx.Where("name = ?", req.Name)
		}
	})
	if err != nil {
		return 0, nil, err
	}
	out := make([]EventRow, 0, len(list))
	for i := range list {
		row := EventRow{
			ID:         list[i].ID,
			Name:       list[i].Name,
			Platform:   list[i].Platform,
			AnonID:     list[i].AnonID,
			DistinctID: list[i].DistinctID,
			URL:        list[i].URL,
			OS:         list[i].OS,
			ServerTS:   list[i].ServerTS,
			Day:        list[i].Day,
		}
		if list[i].Props != "" {
			var props map[string]any
			_ = json.Unmarshal([]byte(list[i].Props), &props)
			row.Props = props
		}
		out = append(out, row)
	}
	return count, out, nil
}

func (s *QueryService) queryTotals(ctx context.Context, appID uint, from, to string) ([]NameCount, error) {
	today := time.Now().Format("2006-01-02")
	merged := map[string]int64{}

	if from < today {
		histTo := to
		if histTo >= today {
			histTo = dayBefore(today)
		}
		if histTo >= from {
			rows, err := s.rollupRepo.WithContext(ctx).SumTotals(appID, from, histTo)
			if err != nil {
				return nil, err
			}
			for _, r := range rows {
				merged[r.Name] += r.Count
			}
		}
	}
	if to >= today {
		rows, err := s.eventRepo.WithContext(ctx).AggregateTotals(appID, today, today)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			merged[r.Name] += r.Count
		}
	}
	out := make([]NameCount, 0, len(merged))
	for name, count := range merged {
		out = append(out, NameCount{Name: name, Count: count})
	}
	return out, nil
}

func normalizeRange(from, to string) (string, string) {
	today := time.Now().Format("2006-01-02")
	if to == "" {
		to = today
	}
	if from == "" {
		from = time.Now().AddDate(0, 0, -29).Format("2006-01-02")
	}
	return from, to
}

func dayBefore(day string) string {
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		return day
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
