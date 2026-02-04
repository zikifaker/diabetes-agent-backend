package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"
)

func GetHealthWeeklyReports(email string) ([]response.GetHealthWeeklyReportsResponse, error) {
	var reports []response.GetHealthWeeklyReportsResponse
	err := DB.Model(&model.HealthWeeklyReport{}).
		Where("user_email = ?", email).
		Order("start_at DESC").
		Find(&reports).Error
	return reports, err
}
