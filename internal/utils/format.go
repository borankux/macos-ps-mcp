package utils

import (
	"fmt"
	"math"
)

// FormatBytes converts bytes to human readable format
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatCPU formats CPU percentage
func FormatCPU(cpu float64) string {
	if cpu < 0.01 {
		return "< 0.01%"
	}
	if cpu >= 100 {
		return "100%"
	}
	return fmt.Sprintf("%.2f%%", math.Round(cpu*100)/100)
}

// FormatDuration formats duration to human readable
func FormatDuration(seconds uint64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	}
	return fmt.Sprintf("%dd %dh", seconds/86400, (seconds%86400)/3600)
}

