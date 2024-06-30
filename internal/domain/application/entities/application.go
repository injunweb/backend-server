package entities

import "gorm.io/gorm"

type Application struct {
	gorm.Model
	Email               string `gorm:"not null"`
	ProjectName         string `gorm:"not null;unique"`
	Description         string `gorm:"not null"`
	GithubRepositoryUrl string `gorm:"not null"`
}
