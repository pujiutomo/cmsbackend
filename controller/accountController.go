package controller

import (
	"fmt"

	"github.com/pujiutomo/cmsbackend/database"
	"github.com/pujiutomo/cmsbackend/models"
	"gorm.io/gorm"
)

func GetUserById(id string) (*models.User, error) {
	var user models.User
	result := database.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("User with id %d not found", id)
		}
		return nil, result.Error
	}
	return &user, nil
}
