package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/validator"

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

	if !validator.IsValidEmail(req.Email) {
		return UpdateUserResponse{}, errors.New("invalid email")
	}
	if !validator.IsValidUsername(req.Username) {
		return UpdateUserResponse{}, errors.New("invalid username")
	}

	if err := s.db.Save(&user).Error; err != nil {
		return UpdateUserResponse{}, errors.New("failed to update user")
	}

	return UpdateUserResponse{
		Message: "User updated successfully",
	}, nil
}

func (s *UserService) AddSubscription(userID uint, endpoint, p256dh, auth string) error {
	subscription := models.Subscription{
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}

	result := s.db.Create(&subscription)
	if result.Error != nil {
		return errors.New("failed to add subscription")
	}
	return nil
}

func (s *UserService) GetUserSubscriptions(userID uint) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	if err := s.db.Where("user_id = ?", userID).Find(&subscriptions).Error; err != nil {
		return nil, errors.New("failed to retrieve subscriptions")
	}
	return subscriptions, nil
}
