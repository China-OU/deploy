package tasks

import (
    "fmt"
    "github.com/astaxie/beego"
    "initial"
    "library/common"
    "models"
    "runtime"
    "time"
)

func DBInfoClean() error {
    //获取panic
    defer func() {
        if pErr := recover(); pErr != nil {
            var buf = make([]byte, 1024)
            c := runtime.Stack(buf, false)
            beego.Error("DBInfoClean 清理DB账户错误:", pErr)
            beego.Error("DBInfoClean 清理DB账户详细信息:", string(buf[0:c]))
        }
    }()

    beego.Info("清理DB账户 Start." + time.Now().Format("2006-01-02 15:04:05"))
    if err := setAccountExpired(); err != nil {
        return err
    }
    beego.Info("清理DB账户 End." + time.Now().Format("2006-01-02 15:04:05"))
    return nil
}

func setAccountExpired() error {
    queryStr := fmt.Sprintf("`expired` = 0 AND `expire_time` <= '%s'", time.Now().Format("2006-01-02 15:04:05"))
    updates := map[string]interface{}{
        "key": common.AesEncrypt("********"),
        "encrypted_pwd": common.AesEncrypt("********"),
        "expired": 1,
        "update_time": time.Now(),
    }
    tx := initial.DB.Begin()
    if err := tx.Model(&models.DBAccount{}).Where(queryStr).Updates(updates).Debug().Error; err != nil {
        tx.Rollback()
        beego.Error(err.Error())
        return err
    }
    tx.Commit()
    return nil
}
