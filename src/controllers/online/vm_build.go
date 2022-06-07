package online

import (
	"controllers"
	"fmt"
	"github.com/astaxie/beego"
	high_conc "high-conc"
	"initial"
	"initial/jenkins_xml"
	"library/cfunc"
	"library/jenkins"
	"models"
	"strings"
)

// @Title 虚机应用构建
// @Description 虚机应用构建
// @Param online_id query string true "vm上线ID"
// @Success 200 true or false
// @Failure 403
// @router /vm/build [post]
func (c *StdVmOnlineController) VmBuild() {
	// TestPass @20191126 by chenhz001
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	// 1. 获取构建配置
	onlineId, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var onlineVm models.OnlineStdVM
	var vm models.UnitConfVM
	con := initial.DB
	err := con.Model(models.OnlineAllList{}).Where("id=? and is_delete = 0", onlineId).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = con.Model(models.OnlineStdVM{}).Where("online_id=? and is_delete = 0", onlineId).First(&onlineVm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = con.Model(models.UnitConfVM{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&vm).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	// 构建前判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有权限构建，只有此发布单元的负责人、开发人员和测试人员才可以构建！")
		return
	}
	if vm.IsConfirm == 0 {
		c.SetJson(0, "", "Jenkins配置没有得到发布单元负责人的确认，请先确认jenkins构建配置！")
		return
	}
	if  onlineVm.BuildStatus == 1 || onlineVm.BuildStatus == 2 || onlineVm.UpgradeStatus == 1 {
		c.SetJson(0, "", "正在发布中/构建成功/已发布完成，不允许再次构建！")
		return
	}

	// 2. 发起构建任务
	// 构建
	unitConf := cfunc.GetUnitInfoById(online.UnitId)
	unitName := strings.ToLower(unitConf.Unit)
	unitName = strings.Replace(unitName, "_", "-", -1)
	unitName = strings.Replace(unitName, " ", "", -1)
	jobName := fmt.Sprintf("%s-%s-%s-%s", strings.ToLower(vm.DeployComp), beego.AppConfig.String("runmode"),
		unitName, online.Version)
	jenkinsOpr := jenkins.JenkOpr{
		BaseUrl:   cfunc.GetJenkBaseUrl(),
		JobName:   jobName,
		ConfigXml: GetVMJenkinsXml(online, vm, jobName),
	}
	// 工作空间包路径
	targetPath := ""
	if vm.ArtifactPath != "" {
		path := strings.Trim(vm.ArtifactPath, " ")
		if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip") || strings.HasSuffix(path, ".jar") || strings.HasSuffix(path, ".war"){
			pathArr := strings.Split(path, "/")
			//if pathArr[len(pathArr) - 1] != vm.AppPKGName {
			//	c.SetJson(0, "", "Jenkins构建配置中工作空间包路径包含的文件名和配置中指定的发布文件名不一致！")
			//	return
			//}
			if len(pathArr) > 1 {
				targetPath = pathArr[0]
				for i := 1; i < len(pathArr)-1; i++ {
					targetPath = targetPath + "/" + pathArr[i]
				}
			}
		} else {
			targetPath = vm.ArtifactPath
		}
	}
	ArtifactUrl := ""
	if targetPath == "" {
		ArtifactUrl = fmt.Sprintf("%sjob/%s/ws/%s", cfunc.GetJenkBaseUrl(), jobName, vm.Artifact)
	} else {
		ArtifactUrl = fmt.Sprintf("%sjob/%s/ws/%s/%s", cfunc.GetJenkBaseUrl(), jobName, targetPath, vm.Artifact)
	}
	vmBuild := VMBuild{
		Opr: jenkinsOpr,
		OnlineId: online.Id,
		VMOnlineId: onlineVm.ID,
		ArtifactUrl: ArtifactUrl,
	}
	// 更改为执行中
	vmBuild.SaveBuildResult(2, 2, 0, "")
	high_conc.JobQueue <- &vmBuild
	c.SetJson(1, "", "虚机构建进入队列，请耐心等待！")
	// 3. 后台队列轮询构建结果
}

// @Title 虚机构建日志
// @Description 虚机构建日志
// @Param online_id query string true "vm上线ID"
// @Success 200 true or false
// @Failure 403
// @router /vm/jenk-log [get]
func (c *StdVmOnlineController) VmJenkinsLog() {
	onlineID := c.GetString("online_id")
	type InfoName struct {
		Info string `json:"info"`
	}

	var err error
	var info InfoName
	if err = initial.DB.Table("online_all_list a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Select("b.info").Where("a.id = ? and a.is_delete = 0", onlineID).First(&info).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var onlineVm models.OnlineStdVM
	if err = initial.DB.Model(models.OnlineStdVM{}).Where("online_id = ? AND is_delete = 0", onlineID).First(&onlineVm).Error; err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	if onlineVm.BuildStatus == 10 {
		ret := map[string]interface{}{
			"unit_info": info.Info,
			"log": "",
		}
		c.SetJson(0, ret, "构建任务未开始，请先选择构建！")
		return
	}

	if len(onlineVm.JenkinsName) < 10 {
		ret := map[string]interface{}{
			"unit_info": info.Info,
			"log": "jenkins的job名称有误！",
		}
		c.SetJson(0, ret, "jenkins的job名称有误！")
		return
	}

	var jenk_opr jenkins.JenkOpr
	jenk_opr.BaseUrl = cfunc.GetJenkBaseUrl()
	jenk_opr.JobName = onlineVm.JenkinsName
	jenk_opr.Init()
	output, err := jenk_opr.GetBuildLog()
	if err != nil {
		beego.Info(err.Error())
		ret := map[string]interface{}{
			"unit_info": info.Info,
			"log": err.Error(),
		}
		c.SetJson(0, ret, err.Error())
		return
	}

	//beego.Info(*output)
	ret := map[string]interface{}{
		"unit_info": info.Info,
		"log": output,
	}
	c.SetJson(1, ret, "jenkins日志获取成功！")
}

func GetVMJenkinsXml(online models.OnlineAllList, vm models.UnitConfVM, jobName string) string {
	xml := jenkins_xml.GetVMJenkinsXml()
	xml = jenkins_xml.GetBaseJenkinsXmlAssignNode(vm.JenkinsNode, xml)
	xml = strings.Replace(xml, "{{ xml.git_http_url }}", vm.GitURL, -1)
	xml = strings.Replace(xml, "{{ xml.git_sha }}", online.CommitId, -1)
	xml = strings.Replace(xml, "{{ xml.before_build_shell }}", vm.BeforeXml, -1)
	xml = strings.Replace(xml, "{{ xml.build_shell }}", vm.BuildXml, -1)
	xml = strings.Replace(xml, "{{ xml.after_build_shell }}", vm.AfterXml, -1)
	return xml
}
