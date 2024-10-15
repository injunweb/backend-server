package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	UserID  uint   `gorm:"not null" json:"user_id"`
	User    User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Message string `gorm:"type:varchar(255);not null" json:"message"`
	IsRead  bool   `gorm:"default:false" json:"is_read"`
}
