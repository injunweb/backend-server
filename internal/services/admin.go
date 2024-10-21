package services

import (
	"errors"
	"fmt"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/database"
	"github.com/injunweb/backend-server/pkg/email"
	"github.com/injunweb/backend-server/pkg/github"
	"github.com/injunweb/backend-server/pkg/harbor"
	"github.com/injunweb/backend-server/pkg/kubernetes"
	"github.com/injunweb/backend-server/pkg/vault"

	"gorm.io/gorm"
)

type AdminService struct {
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewAdminService(db *gorm.DB, notificationService *NotificationService) *AdminService {
	return &AdminService{db: db, notificationService: notificationService}
}

type GetUsersByAdminResponse struct {
	Users []struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"users"`
}

func (s *AdminService) GetUsersByAdmin() (GetUsersByAdminResponse, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return GetUsersByAdminResponse{}, errors.New("failed to retrieve users")
	}

	var response GetUsersByAdminResponse
	for _, user := range users {
		response.Users = append(response.Users, struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return response, nil
}

type GetUserByAdminResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

func (s *AdminService) GetUserByAdmin(userId uint) (GetUserByAdminResponse, error) {
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return GetUserByAdminResponse{}, errors.New("user not found")
	}

	return GetUserByAdminResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}, nil
}

type GetApplicationsByAdminResponse struct {
	Applications []struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"applications"`
}

func (s *AdminService) GetApplicationsByAdmin(userId uint) (GetApplicationsByAdminResponse, error) {
	var applications []models.Application
	if err := s.db.Where("owner_id = ?", userId).Find(&applications).Error; err != nil {
		return GetApplicationsByAdminResponse{}, errors.New("failed to retrieve applications")
	}

	var response GetApplicationsByAdminResponse
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

type GetAllApplicationsByAdminResponse struct {
	Applications []struct {
		ID            uint   `json:"id"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		OwnerUsername string `json:"owner_username"`
		GitURL        string `json:"git_url"`
		CreatedAt     string `json:"created_at"`
	} `json:"applications"`
}

func (s *AdminService) GetAllApplicationsByAdmin() (GetAllApplicationsByAdminResponse, error) {
	var applications []models.Application
	if err := s.db.Preload("Owner").Order("created_at DESC").Find(&applications).Error; err != nil {
		return GetAllApplicationsByAdminResponse{}, errors.New("failed to retrieve applications")
	}

	var response GetAllApplicationsByAdminResponse
	for _, app := range applications {
		response.Applications = append(response.Applications, struct {
			ID            uint   `json:"id"`
			Name          string `json:"name"`
			Status        string `json:"status"`
			OwnerUsername string `json:"owner_username"`
			GitURL        string `json:"git_url"`
			CreatedAt     string `json:"created_at"`
		}{
			ID:            app.ID,
			Name:          app.Name,
			Status:        app.Status,
			OwnerUsername: app.Owner.Username,
			GitURL:        app.GitURL,
			CreatedAt:     app.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

type ApproveApplicationByAdminResponse struct {
	Message string `json:"message"`
}

func (s *AdminService) ApproveApplicationByAdmin(appId uint) (ApproveApplicationByAdminResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return ApproveApplicationByAdminResponse{}, errors.New("application not found")
	}

	if application.Status != models.ApplicationStatusPending {
		return ApproveApplicationByAdminResponse{}, errors.New("application already approved")
	}

	if err := vault.InitSecret(application.Name, map[string]interface{}{"PORT": fmt.Sprintf("%d", application.Port)}); err != nil {
		return ApproveApplicationByAdminResponse{}, fmt.Errorf("failed to initialize Vault secret: %v", err)
	}

	if err := github.TriggerWriteValuesWorkflow(application); err != nil {
		return ApproveApplicationByAdminResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
	}

	password, err := database.CreateDatabaseAndUser(application.Name)
	if err != nil {
		return ApproveApplicationByAdminResponse{}, fmt.Errorf("failed to create database and user: %v", err)
	}

	var owner models.User
	if err := s.db.First(&owner, application.OwnerID).Error; err != nil {
		return ApproveApplicationByAdminResponse{}, errors.New("failed to find user email")
	}

	if err := email.SendApprovalEmail(owner.Email, application.Name, password); err != nil {
		return ApproveApplicationByAdminResponse{}, fmt.Errorf("failed to send email: %v", err)
	}

	application.Status = models.ApplicationStatusApproved
	if err := s.db.Save(&application).Error; err != nil {
		return ApproveApplicationByAdminResponse{}, errors.New("failed to update application status")
	}

	s.notificationService.CreateNotification(application.OwnerID, fmt.Sprintf("Your application %s has been approved", application.Name))

	return ApproveApplicationByAdminResponse{
		Message: "Application approved successfully",
	}, nil
}

type CancelApproveApplicationByAdminResponse struct {
	Message string `json:"message"`
}

func (s *AdminService) CancelApproveApplicationByAdmin(appId uint) (CancelApproveApplicationByAdminResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return CancelApproveApplicationByAdminResponse{}, errors.New("application not found")
	}

	if application.Status != models.ApplicationStatusApproved {
		return CancelApproveApplicationByAdminResponse{}, errors.New("application not approved")
	}

	if kubernetes.NamespaceExists(application.Name) {
		if err := kubernetes.DeleteNamespace(application.Name); err != nil {
			return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to delete namespace: %v", err)
		}
	}

	if exists, err := harbor.RepositoryExists(application.Name); err != nil {
		return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to check Harbor repository: %v", err)
	} else if exists {
		if err := harbor.DeleteRepository(application.Name); err != nil {
			return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to delete Harbor repository: %v", err)
		}
	}

	if err := vault.DeleteSecret(application.Name); err != nil {
		return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to delete secret: %v", err)
	}

	if err := database.DeleteDatabaseAndUser(application.Name); err != nil {
		return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to delete database and user: %v", err)
	}

	if err := github.TriggerRemovePipelineWorkflow(application); err != nil {
		return CancelApproveApplicationByAdminResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
	}

	application.Status = models.ApplicationStatusPending
	if err := s.db.Save(&application).Error; err != nil {
		return CancelApproveApplicationByAdminResponse{}, errors.New("failed to update application status")
	}

	return CancelApproveApplicationByAdminResponse{
		Message: "Application approval canceled successfully",
	}, nil
}

type UpdateCustomHostnameRequest struct {
	Hostname string `json:"hostname" binding:"required"`
}

type UpdateCustomHostnameByAdminResponse struct {
	Message string `json:"message"`
}

func (s *AdminService) UpdatePrimaryHostnameByAdmin(appId uint, request UpdateCustomHostnameRequest) (UpdateCustomHostnameByAdminResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return UpdateCustomHostnameByAdminResponse{}, errors.New("application not found")
	}

	if application.Status != models.ApplicationStatusApproved {
		return UpdateCustomHostnameByAdminResponse{}, errors.New("application not approved")
	}

	if err := github.TriggerUpdateCustomHostname(application, request.Hostname); err != nil {
		return UpdateCustomHostnameByAdminResponse{}, fmt.Errorf("failed to trigger GitHub workflow: %v", err)
	}

	return UpdateCustomHostnameByAdminResponse{
		Message: "Custom hostname updated successfully",
	}, nil
}

type GetApplicationByAdminResponse struct {
	ID              uint     `json:"id"`
	Name            string   `json:"name"`
	GitURL          string   `json:"git_url"`
	Branch          string   `json:"branch"`
	Port            int      `json:"port"`
	Description     string   `json:"description"`
	OwnerID         uint     `json:"owner_id"`
	Status          string   `json:"status"`
	OwnerUsername   string   `json:"owner_username"`
	PrimaryHostname string   `json:"primary_hostname"`
	ExtraHostnames  []string `json:"extra_hostnames"`
	CreationDate    string   `json:"creation_date"`
}

func (s *AdminService) GetApplicationByAdmin(appId uint) (GetApplicationByAdminResponse, error) {
	var application models.Application
	if err := s.db.Preload("Owner").Preload("ExtraHostnames").First(&application, appId).Error; err != nil {
		return GetApplicationByAdminResponse{}, errors.New("application not found")
	}

	return GetApplicationByAdminResponse{
		ID:              application.ID,
		Name:            application.Name,
		GitURL:          application.GitURL,
		Branch:          application.Branch,
		Port:            application.Port,
		Description:     application.Description,
		OwnerID:         application.OwnerID,
		Status:          application.Status,
		OwnerUsername:   application.Owner.Username,
		PrimaryHostname: application.PrimaryHostname,
		ExtraHostnames: func() []string {
			var extraHostnames []string
			for _, hostname := range application.ExtraHostnames {
				extraHostnames = append(extraHostnames, hostname.Hostname)
			}
			return extraHostnames
		}(),
		CreationDate: application.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
