package crontab

import (
	"context"
	"sync"

	"tone/agent/pkg/common/distributed"
	"tone/agent/pkg/common/logger"

	"github.com/robfig/cron"
)

type CmdFunc func()

type ICron interface {
	AddCommand(spec string, cronFunc CmdFunc)
	Start()
	Stop()
}

type Crontab struct {
	crontab     *cron.Cron
	single      bool
	distributed *distributed.Distributed
	jobName     sync.Map
}

func New(opts ...Option) *Crontab {
	c := &Crontab{
		crontab: cron.New(),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Crontab) AddCommand(spec string, cronFunc CmdFunc) {
	if err := c.crontab.AddFunc(spec, func() {
		if c.single {
			if c.distributed.Master() {
				cronFunc()
			}
		} else {
			cronFunc()
		}
	}); err != nil {
		logger.Errorf(context.Background(), "add command err: %+#v", err)
	}
}

func (c *Crontab) AddCommandSkipped(spec string, jobName string, cronFunc CmdFunc) {
	c.jobName.Store(jobName, false)
	if err := c.crontab.AddFunc(spec, func() {
		if running, _ := c.jobName.Load(jobName); running.(bool) {
			return
		}
		c.jobName.Store(jobName, true)
		if c.single {
			if c.distributed.Master() {
				cronFunc()
			}
		} else {
			cronFunc()
		}
		c.jobName.Store(jobName, false)
	}); err != nil {
		logger.Errorf(context.Background(), "add command err: %+#v", err)
	}
}

func (c *Crontab) Start() {
	c.crontab.Start()
}

func (c *Crontab) Stop() {
	c.crontab.Stop()
}
