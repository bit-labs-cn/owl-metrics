package v1

import (
	"encoding/json"
	"io"
	"strings"

	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl-metrics/app/service"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*CollectHandle)(nil)

type CollectHandle struct {
	ingest *service.IngestService
}

// NewCollectHandle 创建 CollectHandle 实例。
func NewCollectHandle(ingest *service.IngestService) *CollectHandle {
	return &CollectHandle{ingest: ingest}
}

// ModuleName 返回模块名称。
func (h *CollectHandle) ModuleName() (string, string) {
	return "metrics-collect", "指标采集"
}

type collectPayload struct {
	Events []service.CollectEventReq `json:"events"`
}

// Collect POST /api/v1/metrics/collect
func (h *CollectHandle) Collect(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		router.Fail(ctx, service.CollectInvalid("读取请求体失败"))
		return
	}

	cc := service.CollectContext{
		IP:        clientIP(ctx),
		UA:        ctx.GetHeader("User-Agent"),
		Origin:    firstHeader(ctx, "Origin", "Referer"),
		Timestamp: ctx.GetHeader("X-Metrics-Timestamp"),
		Signature: ctx.GetHeader("X-Metrics-Signature"),
		RawBody:   body,
		Signed:    strings.TrimSpace(ctx.GetHeader("X-Metrics-Signature")) != "",
	}

	events, err := parseCollectBody(body)
	if err != nil {
		router.Fail(ctx, service.CollectInvalid(err.Error()))
		return
	}
	if len(events) == 0 {
		router.Fail(ctx, service.CollectInvalid("事件不能为空"))
		return
	}
	if err := h.ingest.Collect(ctx.Request.Context(), events, cc); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// Pixel GET /api/v1/metrics/pixel.gif
func (h *CollectHandle) Pixel(ctx *gin.Context) {
	appKey := strings.TrimSpace(ctx.Query("appKey"))
	name := strings.TrimSpace(ctx.Query("name"))
	if appKey == "" || name == "" {
		ctx.Data(200, "image/gif", pixelGIF)
		return
	}
	ev := service.CollectEventReq{
		AppKey:   appKey,
		Name:     name,
		Platform: modelPlatform(ctx.Query("platform")),
		URL:      ctx.Query("url"),
		AnonID:   ctx.Query("anonId"),
	}
	_ = h.ingest.Collect(ctx.Request.Context(), []service.CollectEventReq{ev}, service.CollectContext{
		IP: clientIP(ctx),
		UA: ctx.GetHeader("User-Agent"),
	})
	ctx.Data(200, "image/gif", pixelGIF)
}

func parseCollectBody(body []byte) ([]service.CollectEventReq, error) {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return nil, nil
	}
	var single service.CollectEventReq
	if err := json.Unmarshal(body, &single); err == nil && single.AppKey != "" && single.Name != "" {
		return []service.CollectEventReq{single}, nil
	}
	var batch collectPayload
	if err := json.Unmarshal(body, &batch); err != nil {
		return nil, err
	}
	if len(batch.Events) > 0 {
		return batch.Events, nil
	}
	return nil, nil
}

func clientIP(ctx *gin.Context) string {
	if xff := strings.TrimSpace(ctx.GetHeader("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	return ctx.ClientIP()
}

func firstHeader(ctx *gin.Context, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(ctx.GetHeader(k)); v != "" {
			return v
		}
	}
	return ""
}

func modelPlatform(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "web"
	}
	return v
}

// 1x1 transparent GIF
var pixelGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00,
	0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x21, 0xF9, 0x04, 0x01, 0x00, 0x00, 0x00,
	0x00, 0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02,
	0x44, 0x01, 0x00, 0x3B,
}
