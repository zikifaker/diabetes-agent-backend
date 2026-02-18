package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"diabetes-agent-server/dao"
	"diabetes-agent-server/request"
	"diabetes-agent-server/response"
)

func GetHealthWeeklyReports(c *gin.Context) {
	email := c.GetString("email")
	reports, err := dao.GetHealthWeeklyReports(email)
	if err != nil {
		slog.Error(ErrGetHealthWeeklyReports.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetHealthWeeklyReports.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: reports,
	})
}

func UpdateUserEnableNotification(c *gin.Context) {
	var req request.UpdateUserEnableNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	if err := dao.UpdateEnableNotification(email, req.EnableWeeklyReportNotification); err != nil {
		slog.Error(ErrUpdateUserEnableNotification.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUpdateUserEnableNotification.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.Response{})
}
