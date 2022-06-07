package operation

import (
    "controllers"
    "encoding/json"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/jinzhu/gorm"
    high_conc "high-conc"
    "initial"
    "library/cae"
    "library/common"
    "models"
    "strconv"
    "strings"
)

type VMOprController struct {
    controllers.BaseController
}

func (c *VMOprController) URLMapping()  {
    c.Mapping("VMAppStatus", c.VMAppStatus)
}

type VMInstance struct {
    IP          string  `json:"ip"`
    PortStat    string  `json:"port_stat"`
    PID         string  `json:"pid"`
    ProcStat    string  `json:"proc_stat"`
    Status      string  `json:"status"`
}

// @Title 获取虚机应用运行状态
// @Description 获取虚机应用运行状态
// @Param unit_id query string true "发布单元ID"
// @Success 200 {object} {}
// @Failure 403
// @router /vm/state [get]
func (c *VMOprController) VMAppStatus()  {
    unitID := c.GetString("unit_id")
    var unitConf    models.UnitConfList
    var unitConfVM  models.UnitConfVM
    err := initial.DB.Model(&unitConfVM).Where("`unit_id` = ? AND `is_delete` = 0", unitID).First(&unitConfVM).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "数据查询失败！")
        return
    }
    err = initial.DB.Model(&unitConf).Where("`id` = ? AND `is_offline` = 0", unitConfVM.UnitID).First(&unitConf).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "数据查询失败！")
        return
    }

    var instance VMInstance
    var allInstance []VMInstance
    // 检查主机端口状态
    ports := strings.Split(unitConfVM.AppBindPort, ";")
    var portStr string
    for _, p := range ports {
        portStr += fmt.Sprintf("%s|", p)
    }
    protStr := strings.ToLower(unitConfVM.AppBindProt) + "|" + strings.ToLower(unitConfVM.AppBindProt) + "6"
    portStr = portStr[:len(portStr) - 1]
    portCheck := fmt.Sprintf("netstat -anp | grep -w LISTEN | grep -w -E '%s' | grep -w -E '%s'", portStr, protStr)

    for _, h := range strings.Split(unitConfVM.Hosts, ";") {
        instance.IP = h
        err, _, instance.PortStat = cae.ExecCmdSingleHost(portCheck, "/tmp", "root", h)
        if err != nil {
            instance.PID = "N/A"
            instance.ProcStat = "进程状态未知"
            instance.Status = "failed"
            if instance.PortStat == "stdout: stderr: " {
                instance.PortStat = "应用端口未在监听"
            }
            allInstance = append(allInstance, instance)
            continue
        }
        portSplit := strings.Split(instance.PortStat, "/")
        proc := portSplit[0]
        procSplit := strings.Split(proc, " ")
        instance.PID = procSplit[len(procSplit) - 1]
        pid := common.GetInt(instance.PID)
        if pid == 0 {
            beego.Error("进程PID获取失败")
            instance.ProcStat = "进程状态未知"
            instance.Status = "failed"
            continue
        }
        procCheck := fmt.Sprintf("ps -ef | grep -v grep | grep -w %s", instance.PID)
        err, _ , instance.ProcStat = cae.ExecCmdSingleHost(procCheck, "/tmp", "root", h)
        if err != nil {
            if instance.ProcStat == "stdout: stderr: " {
                instance.ProcStat = "未查询到进程，PID: " + instance.PID
                instance.Status = "failed"
            }
        }
        instance.Status = "ok"
        allInstance = append(allInstance, instance)
    }

    dataDetail := map[string]interface{}{
        "unitEN": unitConf.Unit,
        "unitCN": unitConf.Name,
        "bindProt": unitConfVM.AppBindProt,
        "bindPort": unitConfVM.AppBindPort,
        // 需增加查询当前版本的方法
        "version": "N/A",
    }
    data := map[string]interface{}{
        "detail": dataDetail,
        // version 版本应显示在每台主机上
        "instance": allInstance,
    }
    c.SetJson(1, data, "服务数据获取成功！")
    return
}

