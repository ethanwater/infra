package auth

import (
	"net/mail"
	"regexp"
)

var whitelist_text *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9]+")

// Sanitize the user input to only allow valid characters
func Sanitize(input string) string {
	return whitelist_text.ReplaceAllString(input, "")
}

// Checks if the user input complies with Santiziation constraints
func SanitizeCheck(input string) bool {
	return whitelist_text.ReplaceAllString(input, "") == input
}

func SanitizeEmailCheck(input string) bool {
	_, err := mail.ParseAddress(input)
	return err == nil
}

func SanitizePasswordCheck(input string) {
	//TODO:
	//var whitelist_password *regexp.Regexp = regexp.MustCompile("((?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[\W]).{6,20})")
	//return whitelist_password.ReplaceAllString(input, "") == input
}
