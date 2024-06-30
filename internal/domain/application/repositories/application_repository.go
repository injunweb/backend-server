package repositories

import (
	"github.com/injunweb/backend-server/internal/domain/application/entities"
	"gorm.io/gorm"
)

type ApplicationRepository struct {
	database *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{database: db}
}

func (applicationRepository *ApplicationRepository) CreateApplication(application *entities.Application) error {
	return applicationRepository.database.Create(&application).Error
}

func (applicationRepository *ApplicationRepository) GetApplications() ([]entities.Application, error) {
	var applications []entities.Application

	err := applicationRepository.database.Find(&applications).Error

	return applications, err
}

func (applicationRepository *ApplicationRepository) GetApplicationByProjectName(projectName string) (*entities.Application, error) {
	var application entities.Application

	err := applicationRepository.database.Where("project_name = ?", projectName).First(&application).Error

	return &application, err
}

func (applicationRepository *ApplicationRepository) DeleteApplication(application *entities.Application) error {
	return applicationRepository.database.Delete(application).Error
}
