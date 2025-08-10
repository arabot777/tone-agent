package resource

import (
	"context"
	"tone/agent/internal/api/conf"
	"tone/agent/internal/pkg/infra"

	"tone/agent/pkg/common/logger"
)

func InitResource(ctx context.Context) {
	conf.InitConfig()
	// 初始化日志
	logger.MustInit(ctx)
	infra.InitModel()

}

func Close(ctx context.Context) error {

	// 关闭日志
	logger.LogClose(ctx)
	return nil
}
