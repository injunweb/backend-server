package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/database"
	"github.com/injunweb/backend-server/pkg/github"
	"github.com/injunweb/backend-server/pkg/harbor"
	"github.com/injunweb/backend-server/pkg/kubernetes"
	"github.com/injunweb/backend-server/pkg/vault"

	"gorm.io/gorm"
)

type ApplicationService struct {
	db *gorm.DB
}

func NewApplicationService(db *gorm.DB) *ApplicationService {
	return &ApplicationService{db: db}
}

type GetApplicationsResponse struct {
	Applications []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"applications"`
}

func (s *ApplicationService) GetApplications(userId uint) (GetApplicationsResponse, error) {
	var applications []models.Application
	if err := s.db.Where("owner_id = ?", userId).Find(&applications).Error; err != nil {
		return GetApplicationsResponse{}, errors.New("failed to retrieve applications")
	}

	var response GetApplicationsResponse
	for _, app := range applications {
		response.Applications = append(response.Applications, struct {
			ID     uint   `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		}{
			ID:     app.ID,
			Name:   app.Name,
			Status: app.Status,
		})
	}

	return response, nil
}

type SubmitApplicationRequest struct {
	Name        string `json:"name" binding:"required"`
	GitURL      string `json:"git_url" binding:"required"`
	Branch      string `json:"branch" binding:"required"`
	Port        int    `json:"port" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type SubmitApplicationResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) SubmitApplication(userId uint, req SubmitApplicationRequest) (SubmitApplicationResponse, error) {
	pattern := `^[a-z0-9\-]+$`
	forbiddenKeywords := []string{"--", "#", ";", "SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "EXEC", "UNION", "OR", "AND"}

	if matched, err := regexp.MatchString(pattern, req.Name); !matched || err != nil {
		return SubmitApplicationResponse{}, errors.New("invalid application name format")
	}

	nameUpper := strings.ToUpper(req.Name)
	for _, keyword := range forbiddenKeywords {
		if strings.Contains(nameUpper, keyword) {
			return SubmitApplicationResponse{}, errors.New("application name contains forbidden characters or SQL keywords")
		}
	}

	var existingApp models.Application
	if err := s.db.Where("name = ?", req.Name).First(&existingApp).Error; err == nil {
		return SubmitApplicationResponse{}, errors.New("application name already exists")
	}

	if req.Port < 1 || req.Port > 65535 {
		return SubmitApplicationResponse{}, errors.New("invalid port number")
	}

	application := models.Application{
		Name:            req.Name,
		GitURL:          req.GitURL,
		Branch:          req.Branch,
		Port:            req.Port,
		Description:     req.Description,
		Status:          models.ApplicationStatusPending,
		PrimaryHostname: fmt.Sprintf("%s.%s", req.Name, "ijw.app"),
		ExtraHostnames:  []models.ExtraHostnames{},
		OwnerID:         userId,
	}

	if err := s.db.Create(&application).Error; err != nil {
		return SubmitApplicationResponse{}, errors.New("failed to submit application")
	}

	return SubmitApplicationResponse{
		Message: "Application submitted successfully",
	}, nil
}

type GetApplicationResponse struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	GitURL          string   `json:"git_url"`
	Branch          string   `json:"branch"`
	Port            int      `json:"port"`
	Description     string   `json:"description"`
	OwnerID         uint     `json:"owner_id"`
	Status          string   `json:"status"`
	PrimaryHostname string   `json:"primary_hostname"`
	ExtraHostnames  []string `json:"extra_hostnames"`
}

func (s *ApplicationService) GetApplication(userId uint, appId uint) (GetApplicationResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return GetApplicationResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return GetApplicationResponse{}, errors.New("permission denied")
	}

	return GetApplicationResponse{
		ID:              application.ID,
		Name:            application.Name,
		GitURL:          application.GitURL,
		Branch:          application.Branch,
		Port:            application.Port,
		Description:     application.Description,
		OwnerID:         application.OwnerID,
		Status:          application.Status,
		PrimaryHostname: application.PrimaryHostname,
		ExtraHostnames: func() []string {
			var extraHostnames []string
			for _, hostname := range application.ExtraHostnames {
				extraHostnames = append(extraHostnames, hostname.Hostname)
			}
			return extraHostnames
		}(),
	}, nil
}

type DeleteApplicationResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) DeleteApplication(userId uint, appId uint) (DeleteApplicationResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return DeleteApplicationResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return DeleteApplicationResponse{}, errors.New("permission denied")
	}

	if application.Status == models.ApplicationStatusApproved {
		if kubernetes.NamespaceExists(application.Name) {
			if err := kubernetes.DeleteNamespace(application.Name); err != nil {
				return DeleteApplicationResponse{}, fmt.Errorf("failed to delete namespace: %v", err)
			}
		}

		if exists, err := harbor.RepositoryExists(application.Name); err != nil {
			return DeleteApplicationResponse{}, fmt.Errorf("failed to check Harbor repository: %v", err)
		} else if exists {
			if err := harbor.DeleteRepository(application.Name); err != nil {
				return DeleteApplicationResponse{}, fmt.Errorf("failed to delete Harbor repository: %v", err)
			}
		}

		if err := vault.DeleteSecret(application.Name); err != nil {
			return DeleteApplicationResponse{}, fmt.Errorf("failed to delete secret: %v", err)
		}

		if err := database.DeleteDatabaseAndUser(application.Name); err != nil {
			return DeleteApplicationResponse{}, fmt.Errorf("failed to delete database and user: %v", err)
		}

		if err := github.TriggerRemovePipelineWorkflow(application); err != nil {
			return DeleteApplicationResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
		}
	}

	if err := s.db.Delete(&application).Error; err != nil {
		return DeleteApplicationResponse{}, errors.New("failed to delete application")
	}

	return DeleteApplicationResponse{
		Message: "Application deleted successfully",
	}, nil
}

type AddAdditionalHostnameRequest struct {
	Hostname string `json:"hostname" binding:"required"`
}

type AddAdditionalHostnameResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) AddExtralHostname(userId uint, appId uint, req AddAdditionalHostnameRequest) (AddAdditionalHostnameResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return AddAdditionalHostnameResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return AddAdditionalHostnameResponse{}, errors.New("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return AddAdditionalHostnameResponse{}, errors.New("application not approved")
	}

	if application.PrimaryHostname == req.Hostname {
		return AddAdditionalHostnameResponse{}, errors.New("hostname already exists")
	}

	for _, hostname := range application.ExtraHostnames {
		if hostname.Hostname == req.Hostname {
			return AddAdditionalHostnameResponse{}, errors.New("hostname already exists")
		}
	}

	hostname := models.ExtraHostnames{
		ApplicationID: application.ID,
		Hostname:      req.Hostname,
	}

	if err := s.db.Create(&hostname).Error; err != nil {
		return AddAdditionalHostnameResponse{}, errors.New("failed to add additional hostname")
	}

	if err := github.TriggerAddAdditionalHostnameWorkflow(application, req.Hostname); err != nil {
		return AddAdditionalHostnameResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
	}

	return AddAdditionalHostnameResponse{
		Message: "Additional hostname added successfully",
	}, nil
}

type DeleteAdditionalHostnameRequest struct {
	Hostname string `json:"hostname" binding:"required"`
}

type DeleteAdditionalHostnameResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) DeleteExtraHostname(userId uint, appId uint, req DeleteAdditionalHostnameRequest) (DeleteAdditionalHostnameResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return DeleteAdditionalHostnameResponse{}, errors.New("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return DeleteAdditionalHostnameResponse{}, errors.New("application not approved")
	}

	if application.PrimaryHostname == req.Hostname {
		return DeleteAdditionalHostnameResponse{}, errors.New("cannot delete custom hostname")
	}

	var hostname models.ExtraHostnames
	if err := s.db.Where("application_id = ? AND hostname = ?", application.ID, req.Hostname).First(&hostname).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.New("hostname not found")
	}

	if err := s.db.Delete(&hostname).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.New("failed to delete additional hostname")
	}

	if err := github.TriggerDeleteAdditionalHostnameWorkflow(application, req.Hostname); err != nil {
		return DeleteAdditionalHostnameResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
	}

	return DeleteAdditionalHostnameResponse{
		Message: "Additional hostname deleted successfully",
	}, nil
}

type GetEnvironmentsResponse struct {
	Environments []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"environments"`
}

