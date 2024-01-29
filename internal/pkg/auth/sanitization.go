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

func ensureLength(input string, target int) bool {
	return len(input) == target
}

func ensure2FA(input string) bool {
	return ensureLength(input, AUTH_KEY_SIZE) && sanitizeCheck(input)
}

