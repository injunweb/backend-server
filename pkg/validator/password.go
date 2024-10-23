package validator

import (
	"regexp"
)

var (
	passwordRegex = regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*#?&])[A-Za-z\d@$!%*#?&]{8,}$`)
)

func IsValidPassword(password string) bool {
	return passwordRegex.MatchString(password)
}
