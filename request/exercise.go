package request

import "time"

type ExerciseRecordRequest struct {
	Type        string    `json:"type" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Intensity   string    `json:"intensity" binding:"required"`
	StartAt     time.Time `json:"start_at" binding:"required"`
	EndAt       time.Time `json:"end_at" binding:"required"`
	Duration    int       `json:"duration" binding:"required"`
	PreGlucose  float32   `json:"pre_glucose" binding:"gte=1,lte=50"`
	PostGlucose float32   `json:"post_glucose" binding:"gte=1,lte=50"`
	Notes       string    `json:"notes"`
}
