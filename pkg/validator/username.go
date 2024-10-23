package validator

import (
	"regexp"
)

var (
	usernameRegex = regexp.MustCompile("^[a-z0-9\\-_'.]+$")
)

func IsValidUsername(username string) bool {
	if len(username) < 4 || len(username) > 8 {
		return false
	}
	return usernameRegex.MatchString(username)
}
