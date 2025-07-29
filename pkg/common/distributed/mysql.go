package distributed

import (
	"context"
	"fmt"
	"time"

	"tone/agent/pkg/common/config"
	"tone/agent/pkg/common/gin/code"
	"tone/agent/pkg/common/logger"
	"tone/agent/pkg/common/mysql"
	"tone/agent/pkg/common/pkgerror"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var releaseLockTime = time.Date(2018, 11, 29, 0, 0, 0, 0, time.Local)

const mysqlEventPrefix = "platform_service"

type mysqlSelector struct {
	cli          *mysql.DB
	leaseSeconds int
	meta         config.MetaEnv
	election     string
	releaseLock  chan struct{}
}

func newMysqlSelector(d *Distributed) *mysqlSelector {
	ms := &mysqlSelector{
		leaseSeconds: d.leaseSeconds,
		meta:         d.meta,
		cli:          d.mysqlClient,
		releaseLock:  make(chan struct{}),
	}
	ms.election = fmt.Sprintf("%s:%s_%s", mysqlEventPrefix, d.meta.Platform, d.meta.Service)
	ms.initTableOrData()
	return ms
}

func (ms *mysqlSelector) Start() {
	for {
		bf30, _ := time.ParseDuration(fmt.Sprintf("-%ds", ms.leaseSeconds*3))
		permit, err := ms.getLock(context.TODO(), ms.election, time.Now().Add(bf30))
		if !permit || err != nil {
			setMaster(false)
		} else {
			setMaster(true)
			logger.Infof(context.Background(), "%s elect: success", ms.meta.ID)
			for {
				select {
				case <-ms.releaseLock:
					return
				case <-time.After(time.Duration(ms.leaseSeconds) * time.Second):
					if err := ms.Update(context.TODO(), ms.election, time.Now()); err != nil {
						logger.Errorf(context.Background(), "update master election failed: %v", err)
					}
				}
			}
		}
		select {
		case <-ms.releaseLock:
			return
		case <-time.After(time.Duration(ms.leaseSeconds) * time.Second):
			continue
		}
	}
}

func (ms *mysqlSelector) Stop() {
	isMaster := master
	setMaster(false)
	ms.releaseLock <- struct{}{}
	if isMaster {
		_ = ms.Update(context.TODO(), ms.election, releaseLockTime)
	}
}

type lock struct {
	Event      string    `json:"event,omitempty" gorm:"column:event;type:varchar(200);primary_key;"`
	UpdateTime time.Time `json:"update_time,omitempty" gorm:"column:update_time;type:datetime"`
}

// TableName 表名
func (lock) TableName() string {
	return "distributed_lock"
}

type LockStore interface {
	getLock(ctx context.Context, lockEvent string, deadline time.Time) (bool, error)
	BeginTrx(ctx context.Context) (mysql.Transaction, error)
	GetTx(ctx context.Context, trx mysql.Transaction, event string, lastTime time.Time) (*lock, error)
	UpdateTx(ctx context.Context, trx mysql.Transaction, event string, timeNow time.Time) error
	Update(ctx context.Context, event string, timeNow time.Time) error
}

func (ms *mysqlSelector) getLock(ctx context.Context, lockEvent string, deadline time.Time) (bool, error) {
	trx, err := ms.BeginTrx(ctx)
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if err != nil {
		return false, err
	}
	if _, err := ms.GetTx(ctx, trx, lockEvent, deadline); err != nil {
		trx.Rollback()
		return false, err
	}
	err = ms.UpdateTx(ctx, trx, lockEvent, time.Now())
	if err != nil {
		trx.Rollback()
		return false, err
	}
	return true, trx.Commit().Error
}

func (ms *mysqlSelector) initTableOrData() {
	if !ms.cli.Migrator().HasTable(&lock{}) {
		_ = ms.cli.Migrator().CreateTable(&lock{})
	}
	_lock := &lock{}
	if err := ms.cli.First(_lock, "event = ? ", ms.election).Error; err != nil {
		if pkgerror.Is(err, gorm.ErrRecordNotFound) {
			ms.cli.Create(&lock{
				Event:      ms.election,
				UpdateTime: releaseLockTime,
			})
		}
	}
}

func (l *mysqlSelector) BeginTrx(ctx context.Context) (*gorm.DB, error) {
	trx := l.cli.Begin()
	if err := trx.Error; err != nil {
		err = pkgerror.WithCode(code.ErrDatabaseTx, err.Error())
		logger.Errorf(context.Background(), "%+#v", err)
		return nil, err
	}
	return trx, nil
}

func (l *mysqlSelector) GetTx(ctx context.Context, trx *gorm.DB,
	event string, lastTime time.Time,
) (*lock, error) {
	lock := &lock{}
	if err := trx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		First(lock, "event = ? and update_time <= ?", event, lastTime).Error; err != nil {
		logger.Debugf(ctx, "cannot get election lock")
		return nil, err
	}
	return lock, nil
}

func (l *mysqlSelector) UpdateTx(ctx context.Context, trx *gorm.DB, event string, timeNow time.Time) error {
	return trx.WithContext(ctx).Model(&lock{}).Where("event = ?", event).Updates(map[string]interface{}{
		"update_time": timeNow,
	}).Error
}

func (l *mysqlSelector) Update(ctx context.Context, event string, timeNow time.Time) error {
	return l.cli.WithContext(ctx).Model(&lock{}).Where("event = ?", event).Updates(map[string]interface{}{
		"update_time": timeNow,
	}).Error
}
