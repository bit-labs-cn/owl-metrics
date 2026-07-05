package model

const (
	MetricsAppStatusActive   = 1
	MetricsAppStatusDisabled = 2

	PlatformWeb     = "web"
	PlatformDesktop = "desktop"
	PlatformBackend = "backend"

	EventPageView   = "page_view"
	EventDownload   = "download"
	EventMenuClick  = "menu_click"
	EventFeatureUse = "feature_use"

	RollupDimTotal    = "_total"
	RollupDimPlatform = "platform"
	RollupDimOS       = "os"
	RollupDimURL      = "url"
)
