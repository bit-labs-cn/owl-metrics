package metrics

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Config SDK 配置。
type Config struct {
	Endpoint  string
	AppKey    string
	AppSecret string
	Platform  string
	AnonID    string
	Enabled   bool
	BatchSize int
	FlushMs   int
}

// Event 上报事件。
type Event struct {
	Name       string
	DistinctID string
	SessionID  string
	Props      map[string]any
	AppVersion string
	OS         string
	OSVersion  string
	ClientTS   int64
}

// Client 指标采集客户端。
type Client struct {
	cfg    Config
	http   *http.Client
	mu     sync.Mutex
	queue  []collectEvent
	stopCh chan struct{}
	wg     sync.WaitGroup
}

type collectEvent struct {
	AppKey     string         `json:"appKey"`
	Name       string         `json:"name"`
	DistinctID string         `json:"distinctId,omitempty"`
	AnonID     string         `json:"anonId,omitempty"`
	SessionID  string         `json:"sessionId,omitempty"`
	Platform   string         `json:"platform,omitempty"`
	Props      map[string]any `json:"props,omitempty"`
	AppVersion string         `json:"appVersion,omitempty"`
	OS         string         `json:"os,omitempty"`
	OSVersion  string         `json:"osVersion,omitempty"`
	ClientTS   int64          `json:"clientTs,omitempty"`
}

type collectPayload struct {
	Events []collectEvent `json:"events"`
}

// New 创建 Client；Endpoint 与 AppKey 未配置时禁用采集。
func New(cfg Config) *Client {
	if cfg.Platform == "" {
		cfg.Platform = "desktop"
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 20
	}
	if cfg.FlushMs <= 0 {
		cfg.FlushMs = 3000
	}
	if cfg.AnonID == "" {
		cfg.AnonID = newAnonID()
	}
	c := &Client{
		cfg:    cfg,
		http:   &http.Client{Timeout: 8 * time.Second},
		stopCh: make(chan struct{}),
	}
	if cfg.Endpoint == "" || cfg.AppKey == "" {
		c.cfg.Enabled = false
		return c
	}
	c.cfg.Enabled = true
	c.wg.Add(1)
	go c.loopFlush()
	return c
}

// Track 异步上报事件；失败静默丢弃。
func (c *Client) Track(_ context.Context, ev Event) {
	if c == nil || !c.cfg.Enabled || c.cfg.Endpoint == "" || c.cfg.AppKey == "" {
		return
	}
	if ev.Name == "" {
		return
	}
	if ev.ClientTS == 0 {
		ev.ClientTS = time.Now().UnixMilli()
	}
	row := collectEvent{
		AppKey:     c.cfg.AppKey,
		Name:       ev.Name,
		DistinctID: ev.DistinctID,
		AnonID:     c.cfg.AnonID,
		SessionID:  ev.SessionID,
		Platform:   c.cfg.Platform,
		Props:      ev.Props,
		AppVersion: ev.AppVersion,
		OS:         ev.OS,
		OSVersion:  ev.OSVersion,
		ClientTS:   ev.ClientTS,
	}
	c.mu.Lock()
	c.queue = append(c.queue, row)
	shouldFlush := len(c.queue) >= c.cfg.BatchSize
	c.mu.Unlock()
	if shouldFlush {
		go c.flush()
	}
}

// Flush 立即 flush；退出前调用。
func (c *Client) Flush() {
	c.flush()
}

// Close 停止后台 flush。
func (c *Client) Close() {
	if c == nil {
		return
	}
	select {
	case <-c.stopCh:
		return
	default:
		close(c.stopCh)
	}
	c.flush()
	c.wg.Wait()
}

func (c *Client) loopFlush() {
	defer c.wg.Done()
	ticker := time.NewTicker(time.Duration(c.cfg.FlushMs) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.flush()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Client) flush() {
	c.mu.Lock()
	if len(c.queue) == 0 {
		c.mu.Unlock()
		return
	}
	batch := append([]collectEvent(nil), c.queue...)
	c.queue = c.queue[:0]
	c.mu.Unlock()

	body := collectPayload{Events: batch}
	raw, err := json.Marshal(body)
	if err != nil {
		return
	}
	endpoint := trimRightSlash(c.cfg.Endpoint) + "/api/v1/metrics/collect"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	req.Header.Set("X-Metrics-Timestamp", ts)
	if c.cfg.AppSecret != "" {
		req.Header.Set("X-Metrics-Signature", sign(c.cfg.AppSecret, ts, raw))
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func sign(secret, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("\n"))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func trimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func newAnonID() string {
	return uuidNew()
}

func uuidNew() string {
	return uuidString()
}
