package app

type appConfig struct {
	// log monitor
	warnMetric  string
	errorMetric string

	// pprof
	profilerPort *int

	enableConfig bool
}
