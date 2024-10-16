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

type GetUsersNotificationsResponse struct {
	Notifications []struct {
		ID        uint   `json:"id"`
		Message   string `json:"message"`
		IsRead    bool   `json:"is_read"`
		CreatedAt string `json:"created_at"`
	} `json:"notifications"`

	UnreadCount int `json:"unread_count"`
}

func (s *NotificationService) GetUserNotifications(userId uint) (GetUsersNotificationsResponse, error) {
	var notifications []models.Notification
	err := s.db.
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Find(&notifications).Error

	if err != nil {
		return GetUsersNotificationsResponse{}, errors.New("failed to retrieve notifications")
	}

	response := GetUsersNotificationsResponse{
		UnreadCount: 0,
	}

	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, struct {
			ID        uint   `json:"id"`
			Message   string `json:"message"`
			IsRead    bool   `json:"is_read"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        notification.ID,
			Message:   notification.Message,
			IsRead:    notification.IsRead,
			CreatedAt: notification.CreatedAt.Format("2006-01-02 15:04:05"),
		})

		if !notification.IsRead {
			response.UnreadCount++
		}
	}

	return response, nil
}

func (s *NotificationService) MarkAllAsRead(userId uint) error {
	result := s.db.
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userId, false).
		Updates(map[string]interface{}{"is_read": true})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *NotificationService) DeleteNotification(userId uint, notificationId uint) error {
	result := s.db.
		Where("id = ? AND user_id = ?", notificationId, userId).
		Delete(&models.Notification{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}
