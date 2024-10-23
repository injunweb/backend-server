package validator

func IsValidPort(port int) bool {
	return port >= 1 && port <= 65535
}
