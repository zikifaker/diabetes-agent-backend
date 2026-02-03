package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"
	"time"
)

type ExerciseStats struct {
	TotalMinutes   int     `json:"total_minutes"`
	AverageMinutes float64 `json:"average_minutes"`
	Count          int     `json:"count"`
}

func GetExerciseRecords(email string, start, end time.Time) ([]response.GetExerciseRecordsResponse, error) {
	var records []response.GetExerciseRecordsResponse
	err := DB.Model(&model.ExerciseRecord{}).
		Select("id, type, name, intensity, start_at, end_at, duration, pre_glucose, post_glucose, notes").
		Where("user_email = ? AND start_at BETWEEN ? AND ?", email, start, end).
		Order("start_at ASC").
		Find(&records).Error
	return records, err
}

func DeleteExerciseRecord(id uint) error {
	return DB.Where("id = ?", id).
		Delete(&model.ExerciseRecord{}).Error
}

func GetExerciseStats(email string, start, end time.Time) (*ExerciseStats, error) {
	var stats ExerciseStats
	err := DB.Model(&model.ExerciseRecord{}).
		Select("COUNT(*) as count, SUM(duration) as total_minutes, AVG(duration) as average_minutes").
		Where("user_email = ? AND start_at BETWEEN ? AND ?", email, start, end).
		Take(&stats).Error
	return &stats, err
}
