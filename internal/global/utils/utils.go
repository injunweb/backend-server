package utils

import (
	"fmt"
	"regexp"
	"runtime"
)

func IsEmptyString(str string) bool {
	return str == ""
}

func IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	regex := regexp.MustCompile(pattern)

	return regex.MatchString(email)
}

func IsGithubRepositoryUrl(url string) bool {
	pattern := `^(https:\/\/github.com\/)([a-zA-Z0-9-]+)(\/)([a-zA-Z0-9-]+)$`

	regex := regexp.MustCompile(pattern)

	return regex.MatchString(url)
}

func GetGoroutineID() uint64 {
	return uint64(goID())
}

func goID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stack := string(buf[:n])
	var id int
	fmt.Sscanf(stack, "goroutine %d ", &id)
	return id
}
