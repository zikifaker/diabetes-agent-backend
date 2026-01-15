package controller

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
	"log/slog"
	"net/http"
	"time"

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

	if req.Value < model.ValueMin || req.Value > model.ValueMax {
		slog.Error(ErrInvalidBloodGlucose.Error(), "value", req.Value)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrInvalidBloodGlucose.Error(),
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
}

func GetBloodGlucoseRecords(c *gin.Context) {
	email := c.GetString("email")
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		slog.Error(ErrInvalidDate.Error(),
			"start", startStr,
			"err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrInvalidDate.Error(),
		})
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		slog.Error(ErrInvalidDate.Error(),
			"end", endStr,
			"err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrInvalidDate.Error(),
		})
		return
	}

	if start.After(end) {
		slog.Error(ErrInvalidDateRange.Error(),
			"start", startStr,
			"end", endStr)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrInvalidDateRange.Error(),
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

	var resp response.GetBloodGlucoseRecordsResponse
	for _, record := range records {
		resp.Records = append(resp.Records, response.BloodGlucoseRecordResponse{
			Value:        record.Value,
			MeasuredAt:   record.MeasuredAt,
			DiningStatus: record.DiningStatus,
		})
	}

	c.JSON(http.StatusOK, response.Response{
		Data: resp,
	})
}
