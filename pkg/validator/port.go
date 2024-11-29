package validator

func IsValidPort(port string) bool {
	return port >= "1" && port <= "65535"
}
