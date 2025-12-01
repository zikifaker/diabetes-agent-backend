package main

import (
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/router"
	"diabetes-agent-backend/service/mq"
	"diabetes-agent-backend/service/summarization"
	"log/slog"
	"os"
)

func main() {
	// 设置日志
	setSysLog()

	// 启动对话摘要生成服务
	summarization.SummarizerInstance.Run()

	// 启动MQ服务
	if err := mq.Run(); err != nil {
		slog.Error("Failed to start MQ service", "err", err)
		return
	}
	defer mq.Shutdown()

	// 启动HTTP服务
	r := router.Register()
	if err := r.Run(":" + config.Cfg.Server.Port); err != nil {
		slog.Error("Failed to start HTTP server", "err", err)
	}
}

func setSysLog() {
	var level slog.Leveler
	switch config.Cfg.Server.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}
