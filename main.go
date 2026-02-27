package main

import (
	"diabetes-agent-server/config"
	"diabetes-agent-server/router"
	healthreport "diabetes-agent-server/service/health-weekly-report"
	"diabetes-agent-server/service/mq"
	"log/slog"
	"os"
	"strings"
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

	// 启动健康周报定时任务
	go healthreport.SetupHealthWeeklyReportScheduler()

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
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(a.Value.Time().Format("2006/01/02 - 15:04:05"))
			} else if a.Key == slog.LevelKey {
				a.Value = slog.StringValue(strings.ToUpper(a.Value.String()))
			}
			return a
		},
	})))
}
