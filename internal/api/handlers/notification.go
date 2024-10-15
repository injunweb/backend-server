package handlers

import (
	"net/http"
	"strconv"

	"github.com/injunweb/backend-server/internal/services"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userId, _ := c.Get("user_id")

	response, err := h.notificationService.GetUserNotifications(userId.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userId, _ := c.Get("user_id")
	notificationId, err := strconv.ParseUint(c.Param("notificationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	response, err := h.notificationService.MarkAsRead(userId.(uint), uint(notificationId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
