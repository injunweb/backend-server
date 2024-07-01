package entities

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Email               string `gorm:"not null"`
	ProjectName         string `gorm:"not null;unique"`
	Description         string `gorm:"not null"`
	GithubRepositoryUrl string `gorm:"not null"`
	AccessKey           string `gorm:"not null"`
}
