package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"diabetes-agent-server/dao"
	"diabetes-agent-server/model"
	"diabetes-agent-server/request"
	"diabetes-agent-server/response"
)

func CreateHealthProfile(c *gin.Context) {
	var req request.HealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	profile := convertRequestToModel(req, email)
	if err := dao.DB.Create(&profile).Error; err != nil {
		slog.Error(ErrCreateHealthProfile.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrCreateHealthProfile.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, response.Response{})
}

func GetHealthProfile(c *gin.Context) {
	email := c.GetString("email")
	profile, err := dao.GetHealthProfile(email)
	if err != nil {
		slog.Error(ErrGetHealthProfile.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetHealthProfile.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response.Response{
		Data: profile,
	})
}

func UpdateHealthProfile(c *gin.Context) {
	var req request.HealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	profile := convertRequestToModel(req, email)
	err := dao.UpdateHealthProfile(profile)
	if err != nil {
		slog.Error(ErrUpdateHealthProfile.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUpdateHealthProfile.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{})
}

func convertRequestToModel(req request.HealthProfileRequest, email string) model.HealthProfile {
	return model.HealthProfile{
		UserEmail:         email,
		Gender:            req.Gender,
		Age:               req.Age,
		Height:            req.Height,
		Weight:            req.Weight,
		DietaryPreference: req.DietaryPreference,
		SmokingStatus:     req.SmokingStatus,
		ActivityLevel:     req.ActivityLevel,
		DiabetesType:      req.DiabetesType,
		DiagnosisYear:     req.DiagnosisYear,
		TherapyMode:       req.TherapyMode,
		Medication:        req.Medication,
		Allergies:         req.Allergies,
		Complications:     req.Complications,
	}
}
