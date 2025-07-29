package container

import (
	"context"
	"sync"

	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/common/utils"
)

type Bundle interface {
	Type() string
	Name() string
	Run(ctx context.Context) error
	Stop() context.Context
}

type Container struct {
	bundles  []Bundle
	bundleWg sync.WaitGroup
}

func New() *Container {
	return &Container{}
}

func bundleDesc(b Bundle) string {
	return b.Type() + "[" + b.Name() + "]"
}

func (c *Container) AddBundle(bundles ...Bundle) {
	c.bundles = append(c.bundles, bundles...)
}

func (c *Container) StartAll(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	for _, b := range c.bundles {
		bundle := b
		logger.Infof(ctx, "Start bundle:%s", bundleDesc(bundle))
		c.bundleWg.Add(1)
		go func() {
			defer c.bundleWg.Done()
			if err := utils.SafelyRun(func() { _ = bundle.Run(ctx) }); err != nil {
				logger.Errorf(
					ctx,
					"Run bundle:%s failed error: %s",
					bundleDesc(bundle),
					err.Error(),
				)
			}
		}()
		logger.Infof(ctx, "Bundle started: %s", bundleDesc(bundle))
	}

	go func() {
		c.bundleWg.Wait()
		cancel()
	}()

	return ctx
}

func (c *Container) StopAll(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	var eg utils.ErrorGroup
	for _, b := range c.bundles {
		bundle := b
		logger.Infof(ctx, "Stop bundle: %s", bundleDesc(bundle))
		eg.Go(func() error {
			stopCtx := bundle.Stop()
			<-stopCtx.Done()
			logger.Infof(ctx, "Bundle stopped: %s", bundleDesc(bundle))
			return nil
		})
	}

	go func() {
		if err := eg.Wait(); err != nil {
			logger.Errorf(ctx, "Stop bundel error: %s", err.Error())
		}
		cancel()
		logger.Infof(ctx, "All bundle stopped")
	}()

	return ctx
}
