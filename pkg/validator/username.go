package validator

import (
	"regexp"
)

var (
	usernameRegex = regexp.MustCompile("^[a-z0-9\\-_'.]+$")
)

func IsValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}
