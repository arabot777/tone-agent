package app

import (
	"fmt"

	"tone/agent/pkg/common/env"
)

type defaults struct {
	// name
	appName string
}

func getDefaults() defaults {
	d := defaults{
		appName: env.Service(),
	}

	return d
}

func defaultWarnLogMetric(appName string) string {
	return fmt.Sprintf("%s.log.warn.count", appName)
}

func defaultErrorLogMetric(appName string) string {
	return fmt.Sprintf("%s.log.error.count", appName)
}
