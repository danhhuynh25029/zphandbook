package data

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"kratos-example/internal/conf"
	"kratos-example/internal/data/ent"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	db *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	db, err := gorm.Open(mysql.Open(c.Database.Source))
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&ent.User{})
	if err != nil {
		log.NewHelper(logger).Errorf("[data] err %v", err)
	}
	return &Data{
		db: db,
	}, cleanup, nil
}
