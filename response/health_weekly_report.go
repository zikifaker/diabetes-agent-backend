package response

type GetHealthWeeklyReportsResponse struct {
	StartAt  string `json:"start_at"`
	EndAt    string `json:"end_at"`
	FileName string `json:"file_name"`
}
