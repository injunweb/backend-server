package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/errors"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userId, _ := c.Get("user_id")

	response, err := h.userService.GetUser(userId.(uint))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var updateRequest services.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.userService.UpdateUser(userId.(uint), updateRequest)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}
