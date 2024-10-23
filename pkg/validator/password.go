package validator

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '!' || char == '@' || char == '#' || char == '$' ||
			char == '%' || char == '^' || char == '&' || char == '*' ||
			char == '-' || char == '_':
			hasSpecial = true
		default:
			return false
		}
	}

	return hasLower && hasDigit && hasSpecial
}
