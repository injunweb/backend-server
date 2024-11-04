package services

import (
	"fmt"
	"time"

	"github.com/injunweb/backend-server/internal/config"
	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/errors"
	"github.com/injunweb/backend-server/pkg/validator"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewAuthService(db *gorm.DB, notificationService *NotificationService) *AuthService {
	return &AuthService{db: db, notificationService: notificationService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

func (s *AuthService) Login(req LoginRequest) (LoginResponse, errors.CustomError) {
	var user models.User
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", req.Username).First(&user).Error; err != nil {
			return errors.Unauthorized("invalid credentials")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return errors.Unauthorized("invalid credentials")
		}

		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return LoginResponse{}, customErr
		}
		return LoginResponse{}, errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return LoginResponse{}, errors.Internal("failed to generate token")
	}

	return LoginResponse{
		Token:   tokenString,
		Message: "Login successful",
	}, nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

func (s *AuthService) Register(req RegisterRequest) (RegisterResponse, errors.CustomError) {
	if !validator.IsValidUsername(req.Username) {
		return RegisterResponse{}, errors.BadRequest("invalid username")
	}

	if !validator.IsValidEmail(req.Email) {
		return RegisterResponse{}, errors.BadRequest("invalid email")
	}

	if !validator.IsValidPassword(req.Password) {
		return RegisterResponse{}, errors.BadRequest("invalid password")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existingUser models.User
		if err := tx.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
			return errors.Conflict("username already exists")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Internal("failed to hash password")
		}

		user := models.User{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
			IsAdmin:  false,
		}

		if err := tx.Create(&user).Error; err != nil {
			return errors.Internal("failed to register user")
		}

		s.notificationService.CreateAdminNotification("New user registered: " + user.Username)
		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return RegisterResponse{}, customErr
		}
		return RegisterResponse{}, errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	return RegisterResponse{
		Message: "User registered successfully",
	}, nil
}
