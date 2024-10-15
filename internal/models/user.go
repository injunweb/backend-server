package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username      string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"username"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"not null" json:"-"`
	Subscriptions []Subscription `gorm:"foreignKey:UserID" json:"-"`
	IsAdmin       bool           `gorm:"default:false" json:"is_admin"`
	Applications  []Application  `gorm:"foreignKey:OwnerID" json:"applications,omitempty"`
}
