package models

import "gorm.io/gorm"

type ExtraHostnames struct {
	gorm.Model
	Hostname      string      `gorm:"uniqueIndex;not null" json:"hostname"`
	ApplicationID uint        `gorm:"not null" json:"application_id"`
	Application   Application `gorm:"foreignKey:ApplicationID" json:"application,omitempty"`
}
