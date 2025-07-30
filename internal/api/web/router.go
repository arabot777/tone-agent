package web

import (
	"net/http"
	"tone/agent/docs"
	"tone/agent/internal/api/web/controller"
	"tone/agent/pkg/common/app/rest"
	"tone/agent/pkg/common/app/rest/middleware"
	"tone/agent/pkg/common/env"

	ginlogger "tone/agent/pkg/common/gin/logger"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func NewRouter() *rest.HTTPBundle {
	return rest.New(
		rest.ReadTimeout(0),
		rest.WriteTimeout(0),
		rest.Timeout(0),
		rest.WithRouter(router()),
		rest.WithoutHTTP2(true),
		rest.Port(env.Port()))
}

func router() http.Handler {
	g := gin.Default()
	if env.IsDevelopEnv() {
		gin.SetMode(gin.DebugMode)
	}
	installMiddleware(g)

	apiV1 := "/api/v1"

	docs.SwaggerInfo.BasePath = apiV1

	v1 := g.Group(apiV1)
	{

		agent := v1.Group("/agent")
		{
			agentController := controller.NewAgentController()
			agent.POST("/ok", agentController.Ok)
			agent.GET("/einoagent/stream", agentController.Einoagent)
			agent.GET("/journal/stream", agentController.Journal)
			agent.GET("/", agentController.WebUI)
			agent.GET("/:file", agentController.WebUIFile)
		}

	}

	// 启用 Swagger UI
	// swagger文档查看路径 http://127.0.0.1:8888/swagger/index.html
	g.GET("swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return g.Handler()
}

func installMiddleware(g *gin.Engine) {

	// 初始化健康检查
	g.Use(middleware.CheckHealth())
	g.Use(corsMiddleware())

	// 记录日志
	g.Use(ginlogger.LogWithWriter())
}
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
