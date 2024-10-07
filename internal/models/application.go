package models

import "gorm.io/gorm"

const (
	ApplicationStatusPending  string = "Pending"
	ApplicationStatusApproved string = "Approved"
)

type Application struct {
	gorm.Model
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	Status      string `json:"status"`
	OwnerID     uint   `json:"owner_id"`
}
