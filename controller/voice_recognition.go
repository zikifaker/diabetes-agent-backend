package controller

import (
	"diabetes-agent-backend/response"
	vr "diabetes-agent-backend/service/voice-recognition"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ChatVoiceRecognition(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		slog.Error(ErrGetAudioFile.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrGetAudioFile.Error(),
		})
		return
	}

	result, err := vr.Recognize(file)
	if err != nil {
		slog.Error(ErrVoiceRecognition.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrVoiceRecognition.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: struct {
			Text string `json:"text"`
		}{
			Text: result,
		},
	})
}
