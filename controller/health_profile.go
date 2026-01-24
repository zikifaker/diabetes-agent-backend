package controller

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
)

func CreateHealthProfile(c *gin.Context) {
	var req request.CreateHealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	profile := model.HealthProfile{
		UserEmail:     email,
		DiabetesType:  req.DiabetesType,
		Medication:    req.Medication,
		Complications: req.Complications,
	}
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
	var req request.UpdateHealthProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	err := dao.UpdateHealthProfile(model.HealthProfile{
		UserEmail:     email,
		DiabetesType:  req.DiabetesType,
		Medication:    req.Medication,
		Complications: req.Complications,
	})
	if err != nil {
		slog.Error(ErrUpdateHealthProfile.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUpdateHealthProfile.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{})
}
