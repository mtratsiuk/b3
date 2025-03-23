package utils

import (
	"regexp"
	"strings"
)

var tagRe = regexp.MustCompile(`<.*?>`)

func StripHtml(input string) string {
	return strings.TrimSpace(strings.ReplaceAll(tagRe.ReplaceAllString(input, ""), "\n", " "))
}
