package request

type UpdateSessionTitleRequest struct {
	SessionID string `json:"session_id"`
	Title     string `json:"title"`
}
