package request

type OSSAuthRequest struct {
	Namespace string `json:"namespace"`
	Email     string `json:"email"`
	SessionID string `json:"session_id"`
	FileName  string `json:"file_name"`
}