func (s *ApplicationService) GetEnvironments(userId uint, appId uint) (GetEnvironmentsResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return GetEnvironmentsResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return GetEnvironmentsResponse{}, errors.New("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return GetEnvironmentsResponse{}, errors.New("application not approved")
	}

	secret, err := vault.GetSecret(application.Name)
	if err != nil {
		return GetEnvironmentsResponse{}, fmt.Errorf("failed to read from Vault: %v", err)
	}

	var response GetEnvironmentsResponse
	for key, value := range secret {
		response.Environments = append(response.Environments, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{
			Key:   key,
			Value: value.(string),
		})
	}

	return response, nil
}

type UpdateEnvironmentRequest struct {
	Environments []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"environments"`
}

type UpdateEnvironmentResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) UpdateEnvironment(userId uint, appId uint, req UpdateEnvironmentRequest) (UpdateEnvironmentResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return UpdateEnvironmentResponse{}, errors.New("application not found")
	}

	if application.OwnerID != userId {
		return UpdateEnvironmentResponse{}, errors.New("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return UpdateEnvironmentResponse{}, errors.New("application not approved")
	}

	data := make(map[string]interface{})
	for _, env := range req.Environments {
		data[env.Key] = env.Value
	}

	if err := vault.UpdateSecret(application.Name, data); err != nil {
		return UpdateEnvironmentResponse{}, fmt.Errorf("failed to write to Vault: %v", err)
	}

	return UpdateEnvironmentResponse{
		Message: "Environment updated successfully",
	}, nil
}
