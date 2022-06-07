package operation

import (
    "controllers"
    "encoding/json"
    "fmt"
    "github.com/astaxie/beego"
    "initial"
    "library/cae"
    "library/cfunc"
    "library/common"
    "models"
    "net"
    "os"
    "strings"
    "time"
)

type HostInitController struct {
    controllers.BaseController
}

func (c *HostInitController) URLMapping()  {
    c.Mapping("GetInitList", c.GetInitList)
    c.Mapping("AddInitTask", c.AddInitTask)
    c.Mapping("ExecInitTask", c.ExecInitTask)
    c.Mapping("DeleteInitTask", c.DeleteInitTask)
}

type TaskInput struct {
    UnitID      int     `json:"unit_id"`
    UnitName    string  `json:"unit_name"`
    Hosts       string  `json:"hosts"`
    DeployType  string  `json:"deploy_type"`
    DeployCorp  string  `json:"deploy_corp"`
    DeployVPC   string  `json:"deploy_vpc"`
    DeployENV   string  `json:"deploy_env"`
}

// Get init tasks 方法
// @Title Get init tasks
// @Description 获取主机初始化记录
// @Param key_word query string false "查询关键词，可按发布单元、操作人、主机IP搜索，支持中英文，支持模糊搜索"
// @Param page query string false "页数"
// @Param rows query string false "每页行数"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/init/list [get]
func (c *HostInitController) GetInitList()  {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var tasks []models.OprVMHost
    keyWord := c.GetString("key_word")
    page, rows := c.GetPageRows()
    queryStr := " type = 'init' and deleted = 0"
    if strings.TrimSpace(keyWord) != "" {
        queryStr += fmt.Sprintf(" and concat(unit, unit_cn, hosts, operator) like '%%%s%%' ", keyWord)
    }
    var count int
    err := initial.DB.Table("opr_vm_host").Where(queryStr).Count(&count).Order("id desc").
        Offset((page - 1)*rows).Limit(rows).Find(&tasks).Error
    if err != nil {
        beego.Error()
    }
    resp := map[string]interface{}{
        "count": count,
        "data": tasks,
    }
    c.SetJson(1, resp, "ok")
    return
}

