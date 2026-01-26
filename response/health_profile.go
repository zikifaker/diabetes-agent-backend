package response

type GetHealthProfileResponse struct {
	Gender            string  `json:"gender"`
	Age               int     `json:"age"`
	Height            float32 `json:"height"`
	Weight            float32 `json:"weight"`
	DietaryPreference string  `json:"dietary_preference"`
	SmokingStatus     bool    `json:"smoking_status"`
	ActivityLevel     string  `json:"activity_level"`
	DiabetesType      string  `json:"diabetes_type"`
	DiagnosisYear     int     `json:"diagnosis_year"`
	TherapyMode       string  `json:"therapy_mode"`
	Medication        string  `json:"medication"`
	Allergies         string  `json:"allergies"`
	Complications     string  `json:"complications"`
}
