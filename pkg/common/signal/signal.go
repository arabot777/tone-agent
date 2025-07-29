package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type ExitFunc func()

func Wait(exitFunc ...ExitFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-ch
		fmt.Printf("signal: service got signal: %s!\n", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			for _, f := range exitFunc {
				f()
			}
			fmt.Println("signal: service exit now!")
			os.Exit(0)
		case syscall.SIGHUP:
			fmt.Println("signal: got signal hup!")
		default:
			fmt.Printf("signal: got unknown signal %s!\n", sig.String())
			os.Exit(0)
		}
	}
}
