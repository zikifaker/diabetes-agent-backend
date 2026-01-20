package controller

import (
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/model"
	"diabetes-agent-backend/request"
	"diabetes-agent-backend/response"
	knowledgebase "diabetes-agent-backend/service/knowledge-base"
	"diabetes-agent-backend/service/knowledge-base/etl"
	"diabetes-agent-backend/service/mq"
	ossauth "diabetes-agent-backend/service/oss-auth"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetKnowledgeMetadata(c *gin.Context) {
	email := c.GetString("email")
	metadata, err := dao.GetKnowledgeMetadataByEmail(email)
	if err != nil {
		slog.Error(ErrGetKnowledgeMetadata.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetKnowledgeMetadata.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: metadata,
	})
}

// UploadKnowledgeMetadata 在前端将文件成功传输到 OSS 后调用
// 存储知识文件元数据，向 MQ 发送向量化任务
func UploadKnowledgeMetadata(c *gin.Context) {
	var req request.UploadKnowledgeMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(ErrParseRequest.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Response{
			Msg: ErrParseRequest.Error(),
		})
		return
	}

	email := c.GetString("email")
	err := knowledgebase.UploadKnowledgeMetadata(req, email)
	if err != nil {
		slog.Error(ErrUploadKnowledgeMetadata.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUploadKnowledgeMetadata.Error(),
		})
		return
	}

	mq.SendMessage(c.Request.Context(), &mq.Message{
		Topic: mq.TopicKnowledgeBase,
		Tag:   mq.TagETL,
		Payload: etl.ETLMessage{
			FileType:   model.FileType(req.FileType),
			ObjectName: req.ObjectName,
		},
	})

	c.JSON(http.StatusOK, response.Response{})
}

// DeleteKnowledgeMetadata 删除知识文件元数据和 OSS 上的文件，向 MQ 发送删除任务
func DeleteKnowledgeMetadata(c *gin.Context) {
	email := c.GetString("email")
	fileName := c.Query("file-name")
	err := dao.DeleteKnowledgeMetadataByEmailAndFileName(email, fileName)
	if err != nil {
		slog.Error(ErrDeleteKnowledgeMetadata.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrDeleteKnowledgeMetadata.Error(),
		})
		return
	}

	objectName, err := ossauth.GenerateKey(request.OSSAuthRequest{
		Namespace: ossauth.OSSKeyPrefixKnowledgeBase,
		Email:     email,
		FileName:  fileName,
	})
	if err != nil {
		slog.Error(ErrGenerateOSSKey.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGenerateOSSKey.Error(),
		})
		return
	}

	extension := filepath.Ext(fileName)
	fileType := strings.TrimPrefix(extension, ".")

	mq.SendMessage(c.Request.Context(), &mq.Message{
		Topic: mq.TopicKnowledgeBase,
		Tag:   mq.TagDelete,
		Payload: etl.DeleteMessage{
			FileType:   model.FileType(fileType),
			ObjectName: objectName,
		},
	})

	c.JSON(http.StatusOK, response.Response{})
}

func SearchKnowledgeMetadata(c *gin.Context) {
	email := c.GetString("email")
	query := c.Query("query")

	if query == "" {
		c.JSON(http.StatusBadRequest, response.Response{
			Msg: "search query is empty",
		})
		return
	}

	metadata, err := dao.SearchKnowledgeMetadataByFullText(email, query)
	if err != nil {
		slog.Error(ErrSearchKnowledgeMetadata.Error(), "err", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
			Msg: ErrSearchKnowledgeMetadata.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: metadata,
	})
}
