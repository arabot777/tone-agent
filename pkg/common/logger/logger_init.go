package logger

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap/zapcore"

	"tone/agent/pkg/common/config/log"
	"tone/agent/pkg/common/env"
)

var (
	path             string
	maxSize                        = 50
	maxBackups                     = 10
	maxAge                         = 7
	compress                       = false
	traceLogMinLevel zapcore.Level = zapcore.InfoLevel
)

func MustInit(_ context.Context) {
	if err := initEnv(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := InitLogger(log.Log{
		Path:             path,
		MaxSize:          maxSize,
		MaxBackups:       maxBackups,
		MaxAge:           maxAge,
		Compress:         compress,
		TraceLogMinLevel: traceLogMinLevel,
	}); err != nil {
		fmt.Println("init log err")
		os.Exit(1)
	}
}

func initEnv() error {
	path = os.Getenv("LOG_PATH")
	if path == "" {
		return fmt.Errorf("%s 初始化失败，请设置环境变量: %s\n", "LOG_PATH", "LOG_PATH")
	}

	if ms, err := strconv.Atoi(os.Getenv("LOG_MAX_SIZE")); err == nil {
		maxSize = ms
	}

	if mb, err := strconv.Atoi(os.Getenv("LOG_MAX_BACKUPS")); err == nil {
		maxBackups = mb
	}
	if ma, err := strconv.Atoi(os.Getenv("LOG_MAX_AGE")); err == nil {
		maxAge = ma
	}
	if c, err := strconv.ParseBool(os.Getenv("LOG_COMPRESS")); err == nil {
		compress = c
	}
	if l, err := strconv.Atoi(os.Getenv("LOG_TRACE_LOG_MIN_LEVEL")); err == nil {
		traceLogMinLevel = zapcore.Level(l)
	}
	return nil
}

func LogClose(_ context.Context) {
	if env.IsDevelopEnv() {
		return
	}
	_ = Close()
}
