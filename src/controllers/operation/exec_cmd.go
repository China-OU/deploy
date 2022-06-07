package operation

import (
    "controllers"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/jinzhu/gorm"
    "initial"
    "library/cae"
    "library/common"
    "models"
    "net"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type ExecCmdController struct {
    controllers.BaseController
}

func (c *ExecCmdController) URLMapping() {
    c.Mapping("GetCmdList", c.GetCmdList)
    c.Mapping("AddTask", c.AddExecTask)
    c.Mapping("ExecTask", c.ExecTask)
}

// Get exec tasks 方法
// @Title GetCmdList
// @Description 获取命令行执行记录
// @Param start query string false "查询开始时间，传递2020-01-02格式"
// @Param end query string false "查询结束时间，传递2020-01-02格式"
// @Param unit query string false "是否绑定发布单元，y是，n否，nil-全部"
// @Param status query string false "命令执行状态，0-未执行，1-执行成功，2-执行失败，3-执行中，nil-全部"
// @Param key_word query string false "查询关键词，可按发布单元、操作人、主机IP搜索，支持中英文，支持模糊搜索"
// @Param page query string true "页数"
// @Param rows query string true "每页行数"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/cmd [get]
func (c *ExecCmdController) GetCmdList() {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }

    var err error
    sTime := c.GetString("start")
    eTime := c.GetString("end")
    filtra := "type = 'cmd' AND deleted = 0"
    if sTime != "" && eTime != "" {
        if sTime, eTime, err = common.FormatTime("",sTime, eTime) ; err != nil {
            beego.Error(err)
            c.SetJson(0,"",err.Error())
            return
        }
        filtra = fmt.Sprintf("%s AND create_time >= '%s' AND create_time < '%s'",filtra, sTime, eTime)
    }

    unit  := c.GetString("unit")
    if unit != "" {
        switch unit {
        case "y":
            filtra = fmt.Sprintf("%s AND unit_id != %d AND unit_cn != ''",filtra, 0)
        case "n":
            filtra = fmt.Sprintf("%s AND unit_id = %d AND unit_cn = ''",filtra, 0)
        default:
            c.SetJson(0,"","是否绑定发布单元只能为:是/否；当不指定时为查询全部！")
            return
        }
    }

    status := c.GetString("status")
    if status != "" && (status == "0" || status == "1" || status == "2" || status == "3") {
        s, err := strconv.Atoi(status)
        if err != nil {
            c.SetJson(0,"",err.Error())
            return
        }
        filtra = fmt.Sprintf("%s AND status = %d",filtra, s)
    }else if status != "" {
        c.SetJson(0,"","命令执行状态只能为:未执行/成功/失败/执行中；当不指定时为查询全部！")
        return
    }

    keyWord := c.GetString("key_word")
    if strings.TrimSpace(keyWord) != "" {
        filtra += fmt.Sprintf(" and concat(unit, unit_cn, hosts, operator) like '%%%s%%' ", keyWord)
    }

    page, rows := c.GetPageRows()
    var count int
    var tasks []models.OprVMHost
    err = initial.DB.Table("opr_vm_host").Where(filtra).Count(&count).Order("id desc").
        Offset((page - 1)*rows).Limit(rows).Find(&tasks).Error
    if err != nil {
        beego.Error(err)
        return
    }
    resp := map[string]interface{}{
        "count": count,
        "data": tasks,
    }
    c.SetJson(0, resp, "服务数据获取成功！")
}

// Post and run 方法
// @Title Post and run
// @Description 远程执行命令
// @Param body body models.OprVMHost "body raw参数传入"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/run [post]
func (c *ExecCmdController) PostAndRun() {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var task models.OprVMHost
    if err := json.Unmarshal(c.Ctx.Input.RequestBody, &task); err != nil {
        c.SetJson(0, "", "数据解析失败！" + err.Error())
        return
    }
    task.Operator = c.UserId
    task.Type = "cmd"

    if err := vmOprDataCheck(&task); err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    var unit models.UnitConfVM
    if task.Unit != "" {
        queryStr := fmt.Sprintf("name = '%s'", task.Unit)
        if err := initial.DB.Model(&unit).Where(queryStr).First(&unit).Error; err != nil {
            beego.Error(err)
        }
    }
}

