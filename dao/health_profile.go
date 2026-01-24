package dao

import (
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/response"
	"errors"

	"gorm.io/gorm"
)

func GetHealthProfile(email string) (*response.GetHealthProfileResponse, error) {
	var profile response.GetHealthProfileResponse
	err := DB.Table("health_profile").
		Select("diabetes_type, medication, complications").
		Where("user_email = ?", email).
		First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &profile, err
}

func UpdateHealthProfile(profile model.HealthProfile) error {
	return DB.Model(&model.HealthProfile{}).
		Where("user_email = ?", profile.UserEmail).
		Updates(profile).Error
}
