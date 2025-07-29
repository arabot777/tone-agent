package distributed

import (
	"tone/agent/pkg/common/config"
	"tone/agent/pkg/common/mysql"
)

const (
	MysqlMode = "mysql"
)

var (
	defaultLeaseSeconds = 30
	master              = false
)

type IDistributedSelector interface {
	// Start 启动分布式选择器, 所有节点对某个服务进行抢占，只选择其中一个节点
	Start()
	Stop()
}

type IDistributed interface {
	Master() bool
}

type Distributed struct {
	mode          string
	leaseSeconds  int
	meta          config.MetaEnv
	mysqlClient   *mysql.DB
	mysqlSelector *mysqlSelector
}

func New(m config.MetaEnv, opts ...Option) *Distributed {
	d := &Distributed{mode: MysqlMode, leaseSeconds: defaultLeaseSeconds, meta: m}
	for _, o := range opts {
		o(d)
	}
	return d
}

func (d *Distributed) Start() {
	switch d.mode {
	case MysqlMode:
		d.mysqlSelector = newMysqlSelector(d)
		go d.mysqlSelector.Start()
	default:
		panic("no corresponding mode")
	}
}

func (d *Distributed) Stop() {
	switch d.mode {
	case MysqlMode:
		d.mysqlSelector.Stop()
	}
}

func (d *Distributed) Master() bool {
	return master
}

func setMaster(res bool) {
	master = res
}
