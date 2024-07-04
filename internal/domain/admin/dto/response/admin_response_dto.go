package response

type ApplicationResponseDTO struct {
	Email               string `json:"email"`
	ProjectName         string `json:"project_name"`
	Description         string `json:"description"`
	GithubRepositoryUrl string `json:"github_repository_url"`
}

type ProjectResponseDTO struct {
	Email               string `json:"email"`
	ProjectName         string `json:"project_name"`
	Description         string `json:"description"`
	GithubRepositoryUrl string `json:"github_repository_url"`
	AccessKey           string `json:"access_key"`
}
