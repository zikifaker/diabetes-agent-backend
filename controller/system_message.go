package controller

import (
	"context"
	"diabetes-agent-server/constants"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/response"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetSystemMessages(c *gin.Context) {
	email := c.GetString("email")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	messages, err := dao.GetSystemMessages(email, page)
	if err != nil {
		slog.Error(ErrGetSystemMessages.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetSystemMessages.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: messages,
	})
}

func UpdateSystemMessageAsRead(c *gin.Context) {
	id := c.Param("id")
	message, err := dao.GetSystemMessageById(id)
	if err != nil {
		slog.Error(ErrUpdateSystemMessageAsRead.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUpdateSystemMessageAsRead.Error(),
		})
		return
	}

	// 检查是否已读，避免重复扣减
	if message.IsRead {
		c.JSON(http.StatusOK, response.Response{})
		return
	}

	if err := dao.UpdateSystemMessageAsRead(id); err != nil {
		slog.Error(ErrUpdateSystemMessageAsRead.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrUpdateSystemMessageAsRead.Error(),
		})
		return
	}

	// 更新未读计数
	ctx := context.Background()
	key := fmt.Sprintf(constants.KeyUnreadMsgCount, message.UserEmail)
	dao.RedisClient.Decr(ctx, key)

	c.JSON(http.StatusOK, response.Response{})
}

func DeleteSystemMessage(c *gin.Context) {
	id := c.Param("id")
	message, err := dao.DeleteSystemMessage(id)
	if err != nil {
		slog.Error(ErrDeleteSystemMessage.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrDeleteSystemMessage.Error(),
		})
		return
	}

	// 若删除的消息未读，需要减去计数
	if !message.IsRead {
		ctx := context.Background()
		key := fmt.Sprintf(constants.KeyUnreadMsgCount, message.UserEmail)
		dao.RedisClient.Decr(ctx, key)
	}

	c.JSON(http.StatusOK, response.Response{})
}

func GetUnreadSystemMessageCount(c *gin.Context) {
	ctx := context.Background()
	email := c.GetString("email")
	key := fmt.Sprintf(constants.KeyUnreadMsgCount, email)

	count, err := dao.RedisClient.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		slog.Error(ErrGetUnreadSystemMessageCount.Error(), "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{
			Msg: ErrGetUnreadSystemMessageCount.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Data: response.GetUnreadSystemMessageCountResponse{
			Count: count,
		},
	})
}
