package request

import "errors"

type AdminRequestDTO struct {
	AdminName string `json:"admin_name"`
	Password  string `json:"password"`
}

func (adminRequestDTO *AdminRequestDTO) Validate() error {
	if adminRequestDTO.AdminName == "" {
		return errors.New("admin_name is required")
	}

	if adminRequestDTO.Password == "" {
		return errors.New("password is required")
	}

	return nil
}
