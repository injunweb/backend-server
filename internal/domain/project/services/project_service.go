package services

import (
	"errors"

	"github.com/injunweb/backend-server/internal/domain/project/repositories"
	"github.com/injunweb/backend-server/internal/global/integrations/dispatcher"
)

type ProjectService struct {
	projectRepository *repositories.ProjectRepository
}

func NewProjectService(projectRepository *repositories.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepository: projectRepository}
}

func (projectService *ProjectService) ValidateKey(projectName string, accessKey string) error {
	project, err := projectService.projectRepository.GetProjectByProjectName(projectName)

	if err != nil {
		return errors.New("project not found")
	}

	if project.AccessKey != accessKey {
		return errors.New("invalid access key")
	}

	return nil
}

func (projectService *ProjectService) TriggerDispatch(projectName string, accessKey string, repo string, event string) error {
	err := projectService.ValidateKey(projectName, accessKey)

	if err != nil {
		return err
	}

	project, err := projectService.projectRepository.GetProjectByProjectName(projectName)

	if err != nil {
		return errors.New("project not found")
	}

	clientPayload := dispatcher.ClientPayload{
		"email":                 project.Email,
		"project_name":          project.ProjectName,
		"github_repository_url": project.GithubRepositoryUrl,
	}

	err = dispatcher.TriggerDispatch(repo, event, clientPayload)

	if err != nil {
		return errors.New("failed to trigger dispatch")
	}

	return nil
}
