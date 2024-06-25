package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/injunweb/backend-server/env"
)

type Claims struct {
	ID    string `json:"id"`
	Roles Role   `json:"roles"`
	jwt.StandardClaims
}

var jwtKey []byte

func init() {
	jwtKey = []byte(env.JWT_SECRET_KEY)
	if len(jwtKey) == 0 {
		panic("JWT_SECRET_KEY is not set")
	}
}

func GenerateToken(id string, roles Role) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		ID:    id,
		Roles: roles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (string, Role, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", "", errors.New("invalid token signature")
		}
		return "", "", errors.New("invalid token")
	}

	if !token.Valid {
		return "", "", errors.New("invalid token")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return "", "", errors.New("token has expired")
	}

	return claims.ID, claims.Roles, nil
}
