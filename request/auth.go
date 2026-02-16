package request

type UserRegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password"`
	Code     string `json:"code"`
	Type     string `json:"type" binding:"required"`
}

type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}
