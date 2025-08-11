package web

import (
	"context"
	"strconv"
	"tone/agent/docs"
	"tone/agent/internal/api/web/controller"
	"tone/agent/pkg/common/app/rest"
	"tone/agent/pkg/common/app/rest/middleware"
	"tone/agent/pkg/common/env"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func NewRouter() *rest.HTTPBundle {
	h := server.Default(server.WithHostPorts(":" + strconv.Itoa(env.Port())))
	setupRoutes(h)

	// Create a wrapper to adapt Hertz to the existing HTTP bundle interface
	return rest.New(
		rest.ReadTimeout(0),
		rest.WriteTimeout(0),
		rest.Timeout(0),
		rest.WithHertzServer(h),
		rest.WithoutHTTP2(true),
		rest.Port(env.Port()))
}

func setupRoutes(h *server.Hertz) {
	installMiddleware(h)

	apiV1 := "/api/v1"
	docs.SwaggerInfo.BasePath = apiV1

	agentController := controller.NewAgentController()

	v1 := h.Group(apiV1)
	{
		agent := v1.Group("/agent")
		{
			agent.POST("/ok", agentController.Ok)
			agent.GET("/einoagent/stream", agentController.Einoagent)
			agent.GET("/drawing/stream", agentController.Drawing)
			agent.GET("/researcher/stream", agentController.Researcher)
			agent.GET("/journal/stream", agentController.Journal)
			agent.GET("/", agentController.WebUI)
			agent.GET("/:file", agentController.WebUIFile)
		}
	}

	// deer flow demo接口
	api := h.Group("/api")
	{
		api.POST("/chat/stream", agentController.Drawing)
	}

	// TODO: 启用 Swagger UI - 需要适配 Hertz
	// swagger文档查看路径 http://127.0.0.1:8888/swagger/index.html
	// h.GET("/swagger/*any", swagger.WrapHandler(swaggerFiles.Handler))
}

func installMiddleware(h *server.Hertz) {
	// 初始化健康检查
	h.Use(middleware.CheckHealth())
	h.Use(corsMiddleware())

	// TODO: 添加日志中间件 - 需要适配 Hertz
	// h.Use(hertzlogger.LogWithWriter())
}
func corsMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Response.Header.Set("Access-Control-Allow-Origin", "*")
		c.Response.Header.Set("Access-Control-Allow-Headers", "*")
		c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		c.Response.Header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if string(c.Method()) == "OPTIONS" {
			c.AbortWithStatus(consts.StatusNoContent)
			return
		}
		c.Next(ctx)
	}
}
