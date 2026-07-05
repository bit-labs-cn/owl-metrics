package useragent

import (
	"regexp"
	"strings"
)

var (
	reWindows = regexp.MustCompile(`Windows NT ([0-9.]+)`)
	reMac     = regexp.MustCompile(`Mac OS X ([0-9_\.]+)`)
	reLinux   = regexp.MustCompile(`Linux`)
	reAndroid = regexp.MustCompile(`Android ([0-9.]+)`)
	reIOS     = regexp.MustCompile(`iPhone OS ([0-9_]+)`)
)

// Parse 轻量解析 UA，返回 os 与 osVersion。
func Parse(ua string) (os, osVersion string) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "", ""
	}
	switch {
	case strings.Contains(ua, "Windows"):
		os = "windows"
		if m := reWindows.FindStringSubmatch(ua); len(m) > 1 {
			osVersion = m[1]
		}
	case strings.Contains(ua, "Mac OS X"):
		os = "macos"
		if m := reMac.FindStringSubmatch(ua); len(m) > 1 {
			osVersion = strings.ReplaceAll(m[1], "_", ".")
		}
	case reAndroid.MatchString(ua):
		os = "android"
		if m := reAndroid.FindStringSubmatch(ua); len(m) > 1 {
			osVersion = m[1]
		}
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		os = "ios"
		if m := reIOS.FindStringSubmatch(ua); len(m) > 1 {
			osVersion = strings.ReplaceAll(m[1], "_", ".")
		}
	case reLinux.MatchString(ua):
		os = "linux"
	default:
		os = "unknown"
	}
	return os, osVersion
}
