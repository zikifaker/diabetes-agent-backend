package response

import "time"

type GetBloodGlucoseRecordsResponse struct {
	Value        float32   `json:"value"`
	MeasuredAt   time.Time `json:"measured_at"`
	DiningStatus string    `json:"dining_status"`
}
