package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"
	"errors"

	"gorm.io/gorm"
)

func GetHealthProfile(email string) (*response.GetHealthProfileResponse, error) {
	var profile response.GetHealthProfileResponse
	err := DB.Model(&model.HealthProfile{}).
		Select("gender, age, height, weight, dietary_preference, smoking_status, activity_level, diabetes_type, diagnosis_year, therapy_mode, medication, allergies, complications").
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
