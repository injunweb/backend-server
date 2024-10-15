package models

import "gorm.io/gorm"

type Subscription struct {
	gorm.Model
	UserID   uint   `gorm:"not null" json:"user_id"`
	User     User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Endpoint string `gorm:"type:text;not null" json:"endpoint"`
	P256dh   string `gorm:"type:varchar(255);not null" json:"p256dh"`
	Auth     string `gorm:"type:varchar(255);not null" json:"auth"`
}
