package tasks

import (
    "github.com/astaxie/beego"
    "initial"
    "models"
    "runtime"
    "time"
)

func SqlDetailClean() error {
    //获取panic
    defer func() {
        if panic_err := recover(); panic_err != nil {
            var buf = make([]byte, 1024)
            c := runtime.Stack(buf, false)
            beego.Error("SqlDetailClean 清理脚本内容错误:", panic_err)
            beego.Error("SqlDetailClean 清理脚本内容详细信息:", string(buf[0:c]))
        }
    }()

    beego.Info("清理脚本内容 Start." + time.Now().Format(initial.DatetimeFormat))
    if err := clear_sql_content(); err != nil {
        return err
    }
    beego.Info("清理脚本内容 End." + time.Now().Format(initial.DatetimeFormat))
    return nil
}

func clear_sql_content() error {
    one_month := time.Now().AddDate(0, -1, 0).Format(initial.DatetimeFormat)
    tx := initial.DB.Begin()
    err := tx.Model(models.OnlineDbLog{}).Where("start_time<?", one_month).Update("file_content", "").Error
    if err != nil {
        beego.Error(err.Error())
        tx.Rollback()
        return err
    }
    tx.Commit()
    return nil
}
