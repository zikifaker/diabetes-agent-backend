package response

import "time"

type GetBloodGlucoseRecordsResponse struct {
	Records []BloodGlucoseRecordResponse `json:"records"`
}

type BloodGlucoseRecordResponse struct {
	Value        float32   `json:"value"`
	MeasuredAt   time.Time `json:"measured_at"`
	DiningStatus string    `json:"dining_status"`
}
