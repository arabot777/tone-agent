package service

type AgentService struct {
}

func NewAgentService() *AgentService {
	return &AgentService{}
}

func (s *AgentService) Ok() string {
	return "agent"
}
