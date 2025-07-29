package logger

import (
	"fmt"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type logger struct {
	*otelzap.Logger
}

func (l logger) Debugf(format string, v ...interface{}) {
	l.Logger.Debug(fmt.Sprintf(format, v...))
}

func (l logger) Infof(format string, v ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, v...))
}

func (l logger) Warnf(format string, v ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(format, v...))
}

func (l logger) Errorf(format string, v ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, v...))
}

func (l logger) Fatalf(format string, v ...interface{}) {
	l.Logger.Fatal(fmt.Sprintf(format, v...))
}

func (l logger) Close() error {
	return nil
}

var (
	i     *logger
	iOnce sync.Once
)

func GetLogger() Logger {
	iOnce.Do(func() {
		i = &logger{
			Logger: ctxLogger,
		}
	})
	return i
}
