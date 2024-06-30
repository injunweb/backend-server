package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/domain/application/dto/request"
	"github.com/injunweb/backend-server/internal/domain/application/entities"
	"github.com/injunweb/backend-server/internal/domain/application/repositories"
)

type ApplicationService struct {
	applicationRepository *repositories.ApplicationRepository
}

func NewApplicationService(applicationRepository *repositories.ApplicationRepository) *ApplicationService {
	return &ApplicationService{applicationRepository: applicationRepository}
}

func (applicationService *ApplicationService) CreateApplication(applicationRequestDTO *request.ApplicationRequestDTO) error {
	err := applicationService.CheckProjectName(applicationRequestDTO.ProjectName)

	if err != nil {
		return err
	}

	entitie := entities.Application{
		Email:               applicationRequestDTO.Email,
		ProjectName:         applicationRequestDTO.ProjectName,
		Description:         applicationRequestDTO.Description,
		GithubRepositoryUrl: applicationRequestDTO.GithubRepositoryUrl,
	}

	return applicationService.applicationRepository.CreateApplication(&entitie)
}

func (applicationService *ApplicationService) CheckProjectName(projectName string) error {
	application, _ := applicationService.applicationRepository.GetApplicationByProjectName(projectName)

	if application != nil {
		return errors.New("project name already exists")
	}

	return nil
}
