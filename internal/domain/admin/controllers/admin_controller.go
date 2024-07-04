package controllers

import (
	"github.com/injunweb/backend-server/internal/domain/admin/dto/request"
	"github.com/injunweb/backend-server/internal/domain/admin/dto/response"
	"github.com/injunweb/backend-server/internal/domain/admin/services"
	"github.com/injunweb/backend-server/internal/interfaces/handlers"
)

type AdminController struct {
	adminService *services.AdminService
}

func NewAdminController(adminService *services.AdminService) *AdminController {
	return &AdminController{adminService: adminService}
}

func (adminController *AdminController) Login(adminRequestDTO *request.AdminRequestDTO) (string, error) {
	if err := adminRequestDTO.Validate(); err != nil {
		return "", err
	}

	token, err := adminController.adminService.Login(adminRequestDTO)
	return token, err
}

func (adminController *AdminController) GetApplications() ([]response.ApplicationResponseDTO, error) {
	applications, err := adminController.adminService.GetApplications()
	return applications, err
}

func (adminController *AdminController) GetApplication(params handlers.Params) (*response.ApplicationResponseDTO, error) {
	projectName := params.Get("project_name")
	application, err := adminController.adminService.GetApplication(projectName)
	return application, err
}

func (adminController *AdminController) GetProjects() ([]response.ProjectResponseDTO, error) {
	projects, err := adminController.adminService.GetProjects()
	return projects, err
}

func (adminController *AdminController) GetProject(params handlers.Params) (*response.ProjectResponseDTO, error) {
	projectName := params.Get("project_name")
	project, err := adminController.adminService.GetProject(projectName)
	return project, err
}

func (adminController *AdminController) ApproveApplication(params handlers.Params) error {
	projectName := params.Get("project_name")
	err := adminController.adminService.ApproveApplication(projectName)
	return err
}

func (adminController *AdminController) RejectApplication(params handlers.Params) error {
	projectName := params.Get("project_name")
	err := adminController.adminService.RejectApplication(projectName)
	return err
}
