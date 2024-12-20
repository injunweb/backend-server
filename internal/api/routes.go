package api

import (
	"github.com/injunweb/backend-server/internal/api/handlers"
	"github.com/injunweb/backend-server/internal/middleware"
	"github.com/injunweb/backend-server/internal/services"
	"github.com/injunweb/backend-server/pkg/database"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	userService := services.NewUserService(database.DB)
	notificationService := services.NewNotificationService(database.DB, userService)
	authService := services.NewAuthService(database.DB, notificationService)
	appService := services.NewApplicationService(database.DB, notificationService)
	adminService := services.NewAdminService(database.DB, notificationService)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	appHandler := handlers.NewApplicationHandler(appService)
	notificationHandler := handlers.NewNotificationHandler(notificationService, userService)
	adminHandler := handlers.NewAdminHandler(adminService)

	router.Use(middleware.ErrorMiddleware())

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

	notifications := router.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		notifications.GET("", notificationHandler.GetNotifications)
		notifications.POST("/read", notificationHandler.MarkAllAsRead)
		notifications.DELETE("/:notificationId", notificationHandler.DeleteNotification)
		notifications.POST("/subscribe", notificationHandler.Subscribe)
		notifications.GET("/vapid-public-key", notificationHandler.GetVAPIDPublicKey)
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
			adminApplications.POST("/:appId/cancel-approve", adminHandler.CancelApproveApplicationByAdmin)
			adminApplications.POST("/:appId/primary-hostname", adminHandler.UpdatePrimaryHostnameByAdmin)
			adminApplications.GET("/:appId", adminHandler.GetApplicationByAdmin)
		}
	}
}
