package controller

import (
	"diabetes-agent-server/request"
	"diabetes-agent-server/response"
	ossauth "diabetes-agent-server/service/oss-auth"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPolicyToken(c *gin.Context) {
	policyToken, err := ossauth.GeneratePolicyToken(request.OSSAuthRequest{
		Namespace: c.Query("namespace"),
		Email:     c.GetString("email"),
		SessionID: c.Query("session-id"),
		FileName:  c.Query("file-name"),
	})
	if err != nil {
		slog.Error(ErrGeneratePolicyToken.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGeneratePolicyToken.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: policyToken,
	})
}

func GetPresignedURL(c *gin.Context) {
	url, err := ossauth.GeneratePresignedURL(request.OSSAuthRequest{
		Namespace: c.Query("namespace"),
		Email:     c.GetString("email"),
		SessionID: c.Query("session-id"),
		FileName:  c.Query("file-name"),
	})
	if err != nil {
		slog.Error(ErrGetPreSignedURL.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetPreSignedURL.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: response.GetPreSignedURLResponse{
			URL: url,
		},
	})
}
