package handlers

import (
	"net/http"
	"strconv"

	"github.com/injunweb/backend-server/internal/services"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	applicationService *services.ApplicationService
}

func NewApplicationHandler(applicationService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{applicationService: applicationService}
}

func (h *ApplicationHandler) GetApplications(c *gin.Context) {
	userId, _ := c.Get("user_id")

	response, err := h.applicationService.GetApplications(userId.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var request services.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.applicationService.SubmitApplication(userId.(uint), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.applicationService.GetApplication(userId.(uint), uint(appId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) GetEnvironments(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.applicationService.GetEnvironments(userId.(uint), uint(appId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) UpdateEnvironment(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	var request services.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	response, err := h.applicationService.UpdateEnvironment(userId.(uint), uint(appId), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
