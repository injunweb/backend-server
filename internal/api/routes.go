package api

import (
	"github.com/injunweb/backend-server/internal/api/handlers"
	"github.com/injunweb/backend-server/internal/middleware"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/database"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	authService := services.NewAuthService(database.DB)
	userService := services.NewUserService(database.DB)
	appService := services.NewApplicationService(database.DB)
	adminService := services.NewAdminService(database.DB)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	appHandler := handlers.NewApplicationHandler(appService)
	adminHandler := handlers.NewAdminHandler(adminService)

	auth := router.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
	}

	users := router.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("", userHandler.GetUser)
		users.PATCH("", userHandler.UpdateUser)
	}

	applications := router.Group("/applications")
	applications.Use(middleware.AuthMiddleware())
	{
		applications.POST("", appHandler.SubmitApplication)
		applications.GET("", appHandler.GetApplications)
		applications.GET("/:appId", appHandler.GetApplication)
		applications.DELETE("/:appId", appHandler.DeleteApplication)
		applications.POST("/:appId/extra-hostnames", appHandler.AddExtralHostname)
		applications.DELETE("/:appId/extra-hostnames", appHandler.DeleteExtraHostname)

		environments := applications.Group("/:appId/environments")
		{
			environments.GET("", appHandler.GetEnvironments)
			environments.POST("", appHandler.UpdateEnvironment)
		}
	}

	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("", adminHandler.GetUsersByAdmin)
			adminUsers.GET("/:userId", adminHandler.GetUserByAdmin)

			adminApplications := adminUsers.Group("/:userId/applications")
			{
				adminApplications.GET("", adminHandler.GetApplicationsByAdmin)
			}
		}

		adminApplications := admin.Group("/applications")
		{
			adminApplications.GET("", adminHandler.GetAllApplicationsByAdmin)
			adminApplications.POST("/:appId/approve", adminHandler.ApproveApplicationByAdmin)
			adminApplications.POST("/:appId/cancle-approve", adminHandler.CancleApproveApplicationByAdmin)
			adminApplications.POST("/:appId/primary-hostname", adminHandler.UpdatePrimaryHostnameByAdmin)
			adminApplications.GET("/:appId", adminHandler.GetApplicationByAdmin)
		}
	}
}
