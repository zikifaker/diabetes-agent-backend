package dao

import (
	"diabetes-agent-server/model"
	"diabetes-agent-server/response"
	"time"
)

type BloodGlucoseStats struct {
	Min   float32 `gorm:"column:min_val" json:"min_val"`
	Max   float32 `gorm:"column:max_val" json:"max_val"`
	Avg   float32 `gorm:"column:avg_val" json:"avg_val"`
	Count int     `gorm:"column:count" json:"count"`
}

func GetBloodGlucoseRecords(email string, start, end time.Time) ([]response.GetBloodGlucoseRecordsResponse, error) {
	var records []response.GetBloodGlucoseRecordsResponse
	err := DB.Model(&model.BloodGlucoseRecord{}).
		Select("value, measured_at, dining_status").
		Where("user_email = ? AND measured_at BETWEEN ? AND ?", email, start, end).
		Order("measured_at ASC").
		Find(&records).Error
	return records, err
}

// GetBloodGlucoseStats 获取指定时间范围内的血糖统计信息
func GetBloodGlucoseStats(email string, start, end time.Time) (*BloodGlucoseStats, error) {
	var stats BloodGlucoseStats
	err := DB.Model(&model.BloodGlucoseRecord{}).
		Select("MIN(value) as min_val, MAX(value) as max_val, AVG(value) as avg_val, COUNT(*) as count").
		Where("user_email = ? AND measured_at BETWEEN ? AND ?", email, start, end).
		Take(&stats).Error
	return &stats, err
}
