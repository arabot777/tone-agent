package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"tone/agent/internal/pkg/service/einoagent"
	"tone/agent/pkg/common/logger"

	"github.com/cloudwego/eino-ext/devops"
)

func main() {
	ctx := context.Background()

	// Init eino devops server
	err := devops.Init(ctx)
	if err != nil {
		logger.Errorf(ctx, "[eino dev] init failed, err=%v", err)
		return
	}

	einoagent.BuildeinoagentAgent(ctx)

	// Blocking process exits
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Exit
	logger.Infof(ctx, "[eino dev] process exiting")
}
