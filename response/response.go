package response

type Response struct {
	// 错误信息
	Msg string `json:"msg"`

	// 业务数据
	Data any `json:"data"`
}
