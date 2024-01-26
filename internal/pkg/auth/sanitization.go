package auth

import (
	"regexp"
)

var WHITELIST *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9]+")

func sanitize(input string) string {
	return WHITELIST.ReplaceAllString(input, "")
}

func sanitizeCheck(input string) bool {
	return WHITELIST.ReplaceAllString(input, "") == input
}