func vmOprDataCheck(t *models.OprVMHost) error {
    //var unit models.UnitConfList
    var unitConf models.UnitConfVM
    conn := initial.DB

    // 主机IP合法性校验
    if t.Hosts == "" {
        return errors.New("主机IP不能为空！")
    }
    ips := strings.Split(t.Hosts, ";")
    ipM := make(map[string]bool)
    for _, i := range ips {
        ip := net.ParseIP(i)
        if ip.To4() == nil {
            errMsg := fmt.Sprintf("IP(%s)格式错误！", i)
            return errors.New(errMsg)
        }
        // IP&发布单元关联校验
        if !common.Empty(t.UnitID) && t.UnitID != 0 {
            var count int
            // 发布单元是否存在校验
            queryStr := fmt.Sprintf("`is_delete` = 0 AND `unit_id` = %d", t.UnitID)
            if err := conn.Model(&unitConf).Where(queryStr).Count(&count).Error; err != nil {
                beego.Error(err)
                return err
            }
            if count == 0 {
                return errors.New("未查询到该发布单元配置！")
            }
            // 主机IP是否匹配校验
            queryStr = fmt.Sprintf("`is_delete` = 0 AND `unit_id` = %d AND `hosts` like '%%%s%%'", t.UnitID, i)
            if err := conn.Model(&unitConf).Where(queryStr).Count(&count).Error; err != nil {
                return err
            }
            if count == 0 {
                errMsg := fmt.Sprintf("主机 %s 不属于发布单元 %s，请检查！", i, t.Unit)
                return errors.New(errMsg)
            }
        }
        // IP重复校验
        if _, ok := ipM[i]; ok {
            msg := fmt.Sprintf("主机 %s 已存在，请勿重复输入！", i)
            return errors.New(msg)
        }
        ipM[i] = true
    }
    // 设置默认 exec_path 和 exec_user
    if t.ExecPath == "" {
        t.ExecPath = "/tmp"
    }
    if t.ExecUser == "" {
        t.ExecUser = "rhlog"
    }

    // 命令白名单校验
    if t.Type == "cmd" {
        cmd := strings.TrimSpace(t.Command)
        whitelist := "(^(curl|ping|netstat|ps|free|uptime|file|w|who|hostname|df|uname|stat))"
        blacklist := "(&&|;|\\|\\||\\|)"
        wMatch, err := regexp.Match(whitelist, []byte(cmd))
        if err != nil {
            beego.Error(err.Error())
        }
        bMatch, err := regexp.Match(blacklist, []byte(cmd))
        if err != nil {
            beego.Info(err.Error())
        }
        if !wMatch || bMatch {
            beego.Error("暂不支持该命令！")
            return errors.New("暂不支持该命令！")
        }

        // 针对 curl 和 ping 的优化处理
        if strings.HasPrefix(cmd, "curl") {
            if !strings.Contains(cmd, " --connect-timeout ") {
                cmd += " --connect-timeout 5"
            }
        }
        if strings.HasPrefix(cmd, "ping") {
            if !strings.Contains(cmd, " -c ") {
                cmd += " -c 5"
            }
        }
        t.Command = cmd
    }
    return nil
}

// Add exec task 方法
// @Title AddTask
// @Description 新增命令执行任务
// @Param body body models.OprVMHost "使用raw body进行参数传递，必选commands/hosts，如需关联发布单元需增加unit_id / unit / unit_cn"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/cmd [post]
func (c *ExecCmdController) AddExecTask() {
    if strings.Contains(c.Role, "admin") == false {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var task models.OprVMHost
    err := json.Unmarshal(c.Ctx.Input.RequestBody, &task)
    if err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    task.Operator = c.UserId
    task.Type = "cmd"

    // 输入数据校验
    if err := inputCheck(&task); err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    // 主机是否关联发布单元
    if err := hostByUnit(&task) ; err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    task.CreateTime = time.Now().Format("2006-01-02 15:04:05")
    task.Deleted = 0
    tx := initial.DB.Begin()
    err = tx.Create(&task).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", "新增执行任务失败！")
        return
    }
    tx.Commit()

    c.SetJson(1, task, "新增执行任务成功！")
}

// Exec task 方法
// @Title Exec task
// @Description 执行命令
// @Param task_id query int true "任务ID"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/cmd/:task_id [put]
func (c *ExecCmdController) ExecTask()  {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    taskID, err := c.GetInt("task_id")
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    // 查找记录 taskID
    var task models.OprVMHost
    err = initial.DB.Where("id = ? and deleted = 0 and type = 'cmd'",  taskID).Find(&task).Error
    if err != nil {
        beego.Error()
        c.SetJson(0, "", "没有查询到该记录！")
        return
    }
    if task.Status == 1 {
        c.SetJson(0, "", "已执行成功的任务不允许重复执行！")
        return
    }

    // 执行命令
    hosts := strings.Split(task.Hosts, ";")
    for _, h := range hosts {
        err, msg := cae.ExecCmd(task.Command, task.ExecPath, task.ExecUser, h)
        task.Logs += strings.Join(msg, "\n")
        if err != nil {
            beego.Error(err)
            c.SetJson(1, "", strings.Join(msg, "\n"))
            return
        }
    }
    // 更新状态
    task.Status = 3
    tx := initial.DB.Begin()
    err = initial.DB.Model(&task).Where("id = ? and deleted = 0 and type = 'cmd'", taskID).
        Update("status", task.Status).Update("logs", task.Logs).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, task, "任务状态更新失败！")
        return
    }
    tx.Commit()
    c.SetJson(0, task, "执行成功！")
}

