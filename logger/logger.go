package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger

func Init(verbose bool, json bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	if json {
		logger = slog.New(
			slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}),
		)
	} else {
		logger = slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				Level:      level,
				TimeFormat: time.Kitchen,
			}))
	}
	slog.SetDefault(logger)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

type summaryStatement struct {
	level slog.Level
	msg   string
	args  []any
}

var summary = []summaryStatement{}

func AddSummaryError(msg string, args ...any) {
	summary = append(summary, summaryStatement{slog.LevelError, msg, args})
}

func Close() {
	line := []byte("------------\n")

	os.Stderr.Write(line)
	for _, i := range summary {
		logger.Log(context.TODO(), i.level, i.msg, i.args...)
	}
	os.Stderr.Write(line)
}

func init() {
	Init(false, false)
}
