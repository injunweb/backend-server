package repositories

import (
	"github.com/injunweb/backend-server/internal/domain/project/entities"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	database *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{database: db}
}

func (projectRepository *ProjectRepository) CreateProject(project *entities.Project) error {
	return projectRepository.database.Create(&project).Error
}

func (projectRepository *ProjectRepository) GetProjects() ([]entities.Project, error) {
	var projects []entities.Project
	err := projectRepository.database.Find(&projects).Error
	return projects, err
}

func (projectRepository *ProjectRepository) GetProjectByProjectName(projectName string) (*entities.Project, error) {
	var project entities.Project
	err := projectRepository.database.Where("project_name = ?", projectName).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}
