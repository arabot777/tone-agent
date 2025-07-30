package conf

import (
	"tone/agent/pkg/common/config/reader"
)

type BaseConfig struct {
	// default model
	DefaultLLMModel    string
	DefaultLLMEndpoint string
	DefaultLLMSK       string
}

func InitBaseConfig(r *reader.Reader) BaseConfig {
	c := BaseConfig{}
	return c
}
