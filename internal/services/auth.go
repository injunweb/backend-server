package services

import (
	"errors"
	"time"

	"github.com/injunweb/backend-server/internal/config"
	"github.com/injunweb/backend-server/internal/models"

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

func (s *AuthService) Login(req LoginRequest) (LoginResponse, error) {
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return LoginResponse{}, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return LoginResponse{}, errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return LoginResponse{}, errors.New("failed to generate token")
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

func (s *AuthService) Register(req RegisterRequest) (RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, errors.New("failed to hash password")
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		IsAdmin:  false,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return RegisterResponse{}, errors.New("failed to register user")
	}

	s.notificationService.CreateAdminNotification("New user registered: " + user.Username)

	return RegisterResponse{
		Message: "User registered successfully",
	}, nil
}
