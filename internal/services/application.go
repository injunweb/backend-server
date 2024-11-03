package services

import (
	"fmt"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/database"
	"github.com/injunweb/backend-server/pkg/errors"
	"github.com/injunweb/backend-server/pkg/github"
	"github.com/injunweb/backend-server/pkg/harbor"
	"github.com/injunweb/backend-server/pkg/kubernetes"
	"github.com/injunweb/backend-server/pkg/validator"
	"github.com/injunweb/backend-server/pkg/vault"

	"gorm.io/gorm"
)

type ApplicationService struct {
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewApplicationService(db *gorm.DB, notificationService *NotificationService) *ApplicationService {
	return &ApplicationService{db: db, notificationService: notificationService}
}

type GetApplicationsResponse struct {
	Applications []struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		GitURL      string `json:"git_url"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
	} `json:"applications"`
}

func (s *ApplicationService) GetApplications(userId uint) (GetApplicationsResponse, errors.CustomError) {
	var applications []models.Application
	if err := s.db.Where("owner_id = ?", userId).Find(&applications).Error; err != nil {
		return GetApplicationsResponse{}, errors.Internal("failed to retrieve applications")
	}

	var response GetApplicationsResponse
	for _, app := range applications {
		response.Applications = append(response.Applications, struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
			GitURL      string `json:"git_url"`
			Description string `json:"description"`
			CreatedAt   string `json:"created_at"`
		}{
			ID:          app.ID,
			Name:        app.Name,
			Status:      app.Status,
			GitURL:      app.GitURL,
			Description: app.Description,
			CreatedAt:   app.CreatedAt.Format("2006-01-02 15:04:05"),
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

func (s *ApplicationService) SubmitApplication(userId uint, req SubmitApplicationRequest) (SubmitApplicationResponse, errors.CustomError) {
	if !validator.IsValidApplicationName(req.Name) {
		return SubmitApplicationResponse{}, errors.BadRequest("invalid application name")
	}

	if !validator.IsValidPort(req.Port) {
		return SubmitApplicationResponse{}, errors.BadRequest("invalid port")
	}

	var existingApp models.Application
	if err := s.db.Where("name = ?", req.Name).First(&existingApp).Error; err == nil {
		return SubmitApplicationResponse{}, errors.Conflict("application name already exists")
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
		return SubmitApplicationResponse{}, errors.Internal("failed to submit application")
	}

	s.notificationService.CreateAdminNotification(fmt.Sprintf("New application submitted: %s", req.Name))

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
	CreatedAt       string   `json:"created_at"`
	OwnerID         uint     `json:"owner_id"`
	Status          string   `json:"status"`
	PrimaryHostname string   `json:"primary_hostname"`
	ExtraHostnames  []string `json:"extra_hostnames"`
}

func (s *ApplicationService) GetApplication(userId uint, appId uint) (GetApplicationResponse, errors.CustomError) {
	var application models.Application
	if err := s.db.Preload("ExtraHostnames").First(&application, appId).Error; err != nil {
		return GetApplicationResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return GetApplicationResponse{}, errors.Forbidden("permission denied")
	}

	return GetApplicationResponse{
		ID:              application.ID,
		Name:            application.Name,
		GitURL:          application.GitURL,
		Branch:          application.Branch,
		Port:            application.Port,
		Description:     application.Description,
		CreatedAt:       application.CreatedAt.Format("2006-01-02 15:04:05"),
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

func (s *ApplicationService) DeleteApplication(userId uint, appId uint) (DeleteApplicationResponse, errors.CustomError) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return DeleteApplicationResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return DeleteApplicationResponse{}, errors.Forbidden("permission denied")
	}

	if application.Status == models.ApplicationStatusApproved {
		if kubernetes.NamespaceExists(application.Name) {
			if err := kubernetes.DeleteNamespace(application.Name); err != nil {
				return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to delete namespace: %v", err))
			}
		}

		if exists, err := harbor.RepositoryExists(application.Name); err != nil {
			return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to check Harbor repository: %v", err))
		} else if exists {
			if err := harbor.DeleteRepository(application.Name); err != nil {
				return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to delete Harbor repository: %v", err))
			}
		}

		if err := vault.DeleteSecret(application.Name); err != nil {
			return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to delete secret: %v", err))
		}

		if err := database.DeleteDatabaseAndUser(application.Name); err != nil {
			return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to delete database and user: %v", err))
		}

		if err := github.TriggerRemovePipelineWorkflow(application); err != nil {
			return DeleteApplicationResponse{}, errors.Internal(fmt.Sprintf("failed to trigger GitHub workflow: %v", err))
		}
	}

	if err := s.db.Delete(&application).Error; err != nil {
		return DeleteApplicationResponse{}, errors.Internal("failed to delete application")
	}

	s.notificationService.CreateAdminNotification(fmt.Sprintf("Application deleted: %s", application.Name))

	return DeleteApplicationResponse{
		Message: "Application deleted successfully",
	}, nil
}

type AddExtralHostnameRequest struct {
	Hostname string `json:"hostname" binding:"required"`
}

type AddExtralHostnameResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) AddExtralHostname(userId uint, appId uint, req AddExtralHostnameRequest) (AddExtralHostnameResponse, errors.CustomError) {
	if !validator.IsValidHostname(req.Hostname) {
		return AddExtralHostnameResponse{}, errors.BadRequest("invalid hostname")
	}

	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return AddExtralHostnameResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return AddExtralHostnameResponse{}, errors.Forbidden("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return AddExtralHostnameResponse{}, errors.BadRequest("application not approved")
	}

	if application.PrimaryHostname == req.Hostname {
		return AddExtralHostnameResponse{}, errors.BadRequest("hostname already exists as primary hostname")
	}

	var existingHostname models.ExtraHostnames
	err := s.db.Unscoped().Where("hostname = ?", req.Hostname).First(&existingHostname).Error

	if err == nil {
		if !existingHostname.DeletedAt.Valid {
			return AddExtralHostnameResponse{}, errors.Conflict("hostname already exists")
		}
		if err := s.db.Unscoped().Model(&existingHostname).Update("deleted_at", nil).Error; err != nil {
			return AddExtralHostnameResponse{}, errors.Internal("failed to restore soft deleted hostname")
		}
		if err := s.db.Model(&existingHostname).Update("application_id", application.ID).Error; err != nil {
			return AddExtralHostnameResponse{}, errors.Internal("failed to update hostname")
		}
	} else if err == gorm.ErrRecordNotFound {
		newHostname := models.ExtraHostnames{
			ApplicationID: application.ID,
			Hostname:      req.Hostname,
		}

		if err := s.db.Create(&newHostname).Error; err != nil {
			return AddExtralHostnameResponse{}, errors.Internal("failed to add extra hostname")
		}
	} else {
		return AddExtralHostnameResponse{}, errors.Internal("failed to check hostname existence")
	}

	if err := github.TriggerAddExtraHostnameWorkflow(application, req.Hostname); err != nil {
		return AddExtralHostnameResponse{}, errors.Internal(fmt.Sprintf("failed to trigger GitHub workflow: %v", err))
	}

	return AddExtralHostnameResponse{
		Message: "Additional hostname added successfully",
	}, nil
}

type DeleteAdditionalHostnameRequest struct {
	Hostname string `json:"hostname" binding:"required"`
}

type DeleteAdditionalHostnameResponse struct {
	Message string `json:"message"`
}

func (s *ApplicationService) DeleteExtraHostname(userId uint, appId uint, req DeleteAdditionalHostnameRequest) (DeleteAdditionalHostnameResponse, errors.CustomError) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return DeleteAdditionalHostnameResponse{}, errors.Forbidden("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return DeleteAdditionalHostnameResponse{}, errors.BadRequest("application not approved")
	}

	if application.PrimaryHostname == req.Hostname {
		return DeleteAdditionalHostnameResponse{}, errors.BadRequest("cannot delete primary hostname")
	}

	var hostname models.ExtraHostnames
	if err := s.db.Where("application_id = ? AND hostname = ?", application.ID, req.Hostname).First(&hostname).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.NotFound("hostname not found")
	}

	if err := s.db.Delete(&hostname).Error; err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.Internal("failed to delete extra hostname")
	}

	if err := github.TriggerDeleteExtraHostnameWorkflow(application, req.Hostname); err != nil {
		return DeleteAdditionalHostnameResponse{}, errors.Internal(fmt.Sprintf("failed to trigger GitHub workflow: %v", err))
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

func (s *ApplicationService) GetEnvironments(userId uint, appId uint) (GetEnvironmentsResponse, errors.CustomError) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return GetEnvironmentsResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return GetEnvironmentsResponse{}, errors.Forbidden("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return GetEnvironmentsResponse{}, errors.BadRequest("application not approved")
	}

	secret, err := vault.GetSecret(application.Name)
	if err != nil {
		return GetEnvironmentsResponse{}, errors.Internal(fmt.Sprintf("failed to read from Vault: %v", err))
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

func (s *ApplicationService) UpdateEnvironment(userId uint, appId uint, req UpdateEnvironmentRequest) (UpdateEnvironmentResponse, errors.CustomError) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return UpdateEnvironmentResponse{}, errors.NotFound("application not found")
	}

	if application.OwnerID != userId {
		return UpdateEnvironmentResponse{}, errors.Forbidden("permission denied")
	}

	if application.Status != models.ApplicationStatusApproved {
		return UpdateEnvironmentResponse{}, errors.BadRequest("application not approved")
	}

	data := make(map[string]interface{})
	for _, env := range req.Environments {
		data[env.Key] = env.Value
	}

	if err := vault.UpdateSecret(application.Name, data); err != nil {
		return UpdateEnvironmentResponse{}, errors.Internal(fmt.Sprintf("failed to write to Vault: %v", err))
	}

	return UpdateEnvironmentResponse{
		Message: "Environment updated successfully",
	}, nil
}
