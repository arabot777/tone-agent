package distributed

import (
	"context"

	"tone/agent/pkg/common/config"
	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/common/signal"
)

var distributed *Distributed

func Init(c config.MetaEnv, opts ...Option) *Distributed {
	distributed = New(c, opts...)
	distributed.Start()
	logger.Infof(context.Background(), "distributed select master is running...")
	return distributed
}

func Close() signal.ExitFunc {
	return func() {
		distributed.Stop()
		logger.Infof(context.Background(), "distributed selection service stopped")
	}
}

func IsMaster() bool {
	return distributed.Master()
}
