package controller

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
	knowledgebase "diabetes-agent-backend/service/knowledge-base"
	"diabetes-agent-backend/service/knowledge-base/etl"
	"diabetes-agent-backend/service/mq"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPolicyToken(c *gin.Context) {
	email := c.GetString("email")
	policyToken, err := knowledgebase.GeneratePolicyToken(email)
	if err != nil {
		slog.Error("Get Policy Token", "err", ErrGeneratePolicyToken)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGeneratePolicyToken.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: policyToken,
	})
}

func GetKnowledgeMetadata(c *gin.Context) {
	email := c.GetString("email")
	metadata, err := dao.GetKnowledgeMetadataByEmail(email)
	if err != nil {
		slog.Error("Get Knowledge Metadata", "err", ErrGetKnowledgeMetadata)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetKnowledgeMetadata.Error(),
		})
		return
	}

	var resp response.GetKnowledgeMetadataResponse
	for _, item := range metadata {
		resp.Metadata = append(resp.Metadata, response.MetadataResponse{
			FileName: item.FileName,
			FileType: string(item.FileType),
			FileSize: item.FileSize,
		})
	}

	c.JSON(http.StatusOK, response.Response{
		Data: resp,
	})
}

// UploadKnowledgeMetadata 在前端将文件成功传输到OSS后调用
// 存储知识元数据，向MQ发送向量化任务
func UploadKnowledgeMetadata(c *gin.Context) {
	var req request.UploadKnowledgeMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Upload Knowledge Metadata", "err", ErrParseRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	err := knowledgebase.UploadKnowledgeMetadata(req, email)
	if err != nil {
		slog.Error("Upload Knowledge Metadata", "err", ErrUploadKnowledgeMetadata)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUploadKnowledgeMetadata.Error(),
		})
		return
	}

	mq.SendMessage(c.Request.Context(), &mq.Message{
		Topic: mq.TopicKnowledgeBase,
		Tag:   mq.TagETL,
		Payload: etl.ETLMessage{
			FileType:   req.FileType,
			ObjectName: req.ObjectName,
		},
	})

	c.JSON(http.StatusOK, response.Response{})
}

func DeleteKnowledgeMetadata(c *gin.Context) {
	email := c.GetString("email")
	fileName := c.Query("file-name")
	err := dao.DeleteKnowledgeMetadataByEmailAndFileName(email, fileName)
	if err != nil {
		slog.Error("Delete Knowledge Metadata", "err", ErrDeleteKnowledgeMetadata)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrDeleteKnowledgeMetadata.Error(),
		})
		return
	}

	// TODO: 发送删除消息到MQ

	c.JSON(http.StatusOK, response.Response{})
}
