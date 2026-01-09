package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatBytes converts a size in bytes (uint64) to a human-readable
// string using 1024-based units. Examples:
//
//	0 -> "0 B"
//	512 -> "512 B"
//	1536 -> "1.5 KB"
//	1048576 -> "1 MB"
func FormatBytes(b uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	val := float64(b)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}

	// For plain bytes (no fractional part), prefer integer formatting.
	if i == 0 {
		return fmt.Sprintf("%d B", b)
	}

	// Format with up to 2 decimal places, then trim trailing zeros.
	s := strconv.FormatFloat(val, 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return fmt.Sprintf("%s %s", s, units[i])
}
