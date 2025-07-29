package controller

import (
	"net/http"
	"tone/agent/internal/api/service"

	"github.com/gin-gonic/gin"
)

type AgentController struct {
	agentService *service.AgentService
}

func NewAgentController() *AgentController {
	return &AgentController{
		agentService: service.NewAgentService(),
	}
}

func (c *AgentController) Ok(g *gin.Context) {
	g.JSON(http.StatusOK, c.agentService.Ok())
}

func (c *AgentController) Drawing(g *gin.Context) {
	// TODO SSE 返回

}
