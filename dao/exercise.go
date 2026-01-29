package dao

import (
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/response"
	"time"
)

func GetExerciseRecords(email string, start, end time.Time) ([]response.GetExerciseRecordsResponse, error) {
	var records []response.GetExerciseRecordsResponse
	err := DB.Model(&model.ExerciseRecord{}).
		Select("type, name, intensity, start_at, end_at, duration, pre_glucose, post_glucose, notes").
		Where("user_email = ? AND start_at BETWEEN ? AND ?", email, start, end).
		Order("start_at ASC").
		Find(&records).Error
	return records, err
}
