package database

import (
	"bit-labs.cn/owl-metrics/app/model"
	"gorm.io/gorm"
)

// Migrate 执行数据库迁移。
func Migrate(db *gorm.DB) {
	_ = db.AutoMigrate(
		&model.MetricsApp{},
		&model.MetricsEvent{},
		&model.MetricsRollupDaily{},
	)
}
