package response

type UserAuthResponse struct {
	Email                          string `json:"email"`
	Avatar                         string `json:"avatar"`
	EnableWeeklyReportNotification bool   `json:"enable_weekly_report_notification"`
	Token                          string `json:"token"`
}
