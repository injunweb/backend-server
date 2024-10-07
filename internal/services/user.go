package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type GetUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
}

func (s *UserService) GetUser(userId uint) (GetUserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return GetUserResponse{}, errors.New("user not found")
	}

	return GetUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}, nil
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
}

func (s *UserService) UpdateUser(userId uint, req UpdateUserRequest) (UpdateUserResponse, error) {
	var user models.User
	if err := s.db.First(&user, userId).Error; err != nil {
		return UpdateUserResponse{}, errors.New("user not found")
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Username != "" {
		user.Username = req.Username
	}

	if err := s.db.Save(&user).Error; err != nil {
		return UpdateUserResponse{}, errors.New("failed to update user")
	}

	return UpdateUserResponse{
		Message: "User updated successfully",
	}, nil
}
