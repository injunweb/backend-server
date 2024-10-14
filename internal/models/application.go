package models

import (
	"gorm.io/gorm"
)

const (
	ApplicationStatusPending  string = "Pending"
	ApplicationStatusApproved string = "Approved"
)

type Application struct {
	gorm.Model
	Name            string           `gorm:"uniqueIndex;not null" json:"name"`
	GitURL          string           `gorm:"not null" json:"git_url"`
	Branch          string           `gorm:"not null" json:"branch"`
	Port            int              `gorm:"not null" json:"port"`
	Description     string           `json:"description"`
	Status          string           `gorm:"default:'Pending'" json:"status"`
	OwnerID         uint             `gorm:"not null" json:"owner_id"`
	Owner           User             `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	PrimaryHostname string           `gorm:"uniqueIndex;not null" json:"primary_hostname"`
	ExtraHostnames  []ExtraHostnames `gorm:"foreignKey:ApplicationID" json:"extra_hostnames,omitempty"`
}
