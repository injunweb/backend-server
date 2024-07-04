package routes

import (
	"github.com/gin-gonic/gin"
	adminControllers "github.com/injunweb/backend-server/internal/domain/admin/controllers"
	applicationControllers "github.com/injunweb/backend-server/internal/domain/application/controllers"
	projectControllers "github.com/injunweb/backend-server/internal/domain/project/controllers"
	"github.com/injunweb/backend-server/internal/global/middlewares"
	"github.com/injunweb/backend-server/internal/global/security"
	"github.com/injunweb/backend-server/internal/interfaces/handlers"
)

type Controllers struct {
	ApplicationController *applicationControllers.ApplicationController
	ProjectController     *projectControllers.ProjectController
	AdminController       *adminControllers.AdminController
}

func NewRouter(controllers Controllers) *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.LoggerMiddleware())

	registerApplicationRoutes(router, controllers.ApplicationController)
	registerProjectRoutes(router, controllers.ProjectController)
	registerAdminRoutes(router, controllers.AdminController)

	return router
}

func registerApplicationRoutes(router *gin.Engine, applicationController *applicationControllers.ApplicationController) {
	applications := router.Group("/applications")
	{
		applications.POST("/", handlers.WrapHandler(applicationController.CreateApplication))
		applications.GET("/check", handlers.WrapHandler(applicationController.CheckProjectName))
	}
}

func registerProjectRoutes(router *gin.Engine, projectController *projectControllers.ProjectController) {
	projects := router.Group("/projects")
	{
		projects.GET("/validate", handlers.WrapHandler(projectController.ValidateKey))
		projects.GET("/dispatch", handlers.WrapHandler(projectController.DispatchProject))
	}
}

func registerAdminRoutes(router *gin.Engine, adminController *adminControllers.AdminController) {
	router.POST("/admin/login", handlers.WrapHandler(adminController.Login))

	admin := router.Group("/admin")
	admin.Use(middlewares.AuthMiddleware(security.RoleAdmin))
	{
		admin.GET("/applications", handlers.WrapHandler(adminController.GetApplications))
		admin.GET("/applications/:project_name", handlers.WrapHandler(adminController.GetApplication))
		admin.GET("/projects", handlers.WrapHandler(adminController.GetProjects))
		admin.GET("/projects/:project_name", handlers.WrapHandler(adminController.GetProject))
		admin.POST("/applications/:project_name/approve", handlers.WrapHandler(adminController.ApproveApplication))
		admin.POST("/applications/:project_name/reject", handlers.WrapHandler(adminController.RejectApplication))
	}
}
