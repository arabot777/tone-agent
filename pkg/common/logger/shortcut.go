package logger

import (
	"context"
	"fmt"
	"os"

	"tone/agent/pkg/common/config/log"
	"tone/agent/pkg/common/constant"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

func InitLogger(conf log.Log) error {
	var err error
	env, ok := os.LookupEnv(constant.KeyEnv)
	if !ok {
		env = constant.EnvDev
	}
	platform, ok := os.LookupEnv(constant.KeyPlatform)
	if !ok {
		panic("no platform config")
	}
	service, ok := os.LookupEnv(constant.KeyService)
	if !ok {
		panic("no service config")
	}
	var opts []otelzap.Option
	// opts = append(opts, otelzap.WithMinLevel(conf.TraceLogMinLevel), otelzap.WithTraceIDField(true))
	opts = append(opts, otelzap.WithMinLevel(conf.TraceLogMinLevel))
	switch env {
	case constant.EnvDev: // 开发环境
		InitStdOutCtxLogger(platform, service, opts...)
	case constant.EnvProd, constant.EnvUat, constant.EnvTest, constant.EnvCanary: // 测试生产环境
		InitCtxLogger(conf, platform, service, opts...)
	}
	return err
}

func Debugf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Debug(fmt.Sprintf(format, v...))
}

func Infof(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Info(fmt.Sprintf(format, v...))
}

func Warnf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Warn(fmt.Sprintf(format, v...))
}

func Errorf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Error(fmt.Sprintf(format, v...))
}

func Fatalf(ctx context.Context, format string, v ...interface{}) {
	CtxLogger(ctx).Fatal(fmt.Sprintf(format, v...))
}

func Close() error {
	_ = ctxLogger.Sync()
	lumberjackLoggerClose()
	return nil
}

func IsInitialized() bool { return ctxLogger != nil }
