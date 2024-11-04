package services

import (
	"fmt"

	"github.com/injunweb/backend-server/internal/models"
	"github.com/injunweb/backend-server/pkg/errors"
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

func (s *NotificationService) CreateNotification(userID uint, message string) errors.CustomError {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		notification := models.Notification{
			UserID:  userID,
			Message: message,
		}
		if err := tx.Create(&notification).Error; err != nil {
			return errors.Internal("failed to create notification")
		}

		subscriptions, err := s.userService.GetUserSubscriptions(userID)
		if err != nil {
			return errors.Internal("failed to get user subscriptions")
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
			webpush.SendNotification(subscription, message)
		}

		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return customErr
		}
		return errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	return nil
}

func (s *NotificationService) CreateAdminNotification(message string) errors.CustomError {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var users []models.User
		if err := tx.Where("is_admin = ?", true).Find(&users).Error; err != nil {
			return errors.Internal("failed to retrieve users")
		}

		for _, user := range users {
			if err := s.CreateNotification(user.ID, message); err != nil {
				return errors.Internal("failed to create admin notification")
			}
		}
		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return customErr
		}
		return errors.Internal(fmt.Sprintf("transaction failed: %v", err))
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

func (s *NotificationService) GetUserNotifications(userId uint) (GetUsersNotificationsResponse, errors.CustomError) {
	var notifications []models.Notification
	var response GetUsersNotificationsResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("user_id = ?", userId).
			Order("created_at DESC").
			Find(&notifications).Error

		if err != nil {
			return errors.Internal("failed to retrieve notifications")
		}

		response.UnreadCount = 0
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
		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return GetUsersNotificationsResponse{}, customErr
		}
		return GetUsersNotificationsResponse{}, errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	return response, nil
}

func (s *NotificationService) MarkAllAsRead(userId uint) errors.CustomError {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Notification{}).
			Where("user_id = ? AND is_read = ?", userId, false).
			Updates(map[string]interface{}{"is_read": true})

		if result.Error != nil {
			return errors.Internal("failed to mark notifications as read")
		}
		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return customErr
		}
		return errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	return nil
}

func (s *NotificationService) DeleteNotification(userId uint, notificationId uint) errors.CustomError {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND user_id = ?", notificationId, userId).
			Delete(&models.Notification{})

		if result.Error != nil {
			return errors.Internal("failed to delete notification")
		}
		if result.RowsAffected == 0 {
			return errors.NotFound("notification not found")
		}
		return nil
	})

	if err != nil {
		if customErr, ok := err.(errors.CustomError); ok {
			return customErr
		}
		return errors.Internal(fmt.Sprintf("transaction failed: %v", err))
	}

	return nil
}
