package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/domain/admin/dto/request"
	"github.com/injunweb/backend-server/internal/domain/admin/dto/response"
	adminRepository "github.com/injunweb/backend-server/internal/domain/admin/repositories"
	applicationRepository "github.com/injunweb/backend-server/internal/domain/application/repositories"
	"github.com/injunweb/backend-server/internal/domain/project/entities"
	projectRepository "github.com/injunweb/backend-server/internal/domain/project/repositories"
	"github.com/injunweb/backend-server/internal/global/integrations/mail"
	"github.com/injunweb/backend-server/internal/global/security"
)

type AdminService struct {
	adminRepository       *adminRepository.AdminRepository
	applicationRepository *applicationRepository.ApplicationRepository
	projectRepository     *projectRepository.ProjectRepository
}

func NewAdminService(adminRepository *adminRepository.AdminRepository, applicationRepository *applicationRepository.ApplicationRepository, projectRepository *projectRepository.ProjectRepository) *AdminService {
	return &AdminService{adminRepository: adminRepository, applicationRepository: applicationRepository, projectRepository: projectRepository}
}

func (adminService *AdminService) Login(adminRequestDTO *request.AdminRequestDTO) (string, error) {
	admin, err := adminService.adminRepository.GetAdminByUsername(adminRequestDTO.AdminName)

	if err != nil {
		return "", err
	}

	if admin.Password != adminRequestDTO.Password {
		return "", errors.New("invalid password")
	}

	token, err := security.GenerateToken(admin.AdminName, security.RoleAdmin)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (adminService *AdminService) GetApplications() ([]response.ApplicationResponseDTO, error) {
	applications, err := adminService.applicationRepository.GetApplications()

	if err != nil {
		return nil, err
	}

	var applicationResponseDTOs []response.ApplicationResponseDTO

	for _, application := range applications {
		applicationResponseDTOs = append(applicationResponseDTOs, response.ApplicationResponseDTO{
			Email:               application.Email,
			ProjectName:         application.ProjectName,
			Description:         application.Description,
			GithubRepositoryUrl: application.GithubRepositoryUrl,
		})
	}

	return applicationResponseDTOs, nil
}

func (adminService *AdminService) GetApplication(projectName string) (*response.ApplicationResponseDTO, error) {
	application, err := adminService.applicationRepository.GetApplicationByProjectName(projectName)

	if err != nil {
		return nil, err
	}

	if application == nil {
		return nil, errors.New("application not found")
	}

	applicationResponseDTO := response.ApplicationResponseDTO{
		Email:               application.Email,
		ProjectName:         application.ProjectName,
		Description:         application.Description,
		GithubRepositoryUrl: application.GithubRepositoryUrl,
	}

	return &applicationResponseDTO, nil
}

func (adminService *AdminService) GetProject(projectName string) (*response.ProjectResponseDTO, error) {
	project, err := adminService.projectRepository.GetProjectByProjectName(projectName)

	if err != nil {
		return nil, err
	}

	if project == nil {
		return nil, errors.New("project not found")
	}

	projectResponseDTO := response.ProjectResponseDTO{
		Email:               project.Email,
		ProjectName:         project.ProjectName,
		Description:         project.Description,
		GithubRepositoryUrl: project.GithubRepositoryUrl,
		AccessKey:           project.AccessKey,
	}

	return &projectResponseDTO, nil
}

func (adminService *AdminService) GetProjects() ([]response.ProjectResponseDTO, error) {
	projects, err := adminService.projectRepository.GetProjects()

	if err != nil {
		return nil, err
	}

	var projectResponseDTOs []response.ProjectResponseDTO

	for _, project := range projects {
		projectResponseDTOs = append(projectResponseDTOs, response.ProjectResponseDTO{
			Email:               project.Email,
			ProjectName:         project.ProjectName,
			Description:         project.Description,
			GithubRepositoryUrl: project.GithubRepositoryUrl,
			AccessKey:           project.AccessKey,
		})
	}

	return projectResponseDTOs, nil
}

func (adminService *AdminService) ApproveApplication(projectName string) error {
	application, err := adminService.applicationRepository.GetApplicationByProjectName(projectName)

	if err != nil {
		return err
	}

	if application == nil {
		return errors.New("application not found")
	}

	err = adminService.applicationRepository.DeleteApplication(application)

	if err != nil {
		return err
	}

	accessKey, err := security.GenerateRandomString(16)

	if err != nil {
		return err
	}

	project := entities.Project{
		Email:               application.Email,
		ProjectName:         application.ProjectName,
		Description:         application.Description,
		GithubRepositoryUrl: application.GithubRepositoryUrl,
		AccessKey:           accessKey,
	}

	err = adminService.projectRepository.CreateProject(&project)

	if err != nil {
		return err
	}

	mailTo := []string{application.Email}
	mailSubject := "Application Approved"
	mailBody := "Your application has been approved. Your access key is " + accessKey

	err = mail.SendEmail(mailTo, mailSubject, mailBody)

	if err != nil {
		return err
	}

	return nil
}

func (adminService *AdminService) RejectApplication(projectName string) error {
	application, err := adminService.applicationRepository.GetApplicationByProjectName(projectName)

	if err != nil {
		return err
	}

	if application == nil {
		return errors.New("application not found")
	}

	err = adminService.applicationRepository.DeleteApplication(application)

	if err != nil {
		return err
	}

	mailTo := []string{application.Email}
	mailSubject := "Application Rejected"
	mailBody := "Your application has been rejected"

	err = mail.SendEmail(mailTo, mailSubject, mailBody)

	if err != nil {
		return err
	}

	return nil
}
