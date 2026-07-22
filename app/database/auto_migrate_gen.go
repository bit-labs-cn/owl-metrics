package database

import (
	"bit-labs.cn/owl-metrics/app/model"
)

// Models 返回待自动迁移的 model 列表。
func Models() []any {
	return []any{
		&model.MetricsApp{},
		&model.MetricsEvent{},
		&model.MetricsRollupDaily{},
	}
}
