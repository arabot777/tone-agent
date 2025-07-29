package distributed

import (
	"tone/agent/pkg/common/mysql"
)

type Option func(*Distributed)

// WithMode  base db mode
func WithMode(mode string) Option {
	return func(d *Distributed) {
		d.mode = mode
	}
}

// WithMysql  etcd client
func WithMysql(cli *mysql.DB) Option {
	return func(d *Distributed) {
		d.mysqlClient = cli
	}
}

// WithLeaseTime  master lease time, unit is seconds
func WithLeaseTime(leaseSeconds int) Option {
	return func(d *Distributed) {
		d.leaseSeconds = leaseSeconds
	}
}
