package online

import (
    "controllers"
    "encoding/json"
    "github.com/astaxie/beego"
    "initial"
    "library/cfunc"
    "library/common"
    "library/git"
    "models"
    "strings"
    "time"
)

type StdVmOnlineController struct {
    controllers.BaseController
}

func (c *StdVmOnlineController) URLMapping()  {
    c.Mapping("GetVMOnlineList", c.GetVMOnlineList)
    c.Mapping("AddVMOnlineTask", c.AddVMOnlineTask)
    c.Mapping("DelVMOnlineTask", c.DelVMOnlineTask)
    c.Mapping("VmBuild", c.VmBuild)
    c.Mapping("VmJenkinsLog",c.VmJenkinsLog)
    c.Mapping("VmUpgrade", c.VmUpgrade)
    c.Mapping("VmResultQuery",c.VmResultQuery)
}

// Get vm online list
// @Title Get vm online list
// @Description Get vm online list
// @Param unit_id query string false "发布单元id，为空全部发布单元"
// @Param online_date query string false "上线日期"
// @Param status query string false "状态:未开始not_start, 未完成not_finish, 失败fail"
// @Param	page	query	string	true	"页数"
// @Param	rows	query	string	true	"每页行数"
// @Success 200 {object} models.OnlineAllList
// @Failure 403
// @router /vm/list [get]
func (c *StdVmOnlineController) GetVMOnlineList() {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    unitID := c.GetString("unit_id")
    onlineDate := c.GetString("online_date")
    status := c.GetString("status")
    page, rows := c.GetPageRows()

    filtra := "a.is_delete = 0 AND b.is_delete = 0"
    if unitID != "" {
        filtra = filtra + " AND b.unit_id = " + unitID
    }
    if onlineDate != "" {
        //filtra = filtra + " AND b.online_data = " + "'" + onlineDate + "'"
        filtra = filtra + " AND b.online_date = " + onlineDate
    }

    if  status != "" {
        switch status {
        case "not_start":
            filtra = filtra + " AND b.is_success = 10"
        case "not_finish":
            filtra = filtra + " AND b.is_success = 2"
        case "fail":
            filtra = filtra + " AND b.is_success = 0"
        default:
            c.SetJson(0, "", "状态只允许为：未开始/未完成/失败，请重新选择！")
            return
        }
    }

    var err error
    var count int
    vmList := make([]*models.OnlineAllList, 0)
    if err = initial.DB.Table("online_std_vm a").Joins("LEFT JOIN online_all_list b ON a.online_id = b.id").Select("b.*").Where(filtra).
        Count(&count).Order("a.id desc").Offset((page - 1)*rows).Limit(rows).Find(&vmList).Error ; err != nil {
            c.SetJson(0, "", err.Error())
            return
    }

    type RetData struct {
        AllList      *models.OnlineAllList  `json:"all_list"`
        StdVM        *models.OnlineStdVM  `json:"std_Vm"`
        UnitCnName   string   `json:"unit_cn_name"`
        UnitEnName   string   `json:"unit_en_name"`
        AppType      string    `json:"app_type"`
        OperatorName string   `json:"operator_name"`
    }
    data := make([]*RetData,0)
    for _, v := range vmList {
        vmStd := new(models.OnlineStdVM)
        if err = initial.DB.Model(models.OnlineStdVM{}).Where("online_id = ?",v.Id).First(vmStd).Error ; err != nil {
            c.SetJson(0, "", err.Error())
            return
        }
        var vmConf models.UnitConfVM
        if err = initial.DB.Model(models.UnitConfVM{}).Where("unit_id = ? ", v.UnitId).First(&vmConf).Error ; err != nil {
            c.SetJson(0, "", err.Error())
            return
        }
        unitInfo := cfunc.GetUnitInfoById(v.UnitId)
        data = append(data, &RetData{
            AllList:v,
            StdVM:vmStd,
            UnitCnName:unitInfo.Name,
            UnitEnName: unitInfo.Unit,
            AppType:vmConf.AppType,
            OperatorName: cfunc.GetUserCnName(v.Operator),
        })
    }

    res := map[string]interface{}{
        "count": count,
        "data": data,
    }
    c.SetJson(1, res, "获取数据成功！")
}

