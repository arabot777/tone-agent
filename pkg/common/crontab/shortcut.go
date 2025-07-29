package crontab

import (
	"context"

	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/common/signal"
)

var crontab *Crontab

func Init(opts ...Option) {
	crontab = New(opts...)
	crontab.Start()
	logger.Infof(context.Background(), "crontab task is running...")
}

func Close() signal.ExitFunc {
	return func() {
		crontab.Stop()
		logger.Infof(context.Background(), "crontab task is stop...")
	}
}

func AddCommand(spec string, cronFunc CmdFunc) {
	crontab.AddCommand(spec, cronFunc)
}

// AddCommandSkipped 如果当前任务处在运行中，则下次调度跳过
// jobName 不允许重复
// jobName 为空则任务不跳过
func AddCommandSkipped(spec string, jobName string, cronFunc CmdFunc) {
	if jobName == "" {
		crontab.AddCommand(spec, cronFunc)
	} else {
		crontab.AddCommandSkipped(spec, jobName, cronFunc)
	}
}
