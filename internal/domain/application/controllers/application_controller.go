package controllers

import (
	applicationRequest "github.com/injunweb/backend-server/internal/domain/application/dto/request"
	applicationService "github.com/injunweb/backend-server/internal/domain/application/services"
	"github.com/injunweb/backend-server/internal/interfaces/handlers"
)

type ApplicationController struct {
	applicationService *applicationService.ApplicationService
}

func NewApplicationController(applicationService *applicationService.ApplicationService) *ApplicationController {
	return &ApplicationController{applicationService: applicationService}
}

func (applicationController *ApplicationController) CreateApplication(applicationRequestDTO *applicationRequest.ApplicationRequestDTO) error {
	if err := applicationRequestDTO.Validate(); err != nil {
		return err
	}

	if err := applicationController.applicationService.CreateApplication(applicationRequestDTO); err != nil {
		return err
	}

	return nil
}

func (applicationController *ApplicationController) CheckProjectName(query handlers.Query) error {
	projectName := query.Get("project_name")

	if err := applicationController.applicationService.CheckProjectName(projectName); err != nil {
		return err
	}

	return nil
}
