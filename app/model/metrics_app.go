package model

import "bit-labs.cn/owl/provider/db"

// MetricsApp 指标采集应用（多产品隔离）。
type MetricsApp struct {
	db.BaseModel
	Name           string `gorm:"comment:应用名称;size:120;not null" json:"name"`
	AppKey         string `gorm:"comment:公开AppKey;size:64;not null;uniqueIndex:idx_metrics_app_key" json:"appKey"`
	AppSecret      string `gorm:"comment:服务端签名密钥;size:128;not null" json:"appSecret"`
	AllowedOrigins string `gorm:"comment:允许的网页来源JSON;type:text" json:"allowedOrigins"`
	Status         int    `gorm:"comment:状态1启用2禁用;default:1;index:idx_metrics_app_status" json:"status"`
	Description    string `gorm:"comment:描述;size:500" json:"description"`
}

// TableName 返回表名。
func (MetricsApp) TableName() string { return "metrics_app" }