// Add init task 方法
// @Title Add init task
// @Description 新增命令执行任务
// @Param body body TaskInput "body参数传递"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/init [post]
func (c *HostInitController) AddInitTask()  {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var input TaskInput
    if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
        c.SetJson(0, "", "参数解析失败！")
        return
    }
    // 部署配置检查
    var task models.OprVMHost
    var unit models.UnitConfList
    var unitConf models.UnitConfVM
    conn := initial.DB
    if err := conn.Model(&unitConf).Where("`unit_id` = ? AND `is_delete` = 0", input.UnitID).First(&unitConf).Error;
    err != nil ||unitConf.ID == 0 {
        c.SetJson(0, "", "未找到部署配置，请先添加！")
        return
    }

    // 主机IP校验
    if strings.TrimSpace(input.Hosts) == "" {
        c.SetJson(0, "", "主机IP列表不能未空！")
        return
    }
    inputIPs := strings.Split(input.Hosts, ";")
    for _, i := range inputIPs {
        ip := net.ParseIP(i)
        if ip.To4() == nil {
            msg := fmt.Sprintf("%s IP格式错误！", i)
            c.SetJson(0, "", msg)
            return
        }
        if !strings.Contains(unitConf.Hosts, i) {
            msg := fmt.Sprintf("主机 %s 不属于该发布单元！", i)
            c.SetJson(0, "", msg)
            return
        }
        var contain int
        queryStr := fmt.Sprintf("`unit_id` = %d " +
            "AND `hosts` like '%%%s%%' " +
            "AND `type` = 'init' " +
            "AND (`status` = 1 OR `status` = 3) " +
            "AND `deleted` = 0", input.UnitID, i)
        if err := conn.Model(&task).Where(queryStr).Count(&contain).Error; err != nil && contain != 0 {
            c.SetJson(0, "", "查询时出错！")
            return
        }
        if contain != 0 {
            msg := fmt.Sprintf("主机 %s 正在初始化或已初始化，请勿重复添加！", i)
            c.SetJson(0, "", msg)
            return
        }
    }

    // 应用类型校验
    input.DeployType = strings.TrimSpace(input.DeployType)
    if input.DeployType == "" {
        c.SetJson(0, "", "部署类型不能为空！")
        return
    }

    allowedTypes := []string{"jar", "war", "py2", "py3", "ng"}
    if !common.InList(strings.ToLower(input.DeployType), allowedTypes) {
        msg := fmt.Sprintf("不支持的部署类型 %s！", input.DeployType)
        c.SetJson(0, "", msg)
        return
    }

    // 部署租户校验
    input.DeployCorp = strings.TrimSpace(input.DeployCorp)
    if input.DeployCorp == "" {
        c.SetJson(0, "", "部署租户不能未空！")
        return
    }
    if cfunc.GetCompCnName(strings.ToUpper(input.DeployCorp)) == "" {
        msg := fmt.Sprintf("不支持的部署租户 %s！", input.DeployCorp)
        c.SetJson(0, "", msg)
        return
    }

    // 应用环境校验
    input.DeployENV = strings.TrimSpace(input.DeployENV)
    if input.DeployENV == "" {
        c.SetJson(0, "", "部署环境不能为空！")
        return
    }
    if strings.ToLower(input.DeployENV) != beego.AppConfig.String("runmode") {
        c.SetJson(0, "", "部署环境与当前环境不匹配！")
        return
    }

    if err := conn.Model(&unit).Where("`id` = ? AND `is_offline` = 0", input.UnitID).First(&unit).Error;
    err != nil || unit.Id == 0 {
        c.SetJson(0, "", "未找到该发布单元！")
        return
    }

    // 重复初始化校验
    //var count int
    //if err := conn.Model(&task).Where("`unit_id` = ? AND `type` = 'init' AND `status` != 2 AND `deleted` = 0", input.UnitID).
    //    Count(&count).Error; err != nil  {
    //    c.SetJson(0, "", "查询时发生错误！")
    //    return
    //}
    //if count != 0 {
    //    c.SetJson(0, "", "初始化任务已存在，请勿重复添加！")
    //    return
    //}

    task.UnitID = input.UnitID
    task.Unit = unit.Unit
    task.UnitCN = unit.Name
    task.Hosts = input.Hosts
    task.Type = "init"
    task.Command = "source /etc/profile && sh host_init.sh"
    randStr := common.GenRandString(8)
    task.ExecPath = "/tmp/" + randStr

    deployType := input.DeployType
    // 发布单元名格式化为 ***-*** 形式
    unitName := strings.ToLower(unit.Unit)
    unitName = strings.Replace(unitName, "_", "-", -1)
    deployEnv := strings.ToLower(input.DeployENV)
    deployZone := strings.ToLower(input.DeployCorp) + "-" + strings.ToLower(input.DeployVPC)
    task.Args = fmt.Sprintf("%s;%s;%s;%s", deployType, unitName, deployEnv, deployZone)

    task.ExecUser = "root"
    task.Status = 0
    task.Operator = c.UserId
    task.CreateTime = time.Now().Format("2006-01-02 15:04:05")
    task.Deleted = 0

    tx := conn.Begin()
    if err := tx.Model(&task).Create(&task).Error; err != nil {
        tx.Rollback()
        c.SetJson(0, "", "任务创建失败！")
        return
    }
    tx.Commit()
    c.SetJson(1, "", "任务创建成功！")
    return
}

