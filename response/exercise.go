package response

import "time"

type GetExerciseRecordsResponse struct {
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Intensity   string    `json:"intensity"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Duration    int       `json:"duration"`
	PreGlucose  float32   `json:"pre_glucose"`
	PostGlucose float32   `json:"post_glucose"`
	Notes       string    `json:"notes"`
}
