package crontab

import "tone/agent/pkg/common/distributed"

type Option func(*Crontab)

// WithSingle  single point execution.
// If false all service exec, otherwise, only master does it
func WithSingle(single bool) Option {
	return func(a *Crontab) {
		a.single = single
	}
}

func WithDistributed(distributed *distributed.Distributed) Option {
	return func(a *Crontab) {
		a.distributed = distributed
	}
}
