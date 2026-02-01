package main

import (
	"diabetes-agent-server/config"
	"diabetes-agent-server/router"
	"diabetes-agent-server/service/mq"
	"log/slog"
	"os"
)

func main() {
	// 设置日志
	setSysLog()

	// 启动 MQ 服务
	if err := mq.Run(); err != nil {
		slog.Error("Failed to start MQ service", "err", err)
		return
	}
	defer mq.Shutdown()

	// 启动 HTTP 服务
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