// @Title 虚机应用升级
// @Description 虚机应用升级
// @Param body body operation.UpgradeInput true	"body形式的数据，发布单元id名和镜像"
// @Success 200 {object} {}
// @Failure 403
// @router /vm/upgrade [post]
func (c *VMOprController) VMAppUpgrade()  {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var input UpgradeInput
    var vm UnitVMUpgrade
    err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    unitID, err := strconv.Atoi(input.UnitId)
    if err != nil {
        c.SetJson(0, "", "发布单元ID转换出错")
        return
    }

    err = initial.DB.Model(&vm.ConfigVM).Where("`is_delete` = 0 AND `unit_id` = ?", unitID).First(&vm.ConfigVM).Error
    if err == gorm.ErrRecordNotFound {
        c.SetJson(0, "", "未查询到该发布单元的配置")
        return
    } else if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "查询时出错")
        return
    }

    runningCheck := fmt.Sprintf("`status` = 2 AND `unit_id` = %d", unitID)
    var count int
    err = initial.DB.Model(&models.OprVMUpgrade{}).Where(runningCheck).Count(&count).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        beego.Error(err)
        c.SetJson(0, "", "查询时出错")
        return
    }
    if count > 0 {
        c.SetJson(0, "", "该发布单元有升级或重启操作正在进行中")
        return
    }

    // 获取发布单元中英文名
    var unitConf models.UnitConfList
    err = initial.DB.Model(&models.UnitConfList{}).Where("`id` = ?", unitID).Count(&count).Find(&unitConf).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "查询时出错")
        return
    }
    if count == 0 {
        c.SetJson(0, "", "没有查询到该发布单元")
        return
    }

    vm.OnlineVM.ArtifactURL = input.Image
    vm.UpgradeRecord.UnitName = fmt.Sprintf("%s(%s)", unitConf.Name, unitConf.Unit)
    vm.UpgradeRecord.Operator = c.UserId
    vm.UpgradeRecord.Operation = "upgrade"

    high_conc.JobQueue <- &vm
    c.SetJson(1, "", "升级任务已进入操作队列，请等待")
}

// @Title 获取升级历史
// @Description 获取升级历史
// @Param unit_id query string true "发布单元ID"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页多少行数"
// @Success 200 {object} {}
// @Failure 403
// @router /vm/upgrades [get]
func (c *VMOprController) VMAppUpgradeHistory()  {
    unitID := c.GetString("unit_id")
    page, rows := c.GetPageRows()
    queryStr := "1 = 1"
    if strings.TrimSpace(unitID) != "" {
        queryStr = fmt.Sprintf("`unit_id` = %d", common.GetInt(unitID))
    }
    var upgradeList []models.OprVMUpgrade
    var count int
    err := initial.DB.Model(&models.OprVMUpgrade{}).Where(queryStr).Count(&count).Order("create_time desc").Offset((page - 1)*rows).
        Limit(rows).Find(&upgradeList).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        beego.Error(err)
        c.SetJson(0, "", "查询数据时出错")
        return
    }
    data := map[string]interface{}{
        "count": count,
        "data": upgradeList,
    }
    c.SetJson(1, data, "ok")
    return
}

// @Title 虚机应用重启
// @Description 虚机应用重启
// @Param unit_id query string true "发布单元ID"
// @Success 200 {object} {}
// @Failure 403
// @router /vm/restart [post]
func (c *VMOprController) VMAppRestart()  {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var unitID int
    id := c.GetString("unit_id")
    if strings.TrimSpace(id) == "" {
        c.SetJson(0, "", "发布单元ID不能为空")
        return
    }
    unitID = common.GetInt(id)
    if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(unitID, c.UserId) {
        c.SetJson(0, "", "您没有权限操作此发布单元，请联系该发布单元的负责人、开发人员或测试人员")
        return
    }

    var vm UnitVMUpgrade
    err := initial.DB.Model(&models.UnitConfVM{}).Where("`is_delete` = 0 AND `unit_id` = ?", unitID).First(&vm.ConfigVM).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        beego.Error(err)
        c.SetJson(0, "", "查询数据时出错")
        return
    }

    runningCheck := fmt.Sprintf("`status` = 2 AND `unit_id` = %d", unitID)
    var count int
    err = initial.DB.Model(&models.OprVMUpgrade{}).Where(runningCheck).Count(&count).Error
    if err != nil && err != gorm.ErrRecordNotFound {
        beego.Error(err)
        c.SetJson(0, "", "查询时出错")
        return
    }
    if count > 0 {
        c.SetJson(0, "", "该发布单元有升级或重启操作正在进行中")
        return
    }

    vm.UpgradeRecord.Operator = c.UserId
    vm.UpgradeRecord.Operation = "restart"

    high_conc.JobQueue <- &vm
    c.SetJson(1, "", "重启任务已进入操作队列，请等待")
    return
}
