package repositories

import (
	"github.com/injunweb/backend-server/internal/domain/admin/entities"
	"gorm.io/gorm"
)

type AdminRepository struct {
	database *gorm.DB
}

func NewAdminRepository(database *gorm.DB) *AdminRepository {
	return &AdminRepository{database: database}
}

func (adminRepository *AdminRepository) GetAdminByUsername(username string) (*entities.Admin, error) {
	var admin entities.Admin

	err := adminRepository.database.Where("admin_name = ?", username).First(&admin).Error

	if err != nil {
		return nil, err
	}

	return &admin, nil
}
