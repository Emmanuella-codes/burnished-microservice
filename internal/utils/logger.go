package utils

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *slog.Logger

func InitLogger() {
	rotatingFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    20,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	writer := io.MultiWriter(os.Stdout, rotatingFile)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	Logger = slog.New(handler)
}

func LogInfo(msg string, args ...any) {
	Logger.Info(msg, args...)
}

func LogError(msg string, err error, args ...any) {
	allArgs := append([]any{"error", err}, args...)
	Logger.Error(msg, allArgs...)
}

func LogWarn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}
