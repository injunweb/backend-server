package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/websocket"

	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

func (s *NotificationService) CreateNotification(userID uint, message string) error {
	notification := models.Notification{
		UserID:  userID,
		Message: message,
	}
	err := s.db.Create(&notification).Error
	if err != nil {
		return errors.New("failed to create notification")
	}
	websocket.SendNotification(userID, message)
	return nil
}

func (s *NotificationService) CreateAdminNotification(message string) error {
	var users []models.User
	err := s.db.Where("is_admin = ?", true).Find(&users).Error
	if err != nil {
		return errors.New("failed to retrieve users")
	}

	for _, user := range users {
		notification := models.Notification{
			UserID:  user.ID,
			Message: message,
		}
		err := s.db.Create(&notification).Error
		if err != nil {
			return errors.New("failed to create notification")
		}
		websocket.SendNotification(user.ID, message)
	}
	return nil
}

type GetUserNotificationsResponse struct {
	Notifications []struct {
		ID      uint   `json:"id"`
		Message string `json:"message"`
		IsRead  bool   `json:"is_read"`
	} `json:"notifications"`
}

func (s *NotificationService) GetUserNotifications(userId uint) (GetUserNotificationsResponse, error) {
	var notifications []models.Notification
	err := s.db.
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return GetUserNotificationsResponse{}, errors.New("failed to retrieve notifications")
	}

	var response GetUserNotificationsResponse
	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, struct {
			ID      uint   `json:"id"`
			Message string `json:"message"`
			IsRead  bool   `json:"is_read"`
		}{
			ID:      notification.ID,
			Message: notification.Message,
			IsRead:  notification.IsRead,
		})
	}
	return response, nil
}

type MarkAsReadResponse struct {
	Message string `json:"message"`
}

func (s *NotificationService) MarkAsRead(userId uint, notificationId uint) (MarkAsReadResponse, error) {
	result := s.db.
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationId, userId).
		Updates(map[string]interface{}{"is_read": true})

	if result.Error != nil {
		return MarkAsReadResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return MarkAsReadResponse{}, errors.New("notification not found or already read")
	}

	return MarkAsReadResponse{
		Message: "Notification marked as read successfully",
	}, nil
}
