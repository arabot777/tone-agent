package mysql

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"tone/agent/pkg/common/config"
)

var (
	host     string
	port     int = 3306
	user     string
	password string
	db       string
)

func MustInit(ctx context.Context) *Datastore {
	if err := initEnv(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	c, err := Init(&config.Mysql{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DB:       db,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return c
}

func initEnv() error {
	host = os.Getenv("MYSQL_HOST")
	if host == "" {
		return fmt.Errorf("%s 初始化失败，请设置环境变量: %s\n", "MYSQL_HOST", "MYSQL_HOST")
	}

	if p, err := strconv.Atoi(os.Getenv("MYSQL_PORT")); err == nil {
		port = p
	}

	user = os.Getenv("MYSQL_USER")
	if user == "" {
		return fmt.Errorf("%s 初始化失败，请设置环境变量: %s\n", "MYSQL_USER", "MYSQL_USER")
	}
	password = os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		return fmt.Errorf("%s 初始化失败，请设置环境变量: %s\n", "MYSQL_PASSWORD", "MYSQL_PASSWORD")
	}
	db = os.Getenv("MYSQL_DATABASE")
	if db == "" {
		return fmt.Errorf("%s 初始化失败，请设置环境变量: %s\n", "MYSQL_DATABASE", "MYSQL_DATABASE")
	}
	return nil
}
