package dao

import (
	"diabetes-agent-backend/response"
	"time"
)

func GetBloodGlucoseRecords(userEmail string, start, end time.Time) ([]response.GetBloodGlucoseRecordsResponse, error) {
	var records []response.GetBloodGlucoseRecordsResponse
	err := DB.Table("blood_glucose_record").
		Select("value, measured_at, dining_status").
		Where("user_email = ? AND measured_at BETWEEN ? AND ?", userEmail, start, end).
		Order("measured_at ASC").
		Find(&records).Error
	return records, err
}
