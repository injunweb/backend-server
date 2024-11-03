package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/errors"
	"github.com/injunweb/backend-server/pkg/webpush"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
	userService         *services.UserService
}

func NewNotificationHandler(notificationService *services.NotificationService, userService *services.UserService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		userService:         userService,
	}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userId, _ := c.Get("user_id")

	notifications, err := h.notificationService.GetUserNotifications(userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userId, _ := c.Get("user_id")

	err := h.notificationService.MarkAllAsRead(userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read successfully"})
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userId, _ := c.Get("user_id")
	notificationId, err := strconv.ParseUint(c.Param("notificationId"), 10, 32)
	if err != nil {
		c.Error(errors.BadRequest("invalid notification ID"))
		return
	}

	err = h.notificationService.DeleteNotification(userId.(uint), uint(notificationId))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

func (h *NotificationHandler) Subscribe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var subscription webpush.Subscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.Error(errors.BadRequest("invalid subscription data"))
		return
	}

	err := h.userService.AddSubscription(
		userID.(uint),
		subscription.Endpoint,
		subscription.Keys.P256dh,
		subscription.Keys.Auth,
	)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription saved successfully"})
}

func (h *NotificationHandler) GetVAPIDPublicKey(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"vapidPublicKey": webpush.GetVAPIDPublicKey()})
}
