package controllers

import (
	"github.com/injunweb/backend-server/internal/domain/project/services"
	"github.com/injunweb/backend-server/internal/interfaces/handlers"
)

type ProjectController struct {
	projectService *services.ProjectService
}

func NewProjectController(projectService *services.ProjectService) *ProjectController {
	return &ProjectController{projectService: projectService}
}

func (projectController *ProjectController) ValidateKey(query handlers.Query) error {
	projectName := query.Get("project_name")
	accessKey := query.Get("access_key")

	return projectController.projectService.ValidateKey(projectName, accessKey)
}

func (projectController *ProjectController) DispatchProject(query handlers.Query) {
	projectName := query.Get("project_name")
	accessKey := query.Get("access_key")
	repo := query.Get("repo")
	event := query.Get("event")

	projectController.projectService.TriggerDispatch(projectName, accessKey, repo, event)
}
