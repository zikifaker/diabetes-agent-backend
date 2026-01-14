package request

import "time"

type CreateBloodGlucoseRecordRequest struct {
	Value        float32   `json:"value" binding:"required"`
	MeasuredAt   time.Time `json:"measured_at" binding:"required"`
	DiningStatus string    `json:"dining_status" binding:"required"`
}
