package ent

type Users struct {
	Id     string `gorm:"column:id;primaryKey;type:varchar(36)"`
	Name   string `gorm:"column:name;type:varchar(255)"`
	Age    int    `gorm:"column:age;type:int(11)"`
	Gender int64  `gorm:"column:gender;type:int(1)"`
}

func (Users) TableName() string {
	return "users"
}
