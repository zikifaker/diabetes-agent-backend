package response

type GetHealthProfileResponse struct {
	DiabetesType  string `json:"diabetes_type"`
	Medication    string `json:"medication"`
	Complications string `json:"complications"`
}
