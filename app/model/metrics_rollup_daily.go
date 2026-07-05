package model

import "bit-labs.cn/owl/provider/db"

// MetricsRollupDaily 每日指标汇总。
type MetricsRollupDaily struct {
	db.BaseModel
	AppID    uint   `gorm:"comment:应用ID;uniqueIndex:idx_metrics_rollup_unique" json:"appId"`
	Day      string `gorm:"comment:日期;size:10;uniqueIndex:idx_metrics_rollup_unique" json:"day"`
	Name     string `gorm:"comment:事件名;size:64;uniqueIndex:idx_metrics_rollup_unique" json:"name"`
	DimKey   string `gorm:"comment:维度键;size:64;uniqueIndex:idx_metrics_rollup_unique" json:"dimKey"`
	DimValue string `gorm:"comment:维度值;size:128;uniqueIndex:idx_metrics_rollup_unique" json:"dimValue"`
	Count    int64  `gorm:"comment:计数;default:0" json:"count"`
	Uniques  int64  `gorm:"comment:去重UV;default:0" json:"uniques"`
}

// TableName 返回表名。
func (MetricsRollupDaily) TableName() string { return "metrics_rollup_daily" }
