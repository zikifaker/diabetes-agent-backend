package request

// OSSAuthRequest 用于生成 Object Key
type OSSAuthRequest struct {
	Namespace       string `json:"namespace"`
	Email           string `json:"email"`
	SessionID       string `json:"session_id"`
	FileName        string `json:"file_name"`
	UseCustomDomain bool   `json:"use_custom_domain"`
}
