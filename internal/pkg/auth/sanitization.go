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

//func sanitizeEmailCheck(input string) bool {
//	_, err := mail.ParseAddress(input)
//	return err == nil
//}
//
//func sanitizePasswordCheck(input string) {}
