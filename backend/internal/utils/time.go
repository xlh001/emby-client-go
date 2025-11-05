package utils

import (
	"fmt"
	"time"
)

// TimeToUnixPtr 将时间指针转换为Unix时间戳指针
func TimeToUnixPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	unix := t.Unix()
	return &unix
}

// UnixToTimePtr 将Unix时间戳指针转换为时间指针
func UnixToTimePtr(unix *int64) *time.Time {
	if unix == nil {
		return nil
	}
	t := time.Unix(*unix, 0)
	return &t
}

// FormatDuration 格式化持续时间为易于阅读的字符串
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d小时", int(d.Hours()))
	} else {
		return fmt.Sprintf("%d天", int(d.Hours()/24))
	}
}