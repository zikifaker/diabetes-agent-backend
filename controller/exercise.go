package controller

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
	"diabetes-agent-backend/utils"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetExerciseRecords(c *gin.Context) {
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

	records, err := dao.GetExerciseRecords(email, start, end)
	if err != nil {
		slog.Error(ErrGetExerciseRecords.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetExerciseRecords.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: records,
	})
}

func CreateExerciseRecord(c *gin.Context) {
	var req request.ExerciseRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	exercise := model.ExerciseRecord{
		UserEmail:   email,
		Type:        req.Type,
		Name:        req.Name,
		Intensity:   req.Intensity,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Duration:    req.Duration,
		PreGlucose:  req.PreGlucose,
		PostGlucose: req.PostGlucose,
		Notes:       req.Notes,
	}
	if err := dao.DB.Create(&exercise).Error; err != nil {
		slog.Error(ErrCreateExerciseRecord.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrCreateExerciseRecord.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.Response{})
}

func DeleteExerciseRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := dao.DeleteExerciseRecord(uint(id)); err != nil {
		slog.Error(ErrDeleteExerciseRecord.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrDeleteExerciseRecord.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{})
}
