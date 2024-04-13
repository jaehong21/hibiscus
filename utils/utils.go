package utils

import (
	"fmt"
	"time"
)

func GetSizeFromByte(byte *int64) string {
	if *byte < 1024 {
		return fmt.Sprintf("%dB", *byte)
	} else if *byte < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(*byte)/1024)
	} else if *byte < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(*byte)/1024/1024)
	} else {
		return fmt.Sprintf("%.2fGB", float64(*byte)/1024/1024/1024)
	}
}

func GetAgeFromTime(timestamp *time.Time) string {
	now := time.Now()
	duration := now.Sub(*timestamp)

	years := int(duration.Hours() / 8760)
	days := int(duration.Hours()/24) % 365
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if years > 0 {
		return fmt.Sprintf("%dy %dd %dh", years, days, hours)
	} else if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
