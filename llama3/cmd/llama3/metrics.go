package llama3cmd

import (
	"fmt"
	"time"
)

// formatLatency formats a duration into a human-readable string with appropriate units.
func formatLatency(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%dns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// calculateTPS calculates tokens per second.
func calculateTPS(tokenCount int, duration time.Duration) int {
	if duration == 0 {
		return 0
	}
	return int(float64(tokenCount) / duration.Seconds())
}
