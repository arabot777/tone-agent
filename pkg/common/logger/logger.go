package logger

import (
	"context"
	"io"
	"time"

	gl "gorm.io/gorm/logger"
)

type Logger interface {
	FormatLogger
	io.Closer
}

type FormatLogger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

const AbbrMysql = "db.mysql"

type DBLogger struct {
	LogLevel gl.LogLevel
	Logger
}

func (l *DBLogger) LogMode(lv gl.LogLevel) gl.Interface {
	newlogger := *l
	newlogger.LogLevel = lv
	return &newlogger
}

func (l *DBLogger) Info(ctx context.Context, msg string, v ...interface{}) {
	if l.LogLevel <= gl.Info {
		l.Logger.Infof(AbbrMysql+" %s: %s\n "+msg, v...)
	}
}

func (l *DBLogger) Warn(ctx context.Context, msg string, v ...interface{}) {
	if l.LogLevel <= gl.Warn {
		l.Logger.Warnf(AbbrMysql+" %s: %s\n "+msg, v...)
	}
}

func (l *DBLogger) Error(ctx context.Context, msg string, v ...interface{}) {
	if l.LogLevel <= gl.Error {
		l.Logger.Errorf(AbbrMysql+" %s: %s\n "+msg, v...)
	}
}

func (l *DBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gl.Info {
		// ignore trace now
	}
}