// Add vm online task
// @Title 新增虚机发布任务
// @Description 新增虚机发布任务
// @Param body body online.VMOnlineInput    "传入Body类型数据"
// @Success 200 true or false
// @Failure 403
// @router /vm/save [post]
func (c *StdVmOnlineController) AddVMOnlineTask()  {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }
    var input VMOnlineInput
    err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
    if err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(input.UnitID, c.UserId) {
        c.SetJson(0, "", "您没有此发布单元的添加权限，请联系相关发布单元的负责人、开发人员和测试人员新增！")
        return
    }

    var vmConf models.UnitConfVM
    var onlineTask models.OnlineAllList
    var vmOnlineTask models.OnlineStdVM
    if input.UnitID == 0 {
        c.SetJson(0, "", "发布单元不能为空！")
        return
    }
    if err = initial.DB.Model(&models.UnitConfVM{}).Where("is_delete = 0 AND unit_id = ?", input.UnitID).First(&vmConf).Error ; err != nil {
        beego.Error(err.Error())
        c.SetJson(0, "", err.Error())
        return
    }
    if strings.TrimSpace(input.Branch) == "" {
        c.SetJson(0, "", "发布分支不能为空！")
        return
    }
    onlineTask.UnitId = input.UnitID
    onlineTask.Branch = input.Branch

    // 上线日期和时间的数据合法性校验
    now := time.Now().Format(initial.DatetimeFormat)
    dateOnline, timeOnlie, err := common.CheckOnlineDate(input.OnlineDate, input.OnlineTime, now)
    if err != nil {
        c.SetJson(0, "", err.Error())
        return
    }
    onlineTask.OnlineDate, onlineTask.Version = dateOnline, dateOnline
    onlineTask.OnlineTime = timeOnlie

    //onlineTask.IsProcessing = 0
    onlineTask.IsSuccess = 10
    onlineTask.IsDelete = 0
    onlineTask.Operator = c.UserId
    onlineTask.ExcuteTime = ""
    onlineTask.InsertTime = now
    onlineTask.ErrorLog = ""

    if strings.TrimSpace(input.SHA) == "" {
        branchDetail := git.GetBranchDetail(vmConf.GitID, input.Branch)
        if branchDetail == nil {
            c.SetJson(0, "", "分支输入错误，请重新输入！")
            return
        }
        onlineTask.CommitId = branchDetail.Commit.ID
        onlineTask.ShortCommitId = branchDetail.Commit.ShortID
    } else {
        commitDetail := git.GetCommitDetail(vmConf.GitID, input.SHA)
        if commitDetail == nil {
            c.SetJson(0, "", "sha值输入错误，请重新输入！")
            return
        }
        onlineTask.CommitId = commitDetail.ID
        onlineTask.ShortCommitId = commitDetail.ShortID
    }

    tx := initial.DB.Begin()
    err = tx.Create(&onlineTask).Error
    if err != nil {
        c.SetJson(0, "", err.Error())
        tx.Rollback()
        return
    }
    vmOnlineTask.OnlineID = onlineTask.Id
    vmOnlineTask.CreateTime = now
    vmOnlineTask.BuildStatus = 10
    vmOnlineTask.UpgradeStatus = 10
    vmOnlineTask.UpgradeDuration = 0
    vmOnlineTask.IsDelete = 0
    err = tx.Create(&vmOnlineTask).Error
    if err != nil {
        tx.Rollback()
        c.SetJson(0, "", err.Error())
        return
    }
    tx.Commit()

    c.SetJson(1, "", "标准虚机上线单元创建成功！")
    return
}

type VMOnlineInput struct {
    UnitID      int     `json:"unit_id"`
    Branch      string  `json:"branch"`
    SHA         string  `json:"sha"`
    OnlineDate  string  `json:"online_date"`
    OnlineTime  string  `json:"online_time"`
}