// Delete cmd task 方法
// @Title Delete cmd task
// @Description 删除命令
// @Param task_id path int true "任务ID"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/cmd/:task_id [post]
func (c *ExecCmdController) DeleteCmdTask() {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    taskID := c.Ctx.Input.Param(":task_id")
    var task models.OprVMHost
    err := initial.DB.Where("id = ? and deleted = 0 and type = 'cmd'", taskID).Find(&task).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "没有查询到该记录！")
        return
    }
    if task.Status != 0 {
        c.SetJson(0, "", "已执行过的任务不允许删除！")
        return
    }
    tx := initial.DB.Begin()
    err = initial.DB.Model(&task).Where("id = ? and deleted = 0", taskID).
        Update("deleted", 1).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", "删除失败！")
        return
    }
    tx.Commit()
    c.SetJson(0, "", "删除成功")
}

func inputCheck(task *models.OprVMHost) error {
    if task.Hosts == "" {
        beego.Error("主机IP不能为空！")
        return errors.New("主机IP不能为空！")
    }

    // 设置默认 exec_path 和 exec_user
    if task.ExecPath == "" {
        task.ExecPath = "/tmp"
    }
    if task.ExecUser == "" {
        task.ExecUser = "rhlog"
    }

    // 命令白名单校验
    if task.Type == "cmd" {
        cmd := strings.TrimSpace(task.Command)
        whitelist := "(^(curl|ping|netstat|ps|free|uptime|file|w|who|hostname|df|uname|stat))"
        blacklist := "(&&|;|\\|\\||\\|)"
        wMatch, err := regexp.Match(whitelist, []byte(cmd))
        if err != nil {
            beego.Error(err.Error())
        }
        bMatch, err := regexp.Match(blacklist, []byte(cmd))
        if err != nil {
            beego.Info(err.Error())
        }
        if !wMatch || bMatch {
            beego.Error("暂不支持该命令！")
            return errors.New("暂不支持该命令！")
        }

        // 针对 curl 和 ping 的优化处理
        if strings.HasPrefix(cmd, "curl") {
            if !strings.Contains(cmd, " --connect-timeout ") {
                cmd += " --connect-timeout 5"
            }
        }
        if strings.HasPrefix(cmd, "ping") {
            if !strings.Contains(cmd, " -c ") {
                cmd += " -c 5"
            }
        }
        task.Command = cmd
    }

    return nil
}

func hostByUnit(task *models.OprVMHost) error {
    var (
        vm models.UnitConfVM
        err error
    )

    ips := strings.Split(task.Hosts, ";")
    ipM := make(map[string]bool)
    for _, i := range ips {
        ip := net.ParseIP(i)
        if ip.To4() == nil {
            msg := fmt.Sprintf("IP(%s)格式错误！", i)
            beego.Error(msg)
            return errors.New(msg)
        }
        // 发布单元关联校验
        if task.UnitID != 0 {
            vm.UnitID = task.UnitID
            if err = initial.DB.Model(&models.UnitConfVM{}).Where("unit_id = ?",task.UnitID).First(&vm).Error ; err != nil {
                if err == gorm.ErrRecordNotFound {
                    return errors.New("找不到该发布单元的相关配置，请检查！")
                }
                return err
            }
            if !strings.Contains(vm.Hosts,i) {
                return errors.New("主机地址不属于该发布单元，请检查！")
            }
            if _, ok := ipM[i]; ok {
                msg := fmt.Sprintf("主机 %s 已存在，请勿重复输入！", i)
                beego.Error(msg)
                return errors.New(msg)
            }
            ipM[i] = true
        }else {
            err = initial.DB.Where("hosts like ?","%"+i+"%").Find(&vm).Error
            if err == gorm.ErrRecordNotFound {
                if _, ok := ipM[i]; ok {
                    msg := fmt.Sprintf("主机 %s 已存在，请勿重复输入！", i)
                    beego.Error(msg)
                    return errors.New(msg)
                }
                ipM[i] = true
            }else if err == nil {
                return errors.New("主机地址已关联发布单元，请选择发布单元后重新提交！")
            }else {
                return err
            }
        }
    }
    return nil
}
