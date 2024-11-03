package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/errors"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var loginRequest services.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.authService.Login(loginRequest)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var registerRequest services.RegisterRequest
	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.authService.Register(registerRequest)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response)
}
