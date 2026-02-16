package constants

const (
	// 用户未读消息数的 Redis key
	KeyUnreadMsgCount = "user:%s:unread_msg_count"

	// 邮箱验证码的 Redis key
	KeyVerificationCode = "user:%s:verification_code"
)
