package mysql

import (
	"context"
	"fmt"
	"time"

	log "github.com/fxlwell/golang-log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Conf struct {
	Addr       string
	Username   string
	Password   string
	Database   string
	DsnOptions string

	MaxIdle     int
	MaxOpen     int
	MaxLifeTime time.Duration

	SlowTime   time.Duration
	SlowLogger string
}

var (
	mysqls map[string]*gorm.DB
)

func init() {
	mysqls = make(map[string]*gorm.DB)
}

func Init(ctx context.Context, configs map[string]*Conf) error {
	for sn, conf := range configs {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", conf.Username, conf.Password, conf.Addr, conf.Database)
		if conf.DsnOptions != "" {
			dsn = dsn + "?" + conf.DsnOptions
		}

		gormConfig := &gorm.Config{
			Logger: logger.New(log.DevNull, logger.Config{
				SlowThreshold:             2000 * time.Millisecond,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			}),
		}

		if conf.SlowLogger != "" && conf.SlowTime > 0 {
			gormConfig.Logger = logger.New(log.Get(conf.SlowLogger), logger.Config{
				SlowThreshold:             conf.SlowTime,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: false,
				Colorful:                  false,
			})
		}

		gdb, err := gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			log.Get("error").Fatal("mysql:", err, dsn, sn)
			return err
		}

		if sqlDb, err := gdb.DB(); err != nil {
			log.Get("error").Fatal("mysql:", err, dsn, sn)
			return err
		} else {
			sqlDb.SetMaxIdleConns(conf.MaxIdle)
			sqlDb.SetMaxOpenConns(conf.MaxOpen)
			sqlDb.SetConnMaxLifetime(conf.MaxLifeTime)

			// stats monitor
			go func(ctx context.Context, node string) {
				// type DBStats struct {
				//  MaxOpenConnections int // Maximum number of open connections to the database.

				// Pool Status
				//  OpenConnections int // The number of established connections both in use and idle.
				//  InUse           int // The number of connections currently in use.
				//  Idle            int // The number of idle connections.

				// Counters
				//  WaitCount         int64         // The total number of connections waited for.
				//  WaitDuration      time.Duration // The total time blocked waiting for a new connection.
				//  MaxIdleClosed     int64         // The total number of connections closed due to SetMaxIdleConns.
				//  MaxIdleTimeClosed int64         // The total number of connections closed due to SetConnMaxIdleTime.
				//  MaxLifetimeClosed int64         // The total number of connections closed due to SetConnMaxLifetime.
				// }
				t := time.NewTicker(time.Second * 20)
				defer t.Stop()
				for {

					select {
					case <-ctx.Done():
						return
					case <-t.C:
						stat := sqlDb.Stats()
						infos := fmt.Sprintf("mysql stat: Connection open:%d, inUse:%d, idle:%d, waitCount:%d, waitDuration:%v dsn:%s",
							stat.OpenConnections,
							stat.InUse,
							stat.Idle,
							stat.WaitCount,
							stat.WaitDuration,
							node)
						log.Get("run").Infof(infos)
					}
				}
			}(ctx, sn)
		}

		mysqls[sn] = gdb
	}

	return nil
}

func Get(sn string) *gorm.DB {
	if g, ok := mysqls[sn]; ok {
		return g
	}
	log.Get("error").Fatalf(fmt.Sprintf("mysql node \"%s\" not exists", sn))
	panic(fmt.Sprintf("mysql: not exists node:%s", sn))
	return nil
}
