package dao

import (
	"diabetes-agent-server/model"
	"errors"

	"gorm.io/gorm"
)

func GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		// 邮箱对应的用户不存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := DB.Find(&users).Error
	return users, err
}
