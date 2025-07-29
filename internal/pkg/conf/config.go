package conf

import (
	"tone/agent/pkg/common/config/reader"
)

type BaseConfig struct {
}

type InitConfig struct {
	DefaultOrganizationName string
	DefaultAdminUsername    string
	DefaultAdminPassword    string
}

func InitBaseConfig(r *reader.Reader) BaseConfig {
	c := BaseConfig{}
	return c
}
