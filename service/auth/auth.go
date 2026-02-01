package auth

import (
	"crypto/md5"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/model"
	"diabetes-agent-server/request"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func UserRegister(req request.UserRegisterRequest) (model.User, error) {
	existingUser, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get user %s: %v", req.Email, err)
	}
	if existingUser != nil {
		return model.User{}, fmt.Errorf("email %s has already been used", req.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Avatar:   "https://api.dicebear.com/7.x/avataaars/svg?seed=" + generateAvatarSeed(req.Email),
	}
	if err := dao.DB.Create(&user).Error; err != nil {
		return model.User{}, err
	}

	return user, nil
}

func generateAvatarSeed(email string) string {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(normalizedEmail))
	return fmt.Sprintf("%x", hash)
}

func UserLogin(req request.UserLoginRequest) (*model.User, error) {
	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %v", req.Email, err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email: %s", req.Email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid password: %v", err)
	}
	return user, nil
}
