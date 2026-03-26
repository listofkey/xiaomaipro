package svc

import (
	"context"
	"fmt"
	"time"

	"server/app/rpc/dao"
	"server/app/rpc/program/internal/config"
	"server/pkg/monitoring"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config config.Config
	DB     *gorm.DB
	Query  *dao.Query
	Redis  *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(postgres.Open(c.DB.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	monitoring.StartDBMonitor("program", "postgres", db)

	dao.SetDefault(db)
	q := dao.Use(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConfig.Host,
		Password: c.RedisConfig.Password,
		DB:       c.RedisConfig.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logx.Errorf("redis ping failed, cache will be best-effort only: %v", err)
	} else {
		logx.Infof("redis connected: %s", c.RedisConfig.Host)
	}
	monitoring.InstrumentRedis("program", c.RedisConfig.Host, rdb)
	for _, target := range c.Etcd.Hosts {
		monitoring.StartTCPMonitor("program", "etcd", target, 0)
	}

	return &ServiceContext{
		Config: c,
		DB:     db,
		Query:  q,
		Redis:  rdb,
	}
}

const (
	PrefixEventDetail  = "prog:event:detail:"
	PrefixEventList    = "prog:event:list:"
	PrefixHotRecommend = "prog:hot:"
	PrefixCategoryList = "prog:category:list"
	PrefixEventSearch  = "prog:search:"
)

func EventDetailKey(eventID int64) string {
	return fmt.Sprintf("%s%d", PrefixEventDetail, eventID)
}

func HotRecommendKey(city string) string {
	if city == "" {
		city = "all"
	}
	return fmt.Sprintf("%s%s", PrefixHotRecommend, city)
}
