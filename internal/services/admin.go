package services

import (
	"errors"
	"fmt"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/database"
	"github.com/injunweb/backend-server/pkg/email"
	"github.com/injunweb/backend-server/pkg/github"
	"github.com/injunweb/backend-server/pkg/vault"

	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
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

	if err := vault.InitSecret(application.Name, map[string]interface{}{"hostname": application.Name + ".injunweb.com"}); err != nil {
		return ApproveApplicationByAdminResponse{}, fmt.Errorf("failed to initialize Vault secret: %v", err)
	}

	if err := github.TriggerWorkflow(application); err != nil {
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

	return ApproveApplicationByAdminResponse{
		Message: "Application approved successfully",
	}, nil
}

type GetApplicationByAdminResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	GitURL      string `json:"git_url"`
	Branch      string `json:"branch"`
	Port        int    `json:"port"`
	Description string `json:"description"`
	OwnerID     uint   `json:"owner_id"`
	Status      string `json:"status"`
}

func (s *AdminService) GetApplicationByAdmin(appId uint) (GetApplicationByAdminResponse, error) {
	var application models.Application
	if err := s.db.First(&application, appId).Error; err != nil {
		return GetApplicationByAdminResponse{}, errors.New("application not found")
	}

	return GetApplicationByAdminResponse{
		ID:          application.ID,
		Name:        application.Name,
		GitURL:      application.GitURL,
		Branch:      application.Branch,
		Port:        application.Port,
		Description: application.Description,
		OwnerID:     application.OwnerID,
		Status:      application.Status,
	}, nil
}
