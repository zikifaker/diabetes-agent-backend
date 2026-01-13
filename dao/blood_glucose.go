package dao

import (
	"diabetes-agent-backend/model"
	"time"
)

func GetBloodGlucoseRecords(userEmail string, start, end time.Time) ([]model.BloodGlucoseRecord, error) {
	var records []model.BloodGlucoseRecord
	err := DB.Where("user_email = ? AND measured_at BETWEEN ? AND ?", userEmail, start, end).
		Order("measured_at ASC").
		Find(&records).Error
	return records, err
}
