package rest

import (
	"time"

	"tone/agent/pkg/common/env"
)

type defaults struct {
	// name
	Name string

	port         int
	timeout      time.Duration
	readTimeOut  time.Duration
	writeTimeout time.Duration
}

func getDefaults() defaults {
	d := defaults{
		Name:         env.Service(),
		port:         8080,
		timeout:      time.Minute,
		readTimeOut:  15 * time.Second,
		writeTimeout: 15 * time.Second,
	}

	return d
}
