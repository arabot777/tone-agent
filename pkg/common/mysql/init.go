package mysql

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"gorm.io/plugin/opentelemetry/tracing"

	"tone/agent/pkg/common/config"
	"tone/agent/pkg/common/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
)

/*var (
	once    sync.Once
	Factory *Datastore
)*/

type DB struct {
	gorm.DB
}

type SQLTimeout struct {
	TimeoutExec  time.Duration
	TimeoutQuery time.Duration
	TimeoutTrx   time.Duration
}

type Config struct {
	MDSN string // mysql实例的dsn
	// SDSN   []string // slaves，预留的, for more infos , refer https://github.com/go-gorm/dbresolver
	// TODO
	Active  int
	Idle    int
	Timeout SQLTimeout
	Conns
	Logger logger.Logger
}

type Conns struct {
	TimeoutIdle time.Duration
	MaxIdleConn int
	MaxOpenConn int
}

// type Transaction = *gorm.DB
func Init(conf *config.Mysql) (*Datastore, error) {
	var (
		err   error
		dbIns *DB
	)
	// once.Do(func() {
	dbIns, err = new(logger.GetLogger(), conf.Host, conf.Port, conf.User, conf.Password, conf.DB)
	if err != nil {
		fmt.Println(fmt.Errorf("fail to new mysql: %w", err))
		return nil, err
	}
	if err = dbIns.Use(tracing.NewPlugin(tracing.WithAttributes(
		attribute.String("db.connection_string", fmt.Sprintf("Server=%s;Port=%d,Database=%s;Uid=%s",
			conf.Host, conf.Port, conf.DB, conf.User)),
	), tracing.WithoutMetrics())); err != nil {
		panic(err)
	}
	//Factory = &Datastore{db: &dbIns.DB}
	//})
	/*if Factory == nil || err != nil {
		return nil, fmt.Errorf("failed to get mysql store fatory, mysqlFactory: %+v, error: %w", Factory, err)
	}*/

	return &Datastore{db: &dbIns.DB}, nil
}

func new(log logger.Logger, host string, port int, username, password, db string) (*DB, error) {
	c := &Config{
		MDSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, db),
		Timeout: SQLTimeout{
			TimeoutExec:  time.Second * 3,
			TimeoutQuery: time.Second * 3,
			TimeoutTrx:   time.Second * 3,
		},
		Logger: log,
	}
	return newInstances(c)
}

func newInstances(c *Config) (*DB, error) {
	lg := &logger.DBLogger{
		LogLevel: gl.Info,
		Logger:   c.Logger,
	}

	if c.MDSN == "" {
		return nil, errors.New("mysql address can't be null")
	}

	if c.Timeout.TimeoutExec == 0 || c.Timeout.TimeoutQuery == 0 || c.Timeout.TimeoutTrx == 0 {
		return nil, errors.New("mysql timeout can't be 0")
	}

	db, err := gorm.Open(mysql.Open(c.MDSN), &gorm.Config{
		Logger: lg,
	})
	if err != nil {
		lg.Fatalf(fmt.Sprintf("mysql: can't establish connection with mysql: %s", c.MDSN), err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		lg.Fatalf("mysql: can't select db in mysql", err)
		return nil, err
	}

	// TODO 调整默认链接池
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	if c.Conns.MaxIdleConn > 0 {
		sqlDB.SetMaxIdleConns(c.Conns.MaxIdleConn)
	} else {
		sqlDB.SetMaxIdleConns(5)
	}

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	if c.Conns.MaxOpenConn > 0 {
		sqlDB.SetMaxOpenConns(c.Conns.MaxOpenConn)
	} else {
		sqlDB.SetMaxOpenConns(10)
	}

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	if c.Conns.TimeoutIdle != 0 {
		sqlDB.SetConnMaxLifetime(c.Conns.TimeoutIdle)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// InitHooks(db)
	return &DB{*db}, nil
}