// Delete vm online task
// @Title 删除虚机发布任务
// @Description 删除虚机发布任务
// @Param task_id query int true "待删除的任务ID"
// @Success 200 true or false
// @Failure 403
// @router /vm/del/:task_id [delete]
func (c *StdVmOnlineController) DelVMOnlineTask()  {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有操作权限！")
        return
    }
    taskID := c.Ctx.Input.Param(":task_id")
    var (
        onlineList models.OnlineAllList
        stdVm models.OnlineStdVM
        err error
    )
    if err = initial.DB.Model(models.OnlineAllList{}).Where("id = ? AND is_delete = 0", taskID).First(&onlineList).Error ; err != nil {
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    if err = initial.DB.Model(models.OnlineStdVM{}).Where("online_id = ? AND is_delete = 0", taskID).First(&stdVm).Error ; err != nil {
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(onlineList.UnitId, c.UserId) {
        c.SetJson(0, "", "您没有此发布单元的删除权限！")
        return
    }

    // 只允许删除未构建/构建失败/构建成功未发布 这三种状态的任务
    flag := false
    if onlineList.IsSuccess == 10 && stdVm.BuildStatus == 10{
    flag = true
    }
    if stdVm.UpgradeStatus == 10 && stdVm.BuildStatus == 0 && onlineList.IsSuccess == 2 {
    flag = true
    }
    if stdVm.UpgradeStatus == 10  && stdVm.BuildStatus == 1 && onlineList.IsSuccess == 2{
    flag = true
    }
    if flag == false {
    c.SetJson(0, "", "已执行完成或者执行中的任务，不允许删除！")
    return
    }

    tx := initial.DB.Begin()
    err = tx.Model(&models.OnlineAllList{}).Where("id = ? AND is_delete = 0", taskID).Update("is_delete", 1).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    err = tx.Model(&models.OnlineStdVM{}).Where("online_id = ? AND is_delete = 0", taskID).Update("is_delete", 1).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", err.Error())
        return
    }
    tx.Commit()
    c.SetJson(1, "", "数据删除成功！")
    return
}

// @Title 定时任务获取标准虚机应用发布结果
// @Description 定时任务获取标准虚机应用发布结果
// @Param	online_list	    query	string	ture	"标准虚机应用的上线列表，比如`1,2,3`"
// @Success 200  true or false
// @Failure 403
// @router /vm/result/query [get]
func (c *StdVmOnlineController) VmResultQuery() {
    if strings.Contains(c.Role, "guest") == true {
        c.SetJson(0, "", "您没有权限操作！")
        return
    }

    online_list := c.GetString("online_list")
    type Ret struct {
        OnlineId      int  `json:"online_id"`
        IsSuccess     int  `json:"is_success"`
        BuildStatus   int  `json:"build_status"`
        BuildDuration int  `json:"build_duration"`
        UpgradeStatus int  `json:"upgrade_status"`
        UpgradeDuration int `json:"upgrade_duration"`
        Duration    int     `json:"duration"`
        PkgFile string      `json:"pkg_file"`
    }

    onlineArr := strings.Split(online_list, ",")
    var ret []Ret
    for _, v := range onlineArr {
        if strings.TrimSpace(v) == "" {
            continue
        }
        var vm models.OnlineStdVM
        var err error
        if err = initial.DB.Model(models.OnlineStdVM{}).Where("online_id = ? AND is_delete = 0", v).First(&vm).Error ; err != nil {
            c.SetJson(0, "", err.Error())
            return
        }
        var vmAll models.OnlineAllList
        if err = initial.DB.Model(models.OnlineAllList{}).Where("id = ? AND is_delete = 0", v).First(&vmAll).Error ; err != nil {
            c.SetJson(0, "", err.Error())
            return
        }
        upgradeRet := vm.UpgradeStatus
        ret = append(ret, Ret{
            OnlineId:common.GetInt(v),
            IsSuccess:vmAll.IsSuccess,
            BuildStatus:vm.BuildStatus,
            BuildDuration:vm.BuildDuration,
            UpgradeDuration:vm.UpgradeDuration,
            UpgradeStatus:upgradeRet,
            PkgFile:vm.ArtifactURL,
            Duration: vm.BuildDuration + vm.UpgradeDuration,
        })
    }
    c.SetJson(1, ret, "标准虚机应用发布结果获取成功！")
}
