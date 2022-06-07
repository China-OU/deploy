package initial

import (
	"fmt"
	"github.com/astaxie/beego"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"library/common"
	"github.com/jinzhu/gorm"
	"os"
)

var DB *gorm.DB

func InitSql() {
	user := beego.AppConfig.String("mysqluser")
	passwd := beego.AppConfig.String("mysqlpass")
	host := beego.AppConfig.String("mysqlurls")
	port, err := beego.AppConfig.Int("mysqlport")
	dbname := beego.AppConfig.String("mysqldb")
	if nil != err {
		port = 3306
	}
	pwd := common.AesDecrypt(passwd)

	// 采用beego的orm连接
	//if beego.AppConfig.String("runmode") != "prod" {
	//	orm.Debug = true
	//}
	//orm.RegisterDriver("mysql", orm.DRMySQL)
	//orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pwd, host, port, dbname))

	// 采用gorm连接数据库
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		user, pwd, host, port, dbname))
	if err != nil {
		beego.Error(err.Error())
		os.Exit(-1)
	} else {
		beego.Info("Database connected")
	}
	if beego.AppConfig.String("runmode") != "prd" {
		db.LogMode(true)
	}
	db_max_idle_conn, _ := beego.AppConfig.Int("db_max_idle_conn")
	db_max_open_conn, _ := beego.AppConfig.Int("db_max_open_conn")
	db.DB().SetMaxIdleConns(db_max_idle_conn)
	db.DB().SetMaxOpenConns(db_max_open_conn)
	DB = db
}
