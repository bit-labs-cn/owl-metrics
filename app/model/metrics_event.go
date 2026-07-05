package model

import "bit-labs.cn/owl/provider/db"

// MetricsEvent 指标事件明细（追加式）。
type MetricsEvent struct {
	db.BaseModel
	AppID      uint   `gorm:"comment:应用ID;index:idx_metrics_event_app_day" json:"appId"`
	AppKey     string `gorm:"comment:AppKey;size:64;index:idx_metrics_event_app_key" json:"appKey"`
	Name       string `gorm:"comment:事件名;size:64;not null;index:idx_metrics_event_name" json:"name"`
	DistinctID string `gorm:"comment:用户标识;size:128;index:idx_metrics_event_distinct" json:"distinctId"`
	AnonID     string `gorm:"comment:匿名设备ID;size:128;index:idx_metrics_event_anon" json:"anonId"`
	SessionID  string `gorm:"comment:会话ID;size:128" json:"sessionId"`
	Platform   string `gorm:"comment:平台web/desktop/backend;size:32;index:idx_metrics_event_platform" json:"platform"`
	Props      string `gorm:"comment:自定义属性JSON;type:text" json:"props"`
	URL        string `gorm:"comment:页面URL;size:1024" json:"url"`
	Referrer   string `gorm:"comment:来源;size:1024" json:"referrer"`
	UA         string `gorm:"comment:User-Agent;size:512" json:"ua"`
	OS         string `gorm:"comment:操作系统;size:32" json:"os"`
	OSVersion  string `gorm:"comment:系统版本;size:64" json:"osVersion"`
	AppVersion string `gorm:"comment:应用版本;size:64" json:"appVersion"`
	IP         string `gorm:"comment:IP;size:64" json:"ip"`
	Region     string `gorm:"comment:地域;size:64" json:"region"`
	ClientTS   int64  `gorm:"comment:客户端时间戳毫秒" json:"clientTs"`
	ServerTS   int64  `gorm:"comment:服务端时间戳毫秒;index:idx_metrics_event_server_ts" json:"serverTs"`
	Day        string `gorm:"comment:日期YYYY-MM-DD;size:10;index:idx_metrics_event_app_day" json:"day"`
}

// TableName 返回表名。
func (MetricsEvent) TableName() string { return "metrics_event" }
