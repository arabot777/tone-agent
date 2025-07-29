package crontab

import (
	"fmt"
	"testing"
	"time"

	"tone/agent/pkg/common/signal"
)

var running = false

func TestCron(t *testing.T) {
	crontab := New()
	crontab.AddCommand("0/5 * * * * ?", updateCache())
	crontab.Start()
	signal.Wait()
}

func updateCache() CmdFunc {
	return func() {
		if running {
			return
		}
		running = true
		fmt.Println("start time:", time.Now())
		fmt.Println("hello world")
		time.Sleep(7 * time.Second)
		fmt.Println("hello world end ")
		running = false
	}
}
