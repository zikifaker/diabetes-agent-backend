package dao

import (
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/response"
	"time"
)

func GetBloodGlucoseRecords(email string, start, end time.Time) ([]response.GetBloodGlucoseRecordsResponse, error) {
	var records []response.GetBloodGlucoseRecordsResponse
	err := DB.Model(&model.BloodGlucoseRecord{}).
		Select("value, measured_at, dining_status").
		Where("user_email = ? AND measured_at BETWEEN ? AND ?", email, start, end).
		Order("measured_at ASC").
		Find(&records).Error
	return records, err
}
