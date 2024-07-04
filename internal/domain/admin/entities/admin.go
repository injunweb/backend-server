package entities

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	AdminName string `gorm:"not null;unique"`
	Password  string `gorm:"not null"`
}
