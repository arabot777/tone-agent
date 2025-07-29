package env

import (
	"fmt"
	"os"
	"strconv"
)

var (
	isDebug     bool
	hostname, _ = os.Hostname()

	host = os.Getenv("HOST")
	port = 80

	platform = os.Getenv("PLATFORM") //
	service  = os.Getenv("SERVICE")  // 服务名称
	env      = os.Getenv("ENV")
	version  = os.Getenv("VERSION")
	id       = os.Getenv("ID")
)

const (
	EnvProd    = "prod" // 生产环境
	EnvUAT     = "uat"  // uat环境
	EnvTesting = "test" // 测试环境
	EnvDevelop = "dev"  // 本地开发环境
)

func Check() {
	if platform == "" {
		fmt.Printf("%s 初始化失败，请设置环境变量: %s\n", "PLATFORM", "PLATFORM")
		os.Exit(1)
	}
	if service == "" {
		fmt.Printf("%s 初始化失败，请设置环境变量: %s\n", "SERVICE", "SERVICE")
		os.Exit(1)
	}
	if env == "" {
		fmt.Printf("%s 初始化失败，请设置环境变量: %s\n", "ENV", "ENV")
		os.Exit(1)
	}
	if version == "" {
		fmt.Printf("%s 初始化失败，请设置环境变量: %s\n", "VERSION", "VERSION")
		os.Exit(1)
	}

	if host == "" {
		host = "0.0.0.0"
	}

	if pI, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = pI
	}
}

func Host() string {
	return host
}

func Port() int {
	return port
}

func Hostname() string {
	return hostname
}

func Environment() string {
	return env
}

func Version() string {
	return version
}

func ID() string {
	return id
}

func Platform() string {
	return platform
}

func Service() string {
	return service
}

func IsProdEnv() bool {
	return Environment() == EnvProd
}

func IsTestingEnv() bool {
	return Environment() == EnvTesting
}

func IsDevelopEnv() bool {
	return Environment() == EnvDevelop
}

func IsUATEnv() bool {
	return Environment() == EnvUAT
}

func IsDebug() bool {
	return isDebug
}

func EnableDebug() {
	isDebug = true
}

func Env[T int | string | int64 | bool](envName string, defaultVal T, useDefault bool) T {
	var envTmp T
	envStr := os.Getenv(envName)

	switch any(envTmp).(type) {
	case string:
		if envStr == "" && useDefault {
			return defaultVal
		} else if envStr == "" {
			fmt.Printf("ENV %s is empty\n", envName)
			os.Exit(1)
		}
		envTmp = any(envStr).(T)
	case int:
		envI, err := strconv.Atoi(envStr)
		if err != nil && useDefault {
			return defaultVal
		} else if err != nil {
			fmt.Printf("ENV %s is err value: %s\n", envName, envStr)
			os.Exit(1)
		}
		envTmp = any(envI).(T)
	case int64:
		envI, err := strconv.ParseInt(envStr, 10, 64)
		if err != nil && useDefault {
			return defaultVal
		} else if err != nil {
			fmt.Printf("ENV %s is err value: %s\n", envName, envStr)
			os.Exit(1)
		}
		envTmp = any(envI).(T)
	case bool:
		envI, err := strconv.ParseBool(envStr)
		if err == nil {
			return any(envI).(T)
		}

		if (envStr == "" || err != nil) && useDefault {
			return defaultVal
		}

		fmt.Printf("ENV %s is err value: %s\n", envName, envStr)
		os.Exit(1)
	}

	return envTmp
}
