package models

type UserLogin struct {
	Id         int       `gorm:"column:id" json:"id"`
	Userid     string    `gorm:"column:userid" json:"userid"`
	UserName   string    `gorm:"column:username" json:"username"`
	Phone      string    `gorm:"column:phone" json:"phone"`
	Email      string    `gorm:"column:email" json:"email"`
	Title      string    `gorm:"column:title" json:"title"`
	Company    string    `gorm:"column:company" json:"company"`
	Center     string    `gorm:"column:center" json:"center"`
	Department     string    `gorm:"column:department" json:"department"`
	InsertTime string    `gorm:"column:insert_time" json:"insert_time"`
}

func (UserLogin) TableName() string {
	return "user_login"
}


type UserToken struct {
	Id       int    `gorm:"column:id"`
	UserId string `gorm:"column:userid"`
	Email    string `gorm:"column:email"`
	TokenMd5    string `gorm:"column:token_md5"`
	Expire   string `gorm:"column:expire"`
	Info     string `gorm:"column:info"`
}

func (UserToken) TableName() string {
	return "user_token"
}


type UserRole struct {
	Id         int       `gorm:"column:id" json:"id"`
	Username   string    `gorm:"column:username" json:"username"`
	Realname   string    `gorm:"column:realname" json:"realname"`
	Email      string    `gorm:"column:email" json:"email"`
	Role       string    `gorm:"column:role" json:"role"`
	InsertTime string    `gorm:"column:insert_time" json:"insert_time"`
	IsDelete   int       `gorm:"column:is_delete" json:"is_delete"`
}

func (UserRole) TableName() string {
	return "user_role"
}