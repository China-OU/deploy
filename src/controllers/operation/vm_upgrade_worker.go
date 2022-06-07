package operation

import (
    "fmt"
    "github.com/astaxie/beego"
    "initial"
    "library/common"
    "models"
    "strings"
    "time"
)

func (u *UnitVMUpgrade) Do()  {
    timeout := time.After(10 * time.Minute)
    resultChan := make(chan bool, 1)
    go func() {
        defer func() {
            if err := recover(); err != nil {
                errMsg := fmt.Sprintf("VM Upgrade Panic error: %s", err)
                beego.Error(errMsg)
                u.Logs = append(u.Logs, errMsg)
                u.SaveResult(0)
            }
        }()
        var result bool
        if u.UpgradeRecord.Operation == "upgrade" {
            result = u.UpgradePackage()
        }
        if u.UpgradeRecord.Operation == "restart" {
            result = u.RestartService()
        }
        resultChan <- result
    }()
    select {
    case s := <- resultChan:
        beego.Info(u.UpgradeRecord.Operation + "操作完成")
        if s {
            u.SaveResult(1)
        } else {
            u.SaveResult(0)
        }
    case <- timeout:
        beego.Info(u.UpgradeRecord.Operation + "操作超时")
        u.Logs = append(u.Logs, "执行超时，请检查任务运行状态")
        u.SaveResult(0)
    }
}

// 升级服务
func (u *UnitVMUpgrade) UpgradePackage() bool {
    start:= time.Now()
    err := u.CreateRecord(2)
    if err != nil {
        beego.Error(err)
        return false
    }

    hosts := strings.Split(u.ConfigVM.Hosts, ";")
    err = u.MkWorkspace()
    if err != nil {
        beego.Error(err)
        return false
    }
    err = u.DownloadArtifact()
    if err != nil {
        beego.Error(err)
        return false
    }
    err = u.CheckArtifact()
    if err != nil {
        beego.Error(err)
        return false
    }
    err = u.UploadArtifact()
    if err != nil {
        beego.Error(err)
        return false
    }
    // 串行升级
    for _, h := range hosts {
        err := u.CheckServerDir(h, "dir")
        if err != nil {
            beego.Error(err)
            return false
        }
        if u.ConfigVM.CMDPre != "" {
            err = u.ExecPreCmd(h)
            if err != nil {
                beego.Error(err)
                return false
            }
        }
        if u.ConfigVM.NeedReboot == 1 && u.ConfigVM.CMDStop != "" {
            err = u.StopApp(h)
            if err != nil {
                beego.Error(err)
                return false
            }
        }
        err = u.UpgradeArtifact(h)
        if err != nil {
            beego.Error(err)
            return false
        }
        if u.ConfigVM.NeedReboot == 1 && u.ConfigVM.CMDStartup != "" {
            err = u.StartApp(h)
            if err != nil {
                beego.Error(err)
                return false
            }
        }
        if u.ConfigVM.CMDRear != "" {
            err = u.ExecRearCmd(h)
            if err != nil {
                beego.Error(err)
                return false
            }
        }
    }
    u.CleanWS()
    duration := time.Since(start).Seconds()
    u.UpgradeRecord.Duration = common.GetInt(duration)
    return true
}

// 重启服务
func (u *UnitVMUpgrade) RestartService() bool {
    start := time.Now()
    err := u.CreateRecord(2)
    if err != nil {
        beego.Error(err)
        return false
    }

    // 并行重启
    if u.ConfigVM.NeedReboot == 1 {
        hosts := strings.Split(u.ConfigVM.Hosts, ";")
        for _, h := range hosts {
            err = u.StopApp(h)
            if err != nil {
                beego.Error(err)
                return false
            }

            err = u.StartApp(h)
            if err != nil {
                beego.Error(err)
                return false
            }
        }
    }
    duration := time.Since(start).Seconds()
    u.UpgradeRecord.Duration = common.GetInt(duration)
    return true
}

func (u *UnitVMUpgrade) SaveResult(status int) {
    u.UpgradeRecord.Status = status
    updates := map[string]interface{}{
        "update_time": time.Now().Format("2006-01-02 15:04:05"),
        "status": u.UpgradeRecord.Status,
        "duration": u.UpgradeRecord.Duration,
        "logs": strings.Join(u.Logs, "\n"),
    }
    beego.Info(updates)
    tx := initial.DB.Begin()
    err := tx.Model(&u.UpgradeRecord).Updates(updates).Error
    if err != nil {
        beego.Error(err)
        tx.Rollback()
        return
    }
    tx.Commit()
}

func (u *UnitVMUpgrade) CreateRecord(status int) error {
    u.UpgradeRecord.Status = status
    u.UpgradeRecord.CreateTime = time.Now().Format("2006-01-02 15:04:05")
    u.UpgradeRecord.UpdateTime = u.UpgradeRecord.CreateTime
    u.UpgradeRecord.UnitID = u.ConfigVM.UnitID
    u.UpgradeRecord.ArtifactURL = u.OnlineVM.ArtifactURL

    tx := initial.DB.Begin()
    err := tx.Model(&models.OprVMUpgrade{}).Create(&u.UpgradeRecord).Error
    if err != nil {
        tx.Rollback()
        return err
    }
    tx.Commit()
    return nil
}