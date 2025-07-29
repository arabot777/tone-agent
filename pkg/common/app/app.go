package app

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tone/agent/pkg/common/app/internal/container"
	"tone/agent/pkg/common/app/internal/util"
	"tone/agent/pkg/common/env"
	"tone/agent/pkg/common/logger"
	l "tone/agent/pkg/common/logger"
)

type Bundle = container.Bundle

type Application interface {
	Name() string
	Run()
	AddBundle(bundles ...Bundle)
}

type BaseApplication struct {
	container.Container

	name string
	ctx  context.Context

	// config
	config appConfig

	// lifecycle hooks
	beforeStart []func(ctx context.Context) error
	afterStart  []func(ctx context.Context) error
	beforeStop  []func(ctx context.Context) error
	afterStop   []func(ctx context.Context) error
}

type appContextKeyType string

const appContextKey appContextKeyType = "app-context"

func NewApplication(opts ...Option) *BaseApplication {
	env.Check()
	defaults := getDefaults()
	customOptions := &options{}
	for _, opt := range opts {
		opt(customOptions)
	}

	ctx := util.DerefCtx(customOptions.ctx, context.Background())
	appName := util.DerefString(customOptions.appName, defaults.appName)

	app := &BaseApplication{
		Container: *container.New(),
		name:      appName,
		config: appConfig{
			warnMetric:   defaultWarnLogMetric(appName),
			errorMetric:  defaultErrorLogMetric(appName),
			profilerPort: customOptions.profilerPort,
			enableConfig: customOptions.withConfig,
		},
		ctx: ctx,

		// hooks
		beforeStart: customOptions.beforeStart,
		afterStart:  customOptions.afterStart,
		beforeStop:  customOptions.beforeStop,
		afterStop:   customOptions.afterStop,
	}

	app.initLog(ctx)

	return app
}

func AppFromContext(ctx context.Context) Application {
	v := ctx.Value(appContextKey)
	if v != nil {
		if app, ok := v.(Application); ok {
			return app
		}
	}
	logger.Errorf(ctx, "Not a valid cafe context")
	return nil
}

func (app *BaseApplication) Name() string {
	return app.name
}

func (app *BaseApplication) runBeforeStart() {
	if err := util.RunUntilError(app.ctx, app.beforeStart); err != nil {
		logger.Errorf(app.ctx, "Error run before start hook:%s", err.Error())
	}
}

func (app *BaseApplication) runAfterStart() {
	if err := util.RunUntilError(app.ctx, app.afterStart); err != nil {
		logger.Errorf(app.ctx, "Error run after start hook:%s", err.Error())
	}
}

func (app *BaseApplication) runBeforeStop() {
	if err := util.RunUntilError(app.ctx, app.beforeStop); err != nil {
		logger.Errorf(app.ctx, "Error run before stopAll hook:%s", err.Error())
	}
}

func (app *BaseApplication) runAfterStop() {
	if err := util.RunUntilError(app.ctx, app.afterStop); err != nil {
		logger.Errorf(app.ctx, "Error run after stopAll hook:%s", err.Error())
	}
}

func (app *BaseApplication) initLog(ctx context.Context) {
	l.MustInit(ctx)
}

func (app *BaseApplication) Run() {
	logger.Infof(app.ctx, "Run cafe application,name=%s", app.name)

	// sentry.SetIncludePaths(app.config.includePaths)

	// 这个时候才知道应用的 application 对象是什么
	app.ctx = context.WithValue(app.ctx, appContextKey, app)

	if app.config.profilerPort != nil {
		go func() {
			err := http.ListenAndServe(fmt.Sprintf(":%d", *app.config.profilerPort), nil)
			if err != nil {
				logger.Errorf(app.ctx, "Start pprof error:%s", err.Error())
			}
		}()
	}

	// start all
	app.runBeforeStart()

	finishCtx := app.StartAll(app.ctx)

	app.runAfterStart()

	// wait for shutdown or done
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	select {
	case <-finishCtx.Done():
		logger.Infof(app.ctx, "All bundle finished!")
	case <-shutdownSignal:
		logger.Infof(app.ctx, "Shutdown signal received")
	}

	// stop all
	app.runBeforeStop()

	ctx := app.StopAll(app.ctx)

	shutdownTimeout := time.After(30 * time.Second)
	select {
	case <-ctx.Done():
		logger.Infof(app.ctx, "Application stopped")
	case <-shutdownTimeout:
		logger.Infof(app.ctx, "Shutdown timeout, force stop application")
	}

	app.runAfterStop()

	logger.Infof(app.ctx, "Bye!")
}
