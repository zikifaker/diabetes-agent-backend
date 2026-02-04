package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"diabetes-agent-server/dao"
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
