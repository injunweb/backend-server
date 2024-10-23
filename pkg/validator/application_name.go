package validator

import (
	"regexp"
	"strings"
)

var (
	appNameRegex      = regexp.MustCompile(`^[a-z0-9\-]+$`)
	forbiddenKeywords = []string{
		"--", "#", ";",
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"DROP", "EXEC", "UNION", "OR", "AND",
	}
)

func IsValidApplicationName(name string) bool {
	if !appNameRegex.MatchString(name) {
		return false
	}

	nameUpper := strings.ToUpper(name)
	for _, keyword := range forbiddenKeywords {
		if strings.Contains(nameUpper, keyword) {
			return false
		}
	}
	return true
}
