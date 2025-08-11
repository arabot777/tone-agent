package conf

import (
	"encoding/json"
	"os"
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

	// init mcp
	initMCPConfig(&c)

	return c
}

// initMCPConfig initializes MCP configuration from environment variables
func initMCPConfig(c *BaseConfig) {
	// Get MCP_SERVER environment variable
	mcpServerEnv := os.Getenv("MCP_SERVER")
	if mcpServerEnv == "" {
		return
	}

	// Parse JSON structure
	var envConfig struct {
		McpServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env,omitempty"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal([]byte(mcpServerEnv), &envConfig); err != nil {
		// If direct parsing fails, try to extract from nested structure
		var nestedConfig map[string]interface{}
		if err := json.Unmarshal([]byte(mcpServerEnv), &nestedConfig); err != nil {
			return
		}
		
		// Extract mcpServers from nested structure
		if mcpServersData, ok := nestedConfig["mcpServers"]; ok {
			if mcpServersBytes, err := json.Marshal(mcpServersData); err == nil {
				if err := json.Unmarshal(mcpServersBytes, &envConfig.McpServers); err != nil {
					return
				}
			}
		}
	}

	// Initialize MCP servers map
	c.MCP.Servers = make(map[string]struct {
		Command string            `yaml:"command"`
		Args    []string          `yaml:"args"`
		Env     map[string]string `yaml:"env,omitempty"`
	})

	// Copy configuration
	for name, server := range envConfig.McpServers {
		c.MCP.Servers[name] = struct {
			Command string            `yaml:"command"`
			Args    []string          `yaml:"args"`
			Env     map[string]string `yaml:"env,omitempty"`
		}{
			Command: server.Command,
			Args:    server.Args,
			Env:     server.Env,
		}
	}
}
