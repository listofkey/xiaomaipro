package svc

import (
	"context"
	"time"

	"server/app/rpc/dao"
	"server/app/rpc/user/internal/config"
	"server/app/rpc/user/internal/pkg/encrypt"
	"server/pkg/logging"
	"server/pkg/monitoring"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Query  *dao.Query
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	if err := encrypt.ValidateAESKey(c.AES.Key); err != nil {
		panic("invalid AES key config: " + err.Error())
	}

	db, err := gorm.Open(postgres.Open(c.DB.DSN), &gorm.Config{
		Logger: logging.NewGormLogger("gorm"),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	monitoring.StartDBMonitor("user", "postgres", db)

	dao.SetDefault(db)
	q := dao.Use(db)

	var rdb *redis.Client
	if c.RedisConfig.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     c.RedisConfig.Host,
			Password: c.RedisConfig.Password,
			DB:       c.RedisConfig.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if _, err := rdb.Ping(ctx).Result(); err != nil {
			logx.Errorf("redis connect failed, continue without token blacklist: %v", err)
			rdb = nil
		} else {
			logx.Infof("redis connected: %s", c.RedisConfig.Host)
			monitoring.InstrumentRedis("user", c.RedisConfig.Host, rdb)
		}
	}
	if rdb == nil && c.RedisConfig.Host != "" {
		monitoring.StartTCPMonitor("user", "redis", c.RedisConfig.Host, 0)
	}
	for _, target := range c.Etcd.Hosts {
		monitoring.StartTCPMonitor("user", "etcd", target, 0)
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Query:  q,
		Redis:  rdb,
	}
}
