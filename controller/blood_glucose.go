package controller

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
	"diabetes-agent-backend/utils"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateBloodGlucoseRecord(c *gin.Context) {
	var req request.CreateBloodGlucoseRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	record := model.BloodGlucoseRecord{
		UserEmail:    email,
		Value:        req.Value,
		MeasuredAt:   req.MeasuredAt,
		DiningStatus: req.DiningStatus,
	}
	if err := dao.DB.Create(&record).Error; err != nil {
		slog.Error(ErrCreateBloodGlucoseRecord.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrCreateBloodGlucoseRecord.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.Response{})
}

func GetBloodGlucoseRecords(c *gin.Context) {
	email := c.GetString("email")
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, end, err := utils.ValidateTimeRange(startStr, endStr, "UTC")
	if err != nil {
		slog.Error(err.Error(),
			"start", startStr,
			"end", endStr)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: err.Error(),
		})
		return
	}

	records, err := dao.GetBloodGlucoseRecords(email, start, end)
	if err != nil {
		slog.Error(ErrGetBloodGlucoseRecords.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetBloodGlucoseRecords.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: records,
	})
}
