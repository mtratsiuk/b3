package utils

import (
	"strings"
)

func TrimText(input string, breakAt int) string {
	if len(input)-1 <= breakAt {
		return input
	}

	lastSpace := strings.LastIndex(input[:breakAt+1], " ")

	if lastSpace == -1 {
		lastSpace = breakAt
	}

	return strings.TrimSpace(input[:lastSpace]) + "..."
}
