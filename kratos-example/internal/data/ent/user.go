package ent

type User struct {
	Id       string `json:"id,omitempty" gorm:"column:id;type:string;primary_key"`
	Username string `json:"username,omitempty" gorm:"column:username;type:varchar(255);"`
	Password string `json:"password,omitempty" gorm:"column:password;type:varchar(255);"`
}
