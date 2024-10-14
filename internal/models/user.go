package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string        `gorm:"uniqueIndex;not null" json:"username"`
	Email        string        `gorm:"uniqueIndex;not null" json:"email"`
	Password     string        `gorm:"not null" json:"-"`
	IsAdmin      bool          `gorm:"default:false" json:"is_admin"`
	Applications []Application `gorm:"foreignKey:OwnerID" json:"applications,omitempty"`
}
