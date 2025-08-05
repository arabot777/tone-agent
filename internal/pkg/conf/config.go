package conf

import (
	"tone/agent/pkg/common/config/reader"
)

type BaseConfig struct {
	// default model
	DefaultLLMModel    string
	DefaultLLMEndpoint string
	DefaultLLMSK       string

	MCP struct {
		Servers map[string]struct {
			Command string            `yaml:"command"`
			Args    []string          `yaml:"args"`
			Env     map[string]string `yaml:"env,omitempty"`
		} `yaml:"servers"`
	} `yaml:"mcp"`
}

func InitBaseConfig(r *reader.Reader) BaseConfig {
	c := BaseConfig{}

	return c
}
