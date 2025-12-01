package response

type UserAuthResponse struct {
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Token  string `json:"token"`
}