// Exec init task 方法
// @Title Exec init task
// @Description 执行初始化任务
// @Param task_id path int true "任务ID"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/init/:task_id [put]
func (c *HostInitController) ExecInitTask()  {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    taskID := c.Ctx.Input.Param(":task_id")
    var task models.OprVMHost
    err := initial.DB.Where("`id` = ? and `deleted` = 0 and `type` = 'init'",  taskID).Find(&task).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "没有查询到该记录！")
        return
    }
    if task.Status == 1 {
        c.SetJson(0, "", "已执行成功的任务不允许重复执行！")
        return
    }
    if task.Status == 3 {
        c.SetJson(0, "", "任务正在执行中，请稍等！")
        return
    }

    // 开始执行
    task.Logs = "初始化开始：\n"
    // 本地脚本文件检查
    pwd, err := os.Getwd()
    if err != nil {
        beego.Error(err)
        task.Logs += err.Error()
        task.Status = 2
        if err := updateTask(task); err != nil {
            c.SetJson(0, "", "任务状态更新失败！" + err.Error())
            return
        }
        c.SetJson(0, task, "本地环境检查失败！" + err.Error())
        return
    }

    scriptFile := pwd + "/conf/file/host_init.sh"
    if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
        beego.Error(err)
        task.Logs += err.Error()
        task.Status = 2
        if err := updateTask(task); err != nil {
            c.SetJson(0, "", "任务状态更新失败！" + err.Error())
            return
        }
        c.SetJson(0, task, "未找到初始化脚本文件！")
        return
    }

    // 应用服务器端创建临时工作目录
    hosts := strings.Split(task.Hosts, ";")
    mkdirCmd := fmt.Sprintf("mkdir -p %s && chmod 777 %s", task.ExecPath, task.ExecPath)
    task.Logs += "1. 创建临时工作目录：\n"
    for _, h := range hosts {
        err, msg := cae.ExecCmd(mkdirCmd, "/tmp", task.ExecUser, h)
        task.Logs += strings.Join(msg, "\n") + "\n"
        if err != nil {
            beego.Error(err)
            task.Status = 2
            if err := updateTask(task); err != nil {
                c.SetJson(0, "", "任务状态更新失败！" + err.Error())
                return
            }
            c.SetJson(0, task, "临时目录创建失败！")
            return
        }
        task.Logs += "临时目录创建成功 \n"
    }

    // 上传初始化脚本
    task.Logs += "2. 上传初始化脚本：\n"
    err, msg := cae.TransFile(scriptFile, task.ExecPath, task.Hosts)
    //err, msg := cae.CopyFile(scriptFile, task.ExecPath, hosts)
    task.Logs += strings.Join(msg, "\n") + "\n"
    //task.Logs += msg
    if err != nil {
        beego.Error(err)
        task.Status = 2
        if err := updateTask(task); err != nil {
            c.SetJson(0, "", "任务状态更新失败！" + err.Error())
            return
        }
        c.SetJson(0, task, "初始化脚本上传失败！")
        return
    }
    // 构造脚本执行命令
    initCmd := task.Command
    args := strings.Split(task.Args, ";")
    for _, v := range args {
        initCmd += " " + v
    }

    // 远程执行初始化脚本
    task.Logs += "3. 执行初始化：\n"
    for _, h := range hosts {
        err, msg := cae.ExecCmd(initCmd, task.ExecPath, task.ExecUser, h)
        task.Logs += strings.Join(msg, "\n") + "\n"
        if err != nil {
            beego.Error(err)
            task.Status = 2
            if err := updateTask(task); err != nil {
                c.SetJson(0, "", "任务状态更新失败！" + err.Error())
                return
            }
            c.SetJson(0, task, "任务执行失败！")
            return
        }
    }

    // 清理临时工作目录
    task.Logs += "4. 清理工作目录：\n"
    cleanCmd := fmt.Sprintf("rm -rf %s", task.ExecPath)
    for _, h := range hosts {
        err, msg := cae.ExecCmd(cleanCmd, "/tmp", task.ExecUser, h)
        task.Logs += strings.Join(msg, "\n") + "\n"
        if err != nil {
            beego.Error(err)
            task.Status = 2
            if err := updateTask(task); err != nil {
                c.SetJson(0, "", "任务状态更新失败！" + err.Error())
                return
            }
            c.SetJson(0, "", "临时文件清理失败！")
            return
        }
    }

    // 更新状态
    task.Status = 1
    if err := updateTask(task); err != nil {
        c.SetJson(0, "", "任务状态更新失败！" + err.Error())
        return
    }
    c.SetJson(1, task, "任务执行成功")
    return
}

// Delete init task 方法
// @Title Delete init task
// @Description 删除初始化任务
// @Param task_id path int true "任务ID"
// @Success 200 {object} models.OprVMHost
// @Failure 403
// @router /host/init/:task_id [delete]
func (c *HostInitController) DeleteInitTask() {
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    taskID := c.Ctx.Input.Param(":task_id")
    var task models.OprVMHost
    err := initial.DB.Where("id = ? and deleted = 0 and type = 'init'", taskID).Find(&task).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "没有查询到该记录！")
        return
    }
    if task.Status == 1 {
        c.SetJson(0, "", "已执行成功的任务不允许删除！")
        return
    }
    err = initial.DB.Model(&task).Where("id = ? and deleted = 0", taskID).
        Update("deleted", 1).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "删除失败！")
        return
    }
    c.SetJson(1, "", "删除成功")
    return
}

func updateTask(task models.OprVMHost) error {
    tx := initial.DB.Begin()
    if err := tx.Model(&task).Where("`id` = ? AND `deleted` = 0 AND `type` = 'init'", task.ID).
        Updates(task).Error; err != nil {
            tx.Rollback()
        return err
    }
    tx.Commit()
    return nil
}