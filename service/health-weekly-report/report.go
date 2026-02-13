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
	ossauth "diabetes-agent-server/service/oss-auth"
	"diabetes-agent-server/utils"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
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

	//go:embed template.html
	reportTemplate string
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

// ReportData 用于填充健康周报模板
type ReportData struct {
	ReportPeriod        string
	BloodGlucoseRecords []response.GetBloodGlucoseRecordsResponse
	BloodGlucoseStats   *dao.BloodGlucoseStats
	ExerciseRecords     []response.GetExerciseRecordsResponse
	ExerciseStats       *dao.ExerciseStats
	HealthAnalysis      *HealthAnalysis
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

	// 计算时间范围: 上周一00:00:00 - 上周日23:59:59.999
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

	formattedStart := start.Format("2006/01/02")
	formattedEnd := end.Format("2006/01/02")
	fileName := fmt.Sprintf("%s-%s.html", formattedStart, formattedEnd)

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

	msg := model.SystemMessage{
		UserEmail: email,
		Title:     "健康周报",
		Content:   fmt.Sprintf("您的健康周报(%s 至 %s)已生成，请查收。", formattedStart, formattedEnd),
		IsRead:    false,
	}
	if err := dao.DB.Create(&msg).Error; err != nil {
		return fmt.Errorf("Failed to save system message: %v", err)
	}

	// 更新未读消息计数
	key := fmt.Sprintf(constants.KeyUserUnreadMsgCount, email)
	dao.RedisClient.Incr(ctx, key)

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
		openai.WithBaseURL(constants.BaseURL),
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
