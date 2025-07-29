package conf

import (
	"sync"
	pkgconf "tone/agent/internal/pkg/conf"
	"tone/agent/pkg/common/config/reader"
)

type Config struct {
	pkgconf.BaseConfig
}

var (
	conf *Config
	once sync.Once
)

func InitConfig() {
	c := &Config{}
	r := reader.New()
	c.BaseConfig = pkgconf.InitBaseConfig(r)

	once.Do(func() {
		conf = c
	})
}

func Get() *Config {
	return conf
}
