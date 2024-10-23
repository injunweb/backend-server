package validator

import (
	"regexp"
)

var (
	hostnameRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`)
)

func IsValidHostname(hostname string) bool {
	if len(hostname) < 4 || len(hostname) > 255 {
		return false
	}
	return hostnameRegex.MatchString(hostname)
}
