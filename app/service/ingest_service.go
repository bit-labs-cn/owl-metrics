package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"bit-labs.cn/owl-metrics/app/model"
	"bit-labs.cn/owl-metrics/app/pkg/signing"
	"bit-labs.cn/owl-metrics/app/pkg/useragent"
)

type CollectEventReq struct {
	AppKey     string         `json:"appKey" validate:"required,max=64" label:"AppKey"`
	Name       string         `json:"name" validate:"required,max=64" label:"事件名"`
	DistinctID string         `json:"distinctId" validate:"omitempty,max=128" label:"用户标识"`
	AnonID     string         `json:"anonId" validate:"omitempty,max=128" label:"匿名ID"`
	SessionID  string         `json:"sessionId" validate:"omitempty,max=128" label:"会话ID"`
	Platform   string         `json:"platform" validate:"omitempty,max=32" label:"平台"`
	Props      map[string]any `json:"props" label:"属性"`
	URL        string         `json:"url" validate:"omitempty,max=1024" label:"URL"`
	Referrer   string         `json:"referrer" validate:"omitempty,max=1024" label:"来源"`
	UA         string         `json:"ua" validate:"omitempty,max=512" label:"UA"`
	OS         string         `json:"os" validate:"omitempty,max=32" label:"操作系统"`
	OSVersion  string         `json:"osVersion" validate:"omitempty,max=64" label:"系统版本"`
	AppVersion string         `json:"appVersion" validate:"omitempty,max=64" label:"应用版本"`
	ClientTS   int64          `json:"clientTs" label:"客户端时间戳"`
}

type CollectContext struct {
	IP        string
	UA        string
	Origin    string
	Timestamp string
	Signature string
	RawBody   []byte
	Signed    bool
}

type IngestService struct {
	store   MetricsStore
	appSvc  *MetricsAppService
}

// NewIngestService 创建 IngestService 实例。
func NewIngestService(store MetricsStore, appSvc *MetricsAppService) *IngestService {
	return &IngestService{store: store, appSvc: appSvc}
}

// Collect 采集单条或多条事件。
func (s *IngestService) Collect(ctx context.Context, events []CollectEventReq, cc CollectContext) error {
	if len(events) == 0 {
		return CollectInvalid("事件不能为空")
	}
	appKey := strings.TrimSpace(events[0].AppKey)
	app, err := s.appSvc.FindByAppKey(ctx, appKey)
	if err != nil {
		return err
	}

	if cc.Signed {
		if !signing.VerifyHMAC(app.AppSecret, cc.Timestamp, cc.Signature, string(cc.RawBody)) {
			return SignatureInvalid()
		}
	} else if cc.Origin != "" && !originAllowed(app.AllowedOrigins, cc.Origin) {
		return OriginForbidden()
	}

	now := time.Now()
	serverTS := now.UnixMilli()
	day := now.Format("2006-01-02")
	ua := firstNonEmpty(cc.UA, events[0].UA)
	parsedOS, parsedVer := useragent.Parse(ua)

	rows := make([]model.MetricsEvent, 0, len(events))
	for _, ev := range events {
		if strings.TrimSpace(ev.AppKey) != app.AppKey {
			return CollectInvalid("批量事件 AppKey 不一致")
		}
		name := strings.TrimSpace(ev.Name)
		if name == "" {
			continue
		}
		osName := firstNonEmpty(ev.OS, parsedOS)
		osVer := firstNonEmpty(ev.OSVersion, parsedVer)
		propsJSON := ""
		if len(ev.Props) > 0 {
			b, _ := json.Marshal(ev.Props)
			propsJSON = string(b)
		}
		platform := strings.TrimSpace(ev.Platform)
		if platform == "" {
			platform = model.PlatformWeb
		}
		rows = append(rows, model.MetricsEvent{
			AppID:      app.ID,
			AppKey:     app.AppKey,
			Name:       name,
			DistinctID: strings.TrimSpace(ev.DistinctID),
			AnonID:     strings.TrimSpace(ev.AnonID),
			SessionID:  strings.TrimSpace(ev.SessionID),
			Platform:   platform,
			Props:      propsJSON,
			URL:        truncate(ev.URL, 1024),
			Referrer:   truncate(ev.Referrer, 1024),
			UA:         truncate(ua, 512),
			OS:         osName,
			OSVersion:  osVer,
			AppVersion: strings.TrimSpace(ev.AppVersion),
			IP:         cc.IP,
			ClientTS:   ev.ClientTS,
			ServerTS:   serverTS,
			Day:        day,
		})
	}
	if len(rows) == 0 {
		return CollectInvalid("无有效事件")
	}
	return s.store.Ingest(ctx, rows)
}

func originAllowed(rawOrigins, origin string) bool {
	origin = strings.TrimSpace(strings.ToLower(origin))
	if origin == "" {
		return true
	}
	var origins []string
	_ = json.Unmarshal([]byte(rawOrigins), &origins)
	if len(origins) == 0 {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	for _, allowed := range origins {
		allowed = strings.TrimSpace(strings.ToLower(allowed))
		if allowed == "" {
			continue
		}
		if allowed == "*" || allowed == origin {
			return true
		}
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*.")
			if host == suffix || strings.HasSuffix(host, "."+suffix) {
				return true
			}
		}
		if au, err := url.Parse(allowed); err == nil && strings.EqualFold(au.Host, host) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n]
}
