package server

import (
	"log"

	adminControllers "github.com/injunweb/backend-server/internal/domain/admin/controllers"
	adminRepositories "github.com/injunweb/backend-server/internal/domain/admin/repositories"
	adminServices "github.com/injunweb/backend-server/internal/domain/admin/services"
	applicationControllers "github.com/injunweb/backend-server/internal/domain/application/controllers"
	applicationEntities "github.com/injunweb/backend-server/internal/domain/application/entities"
	applicationRepositories "github.com/injunweb/backend-server/internal/domain/application/repositories"
	applicationServices "github.com/injunweb/backend-server/internal/domain/application/services"
	projectControllers "github.com/injunweb/backend-server/internal/domain/project/controllers"
	projectEntities "github.com/injunweb/backend-server/internal/domain/project/entities"
	projectRepositories "github.com/injunweb/backend-server/internal/domain/project/repositories"
	projectServices "github.com/injunweb/backend-server/internal/domain/project/services"
	"github.com/injunweb/backend-server/internal/global/database"
	applicationRoutes "github.com/injunweb/backend-server/internal/interfaces/routes"
)

func Run() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	db.AutoMigrate(&applicationEntities.Application{}, &projectEntities.Project{})

	applicationRepository := applicationRepositories.NewApplicationRepository(db)
	applicationService := applicationServices.NewApplicationService(applicationRepository)
	applicationController := applicationControllers.NewApplicationController(applicationService)

	projectRepository := projectRepositories.NewProjectRepository(db)
	projectService := projectServices.NewProjectService(projectRepository)
	projectController := projectControllers.NewProjectController(projectService)

	adminRepository := adminRepositories.NewAdminRepository(db)
	adminService := adminServices.NewAdminService(adminRepository, applicationRepository, projectRepository)
	adminController := adminControllers.NewAdminController(adminService)

	controllers := applicationRoutes.Controllers{
		ApplicationController: applicationController,
		ProjectController:     projectController,
		AdminController:       adminController,
	}

	router := applicationRoutes.NewRouter(controllers)
	router.Run(":8080")
}
