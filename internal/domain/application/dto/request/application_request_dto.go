package request

import (
	"errors"

	"github.com/injunweb/backend-server/internal/global/utils"
)

type ApplicationRequestDTO struct {
	Email               string `json:"email"`
	ProjectName         string `json:"project_name"`
	Description         string `json:"description"`
	GithubRepositoryUrl string `json:"github_repository_url"`
}

func (applicationRequestDTO *ApplicationRequestDTO) Validate() error {
	if !utils.IsEmail(applicationRequestDTO.Email) {
		return errors.New("email is required")
	}

	if utils.IsEmptyString(applicationRequestDTO.ProjectName) {
		return errors.New("project_name is required")
	}

	if utils.IsEmptyString(applicationRequestDTO.Description) {
		return errors.New("description is required")
	}

	if !utils.IsGithubRepositoryUrl(applicationRequestDTO.GithubRepositoryUrl) {
		return errors.New("github_repository_url is required")
	}

	return nil
}
