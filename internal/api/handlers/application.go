package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/errors"
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
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) SubmitApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")

	var request services.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.applicationService.SubmitApplication(userId.(uint), request)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.applicationService.GetApplication(userId.(uint), uint(appId))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.applicationService.DeleteApplication(userId.(uint), uint(appId))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) AddExtralHostname(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	var request services.AddExtralHostnameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.applicationService.AddExtralHostname(userId.(uint), uint(appId), request)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ApplicationHandler) DeleteExtraHostname(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	var request services.DeleteAdditionalHostnameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.applicationService.DeleteExtraHostname(userId.(uint), uint(appId), request)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) GetEnvironments(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.applicationService.GetEnvironments(userId.(uint), uint(appId))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplicationHandler) UpdateEnvironment(c *gin.Context) {
	userId, _ := c.Get("user_id")
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	var request services.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Error(errors.BadRequest("invalid request format"))
		return
	}

	response, err := h.applicationService.UpdateEnvironment(userId.(uint), uint(appId), request)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response)
}
