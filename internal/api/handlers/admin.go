package handlers

import (
	"net/http"
	"strconv"

	"github.com/injunweb/backend-server/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) GetUsersByAdmin(c *gin.Context) {
	response, err := h.adminService.GetUsersByAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) GetUserByAdmin(c *gin.Context) {
	userId, _ := strconv.ParseUint(c.Param("userId"), 10, 32)

	response, err := h.adminService.GetUserByAdmin(uint(userId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) GetApplicationsByAdmin(c *gin.Context) {
	userId, _ := strconv.ParseUint(c.Param("userId"), 10, 32)

	response, err := h.adminService.GetApplicationsByAdmin(uint(userId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) ApproveApplicationByAdmin(c *gin.Context) {
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.adminService.ApproveApplicationByAdmin(uint(appId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) CancleApproveApplicationByAdmin(c *gin.Context) {
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.adminService.CancleApproveApplicationByAdmin(uint(appId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) UpdatePrimaryHostnameByAdmin(c *gin.Context) {
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	var req services.UpdateCustomHostnameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.adminService.UpdatePrimaryHostnameByAdmin(uint(appId), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AdminHandler) GetApplicationByAdmin(c *gin.Context) {
	appId, _ := strconv.ParseUint(c.Param("appId"), 10, 32)

	response, err := h.adminService.GetApplicationByAdmin(uint(appId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
