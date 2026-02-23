package healthweeklyreport

import (
	"bytes"
	"context"
	"diabetes-agent-server/config"
	"diabetes-agent-server/constants"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/model"
	"diabetes-agent-server/request"
	"diabetes-agent-server/response"
	"diabetes-agent-server/service/email"
	ossauth "diabetes-agent-server/service/oss-auth"
	"diabetes-agent-server/utils"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/go-co-op/gocron"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const modelName = "deepseek-v3.1"

var (
	//go:embed prompts/report.txt
	reportPrompt string

	//go:embed report.html
	reportTemplate string

	//go:embed notification.html
	notificationTemplate string
)

type UserHealthData struct {
	BloodGlucoseRecords []response.GetBloodGlucoseRecordsResponse `json:"blood_glucose_records"`
	BloodGlucoseStats   *dao.BloodGlucoseStats                    `json:"blood_glucose_stats"`
	ExerciseRecords     []response.GetExerciseRecordsResponse     `json:"exercise_records"`
	ExerciseStats       *dao.ExerciseStats                        `json:"exercise_stats"`
	HealthProfile       *response.GetHealthProfileResponse        `json:"health_profile"`
}

// HealthAnalysis LLM 对近一周健康数据的分析结果
type HealthAnalysis struct {
	BloodGlucoseAnalysis string `json:"blood_glucose_analysis"`
	ExerciseAnalysis     string `json:"exercise_analysis"`
	RecommendedMeals     []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"recommended_meals"`
	Conclusion string `json:"conclusion"`
}

// ReportData 健康周报模板数据
type ReportData struct {
	ReportPeriod        string
	BloodGlucoseRecords []response.GetBloodGlucoseRecordsResponse
	BloodGlucoseStats   *dao.BloodGlucoseStats
	ExerciseRecords     []response.GetExerciseRecordsResponse
	ExerciseStats       *dao.ExerciseStats
	HealthAnalysis      *HealthAnalysis
}

// NotificationData 健康周报通知数据
type NotificationData struct {
	ReportPeriod string
	ReportURL    string
}

func SetupHealthWeeklyReportScheduler() {
	s := gocron.NewScheduler(time.UTC)

	// 每周一 2:00 执行
	_, err := s.Every(1).Monday().At("02:00").Do(GenerateWeeklyReports)
	if err != nil {
		slog.Error("Failed to schedule health report generation task", "err", err)
		return
	}

	s.StartAsync()
}

func GenerateWeeklyReports() {
	ctx := context.Background()

	users, err := dao.GetAllUsers()
	if err != nil {
		slog.Error("Failed to get users for health report", "err", err)
		return
	}

	now := time.Now().UTC()
	thisMonday := now.AddDate(0, 0, -int(now.Weekday())+1).Truncate(24 * time.Hour)
	end := thisMonday.Add(-time.Nanosecond)
	start := thisMonday.AddDate(0, 0, -7)

	for _, user := range users {
		if err := generateWeeklyReport(ctx, user.Email, start, end); err != nil {
			slog.Error("Failed to generate health report",
				"email", user.Email,
				"start", start,
				"end", end,
				"err", err,
			)
		}
	}
}

func generateWeeklyReport(ctx context.Context, email string, start, end time.Time) error {
	userHealthData, err := getUserHealthData(ctx, email, start, end)
	if err != nil {
		return fmt.Errorf("failed to get user health data: %v", err)
	}

	healthAnalysis, err := generateHealthAnalysis(ctx, userHealthData)
	if err != nil {
		return fmt.Errorf("failed to generate health analysis: %v", err)
	}

	formattedStart := start.Format("2006-01-02")
	formattedEnd := end.Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s.html", formattedStart, formattedEnd)

	htmlContent, err := renderReport(ctx, &ReportData{
		ReportPeriod:        formattedStart + " 至 " + formattedEnd,
		BloodGlucoseRecords: userHealthData.BloodGlucoseRecords,
		BloodGlucoseStats:   userHealthData.BloodGlucoseStats,
		ExerciseRecords:     userHealthData.ExerciseRecords,
		ExerciseStats:       userHealthData.ExerciseStats,
		HealthAnalysis:      healthAnalysis,
	})
	if err != nil {
		return fmt.Errorf("failed to render report: %v", err)
	}

	objectName, err := ossauth.GenerateKey(request.OSSAuthRequest{
		Namespace: ossauth.OSSKeyPrefixHealthWeeklyReport,
		Email:     email,
		FileName:  fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to generate oss key: %v", err)
	}

	// 存储健康周报元数据
	if err := dao.DB.Create(&model.HealthWeeklyReport{
		UserEmail:  email,
		StartAt:    start,
		EndAt:      end,
		FileName:   fileName,
		ObjectName: objectName,
	}).Error; err != nil {
		return fmt.Errorf("failed to save health weekly report: %v", err)
	}

	// 上传健康周报到 OSS
	if err := uploadReport(ctx, htmlContent, objectName); err != nil {
		return fmt.Errorf("failed to upload health weekly report: %v", err)
	}

	// 存储系统消息
	msg := model.SystemMessage{
		UserEmail: email,
		Title:     "健康周报",
		Content:   fmt.Sprintf("您的健康周报(%s 至 %s)已生成，请查收。", formattedStart, formattedEnd),
		IsRead:    false,
	}
	if err := dao.DB.Create(&msg).Error; err != nil {
		slog.Error("Failed to save system message", "err", err)
	}

	// 更新未读消息计数
	key := fmt.Sprintf(constants.KeyUnreadMsgCount, email)
	dao.RedisClient.Incr(ctx, key)

	// 推送通知邮件
	if err := sendNotification(email, NotificationData{
		ReportPeriod: formattedStart + " 至 " + formattedEnd,
		ReportURL:    fmt.Sprintf("%s/health-weekly-report", config.Cfg.Client.BaseURL),
	}); err != nil {
		slog.Error("failed to send notification", "err", err)
	}

	return nil
}

func getUserHealthData(ctx context.Context, email string, start, end time.Time) (*UserHealthData, error) {
	bloodGlucoseRecords, err := dao.GetBloodGlucoseRecords(email, start, end)
	if err != nil {
		return nil, err
	}

	bloodGlucoseStats, err := dao.GetBloodGlucoseStats(email, start, end)
	if err != nil {
		return nil, err
	}

	exerciseRecords, err := dao.GetExerciseRecords(email, start, end)
	if err != nil {
		return nil, err
	}

	exerciseStats, err := dao.GetExerciseStats(email, start, end)
	if err != nil {
		return nil, err
	}

	healthProfile, err := dao.GetHealthProfile(email)
	if err != nil {
		return nil, err
	}

	return &UserHealthData{
		BloodGlucoseRecords: bloodGlucoseRecords,
		BloodGlucoseStats:   bloodGlucoseStats,
		ExerciseRecords:     exerciseRecords,
		ExerciseStats:       exerciseStats,
		HealthProfile:       healthProfile,
	}, nil
}

// 调用 LLM 对近一周健康数据进行分析
func generateHealthAnalysis(ctx context.Context, userHealthData *UserHealthData) (*HealthAnalysis, error) {
	userHealthDataJSON, _ := json.Marshal(userHealthData)

	template := prompts.NewPromptTemplate(reportPrompt, []string{"user_health_data"})
	prompt, err := template.Format(map[string]any{"user_health_data": userHealthDataJSON})
	if err != nil {
		return nil, err
	}

	llm, err := openai.New(
		openai.WithModel(modelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(config.Cfg.Model.BaseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		return nil, err
	}

	result, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		return nil, err
	}

	slog.Debug("generated health analysis", "result", result)

	// 去除 Markdown 格式的 json 字符串前后缀
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimSuffix(result, "```")

	var healthAnalysis HealthAnalysis
	if err := json.Unmarshal([]byte(result), &healthAnalysis); err != nil {
		return nil, err
	}

	return &healthAnalysis, nil
}

// 渲染健康周报的 HTML 模板
func renderReport(ctx context.Context, data *ReportData) (string, error) {
	tmpl, err := template.New("report").
		Funcs(template.FuncMap{
			"json": func(v interface{}) string {
				bytes, _ := json.Marshal(v)
				return string(bytes)
			},
		}).
		Parse(reportTemplate)

	if err != nil {
		return "", fmt.Errorf("failed to parse template file: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}
	return buf.String(), nil
}

func uploadReport(ctx context.Context, content string, objectName string) error {
	cfg := oss.NewConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.Cfg.OSS.AccessKeyID,
			config.Cfg.OSS.AccessKeySecret,
		)).
		WithRegion(config.Cfg.OSS.Region).
		WithHttpClient(utils.GlobalHTTPClient)
	client := oss.NewClient(cfg)

	req := oss.PutObjectRequest{
		Bucket:      oss.Ptr(config.Cfg.OSS.BucketName),
		Key:         oss.Ptr(objectName),
		Body:        strings.NewReader(content),
		ContentType: oss.Ptr("text/html; charset=utf-8"),
	}

	_, err := client.PutObject(ctx, &req)
	if err != nil {
		return fmt.Errorf("failed to upload health weekly report: %v", err)
	}
	return nil
}

func sendNotification(toEmail string, data NotificationData) error {
	user, err := dao.GetUserByEmail(toEmail)
	if err != nil {
		slog.Error("failed to get user by email",
			"email", toEmail,
			"err", err,
		)
		return nil
	}

	// 若用户未开启健康周报通知，直接返回
	if !user.EnableWeeklyReportNotification {
		return nil
	}

	cfg := config.Cfg.Email
	message, err := buildNotificationMessage(cfg.FromEmail, toEmail, data)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth(
		"",
		cfg.FromEmail,
		cfg.Password,
		cfg.Host,
	)
	return email.Send(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		auth,
		cfg.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)
}

func buildNotificationMessage(fromEmail string, toEmail string, data NotificationData) (string, error) {
	var content strings.Builder

	header := make(map[string]string)
	header["From"] = fmt.Sprintf("%s <%s>", "Diabetes Agent", fromEmail)
	header["To"] = toEmail
	header["Subject"] = "您的健康周报已生成"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=UTF-8"
	for k, v := range header {
		content.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	tmpl, err := template.New("notification_message").Parse(notificationTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	content.WriteString("\r\n")
	content.WriteString(buf.String())

	return content.String(), nil
}
