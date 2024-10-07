package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique" json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
	IsAdmin  bool   `json:"is_admin"`
}
