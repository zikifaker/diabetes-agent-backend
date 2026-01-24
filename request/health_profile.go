package request

type CreateHealthProfileRequest struct {
	DiabetesType  string `json:"diabetes_type"`
	Medication    string `json:"medication"`
	Complications string `json:"complications"`
}

type UpdateHealthProfileRequest struct {
	DiabetesType  string `json:"diabetes_type"`
	Medication    string `json:"medication"`
	Complications string `json:"complications"`
}
