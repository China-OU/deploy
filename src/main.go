package main

import (
	_ "routers"
	_ "initial"

	"github.com/astaxie/beego"
	_ "github.com/go-sql-driver/mysql"
	"tasks"
	"github.com/astaxie/beego/toolbox"
	"os"
)

func main() {
	//获取全局panic
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Panic error:", err)
		}
	}()

	// 定时任务
	cron_task := os.Getenv("CRON_TASK")
	if cron_task == "true" {
		// 定时拉取容器平台数据
		caas_data_sync := toolbox.NewTask("caas_data_sync", "0 0 3 * * * ", func() error {
			err := tasks.CassDataSync()
			if err != nil {
				beego.Error("定时任务: caas_data_sync 发生错误:", err.Error())
				return err
			}
			return nil
		})
		toolbox.AddTask("caas_data_sync", caas_data_sync)

		// 定时同步发布单元信息
		unit_data_sync := toolbox.NewTask("unit_data_sync", "0 30 3 * * * ", func() error {
			err := tasks.UnitDataSync()
			if err != nil {
				beego.Error("定时任务: unit_data_sync 发生错误:", err.Error())
				return err
			}
			return nil
		})
		toolbox.AddTask("unit_data_sync", unit_data_sync)

		// 定时清理过期的DB账户信息
		dbInfoClean := toolbox.NewTask("db_info_clean", "0 0 4 * * *", func() error {
			if err := tasks.DBInfoClean(); err != nil {
				beego.Error("定时任务：db_info_clean 发生错误：", err.Error())
				return err
			}
			return nil
		})
		toolbox.AddTask("db_info_clean", dbInfoClean)

		// 定时清理数据库内容，防止log表过大
		sql_detail_clean := toolbox.NewTask("sql_detail_clean", "0 30 4 * * *", func() error {
			if err := tasks.SqlDetailClean(); err != nil {
				beego.Error("定时任务：sql_detail_clean 发生错误：", err.Error())
				return err
			}
			return nil
		})
		toolbox.AddTask("sql_detail_clean", sql_detail_clean)
	}

	if beego.BConfig.RunMode != "prod" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		beego.SetStaticPath("/mdeploy/swagger", "swagger")
	}

	toolbox.StartTask()
	defer toolbox.StopTask()
	beego.Run()
}

