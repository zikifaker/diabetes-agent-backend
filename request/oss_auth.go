package request

// OSSAuthRequest 用于生成 Object Key
type OSSAuthRequest struct {
	Namespace string `json:"namespace"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	FileName  string `json:"file_name"`
	StartAt   string `json:"start_at"`
	EndAt     string `json:"end_at"`
}
