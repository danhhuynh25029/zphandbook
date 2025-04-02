package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"kratos-example/internal/conf"
	"kratos-example/internal/data/ent"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewUserRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	db    *gorm.DB
	redis *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	conn, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = conn.AutoMigrate(&ent.Users{})
	if err != nil {
		panic(err)
	}
	db, err := conn.DB()
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: "",
	})
	if err := redisClient.Ping().Err(); err != nil {
		panic(err)
	}
	cleanup := func() {
		db.Close()
		redisClient.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db: conn,
	}, cleanup, nil
}
