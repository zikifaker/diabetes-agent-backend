package request

import "time"

type CreateBloodGlucoseRecordRequest struct {
	Value        float32   `json:"value" binding:"required,gte=1,lte=50"`
	MeasuredAt   time.Time `json:"measured_at" binding:"required"`
	DiningStatus string    `json:"dining_status" binding:"required"`
}
