package template

import (
	"fmt"
	"strings"
)

// formatDuration converts fractional seconds to M:SS display format.
// Sub-second precision is intentionally truncated — display only.
func formatDuration(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

// extractTimestamp returns the [mm:ss.xx] part of a synced lyric line.
func extractTimestamp(line string) string {
	if len(line) > 0 && line[0] == '[' {
		end := strings.Index(line, "]")
		if end > 1 {
			return line[:end+1]
		}
	}
	return ""
}

// extractLyricText returns the text after the timestamp in a synced lyric line.
func extractLyricText(line string) string {
	if len(line) > 0 && line[0] == '[' {
		end := strings.Index(line, "]")
		if end > 0 && end+1 < len(line) {
			return strings.TrimSpace(line[end+1:])
		}
	}
	return line
}
