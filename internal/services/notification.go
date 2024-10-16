package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/webpush"

	"gorm.io/gorm"
)

type NotificationService struct {
	db          *gorm.DB
	userService *UserService
}

func NewNotificationService(db *gorm.DB, userService *UserService) *NotificationService {
	return &NotificationService{db: db, userService: userService}
}

func (s *NotificationService) CreateNotification(userID uint, message string) error {
	notification := models.Notification{
		UserID:  userID,
		Message: message,
	}
	if err := s.db.Create(&notification).Error; err != nil {
		return errors.New("failed to create notification")
	}

	subscriptions, err := s.userService.GetUserSubscriptions(userID)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		subscription := webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: struct {
				P256dh string `json:"p256dh"`
				Auth   string `json:"auth"`
			}{
				P256dh: sub.P256dh,
				Auth:   sub.Auth,
			},
		}
		if err := webpush.SendNotification(subscription, message); err != nil {
			return errors.New("failed to send notification")
		}
	}

	return nil
}

func (s *NotificationService) CreateAdminNotification(message string) error {
	var users []models.User
	err := s.db.Where("is_admin = ?", true).Find(&users).Error
	if err != nil {
		return errors.New("failed to retrieve users")
	}

	for _, user := range users {
		if err := s.CreateNotification(user.ID, message); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotificationService) GetUserNotifications(userId uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return nil, errors.New("failed to retrieve notifications")
	}

	return notifications, nil
}

func (s *NotificationService) MarkAsRead(userId uint, notificationId uint) error {
	result := s.db.
		Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationId, userId).
		Updates(map[string]interface{}{"is_read": true})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found or already read")
	}

	return nil
}
