package security

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) HasRole(role Role) bool {
	return r == role
}
