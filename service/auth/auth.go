package auth

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"diabetes-agent-server/config"
	"diabetes-agent-server/constants"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/model"
	"diabetes-agent-server/request"
	"diabetes-agent-server/service/email"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"math/big"
	"net/smtp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	loginTypePassword = "password"
	loginTypeCode     = "code"

	// 验证码的随机数源
	codeNumbers = "0123456789"

	// 验证码过期时间
	codeExpiration = 5 * time.Minute

	// 验证码发送间隔
	codeInterval = 1 * time.Minute
)

//go:embed template.html
var emailTemplate string

type EmailData struct {
	Code       string
	Expiration int
}

func UserRegister(req request.UserRegisterRequest) (*model.User, error) {
	// 检查邮箱是否已注册
	existingUser, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %v", req.Email, err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("%s is used", req.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Avatar:   "https://api.dicebear.com/7.x/avataaars/svg?seed=" + generateAvatarSeed(req.Email),
	}
	if err := dao.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func generateAvatarSeed(email string) string {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(normalizedEmail))
	return fmt.Sprintf("%x", hash)
}

func UserLogin(req request.UserLoginRequest) (*model.User, error) {
	// 检查用户是否存在
	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %v", req.Email, err)
	}
	if user == nil {
		return nil, fmt.Errorf("%s is not registered", req.Email)
	}

	switch req.Type {
	case loginTypePassword:
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			return nil, fmt.Errorf("invalid password: %v", err)
		}
	case loginTypeCode:
		if err := verifyCode(req.Email, req.Code); err != nil {
			return nil, fmt.Errorf("invalid code: %v", err)
		}
	default:
		return nil, fmt.Errorf("invalid login type: %s", req.Type)
	}

	return user, nil
}

func verifyCode(email, code string) error {
	if code == "" {
		return fmt.Errorf("required verification code")
	}

	ctx := context.Background()
	key := fmt.Sprintf(constants.KeyVerificationCode, email)
	exists, err := dao.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check verification code: %v", err)
	}
	if exists == 0 {
		return fmt.Errorf("verification code not found")
	}

	storedCode, err := dao.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get verification code: %v", err)
	}
	if storedCode != code {
		return fmt.Errorf("verification code mismatch")
	}

	// 校验成功后删除 key，防止重复使用
	if _, err := dao.RedisClient.Del(ctx, key).Result(); err != nil {
		slog.Error("error deleting verification code", "err", err)
	}

	return nil
}

func SendVerificationCode(req request.SendEmailCodeRequest) error {
	// 检查用户是否存在
	user, err := dao.GetUserByEmail(req.Email)
	if err != nil {
		return fmt.Errorf("failed to get user %s: %v", req.Email, err)
	}
	if user == nil {
		return fmt.Errorf("%s is not registered", req.Email)
	}

	// 检查 code key 是否存在
	key := fmt.Sprintf(constants.KeyVerificationCode, req.Email)
	ctx := context.Background()
	exists, err := dao.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check verification code: %v", err)
	}

	// 若 code key 存在，检查是否在间隔窗口内
	if exists > 0 {
		ttl := dao.RedisClient.TTL(ctx, key).Val()
		if ttl > 0 {
			// key 从创建到当前的时间间隔
			elapsedTime := codeExpiration - ttl
			if elapsedTime < codeInterval {
				return fmt.Errorf("sending code too frequently, please try again later")
			}
		}
	}

	// 存储验证码到 Redis，设置过期时间
	code := generateCode()
	err = dao.RedisClient.Set(ctx, key, code, codeExpiration).Err()
	if err != nil {
		return fmt.Errorf("failed to save verification code: %v", err)
	}

	if err := sendEmail(req.Email, code); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

func generateCode() string {
	code := ""
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += string(codeNumbers[num.Int64()])
	}
	return code
}

func sendEmail(toEmail, code string) error {
	cfg := config.Cfg.Email
	message, err := buildVerificationCodeMessage(cfg.FromEmail, toEmail, code)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth(
		"",
		cfg.FromEmail,
		cfg.Password,
		cfg.Host,
	)
	return email.Send(
		fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		auth,
		cfg.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)
}

func buildVerificationCodeMessage(fromEmail, toEmail, code string) (string, error) {
	var content strings.Builder

	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", "Diabetes Agent", fromEmail)
	header["To"] = toEmail
	header["Subject"] = "验证码"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	for k, v := range header {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	tmpl, err := template.New("verification_code_message").Parse(emailTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, EmailData{
		Code:       code,
		Expiration: int(codeExpiration.Minutes()),
	})
	if err != nil {
		return "", err
	}

	content.WriteString("\r\n")
	content.WriteString(buf.String())

	return content.String(), nil
}
