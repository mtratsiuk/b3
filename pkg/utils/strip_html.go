package utils

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`(<.*?>)|\s`)
var spaceRe = regexp.MustCompile(`\s+`)
var space = " "

func StripHtml(input string) string {
	return strings.TrimSpace(spaceRe.ReplaceAllString(re.ReplaceAllString(input, space), space))
}
