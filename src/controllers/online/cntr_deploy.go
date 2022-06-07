package online

import (
	"strings"
	"controllers"
	"models"
	"initial"
	"library/jenkins"
	"library/cfunc"
	"fmt"
	"github.com/astaxie/beego"
	"initial/jenkins_xml"
	"high-conc"
	"library/caas"
	"library/common"
	"controllers/operation"
	"controllers/unit_conf"
	"errors"
)

// @Title 标准容器应用的jenkins构建，分两步，先构建再更新容器
// @Description 标准容器应用的jenkins构建，分两步，先构建再更新容器
// @Param	online_id	    query	string	true	"标准容器发布单元的id"
// @Success 200 true or false
// @Failure 403
// @router /cntr/build [post]
func (c *StdCntrOnlineController) CntrBuild() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var online_cntr models.OnlineStdCntr
	var cntr models.UnitConfCntr
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=? and is_delete=0", online_id).First(&online_cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfCntr{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	cntr = GetMcpSubinfo(cntr)

	// 构建前判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if online.IsSuccess == 1 || online_cntr.JenkinsSuccess == 2 {
		// 什么情况下不允许再次构建，发布成功不允许再次构建
		c.SetJson(0, "", "正在构建中或者已发布完成，不允许再次点击！")
		return
	}
	if online_cntr.JenkinsSuccess == 1 {
		c.SetJson(0, "", "该sha值已构建成功，不允许再次构建！")
		return
	}
	if cntr.IsConfirm == 0 {
		// 如果confirm==0，开发未确认，需要失败处理
		c.SetJson(0, "", "jenkins配置没有得到开发人员的确认，请先确认jenkins构建配置！")
		return
	}

	// 构建
	cntr_base_conf := cfunc.GetUnitInfoById(cntr.UnitId)
	unit_name := strings.ToLower(cntr_base_conf.Unit)
	unit_name = strings.Replace(unit_name, "_", "-", -1)
	unit_name = strings.Replace(unit_name, " ", "", -1)
	job_name := fmt.Sprintf("%s-%s-%s-%s", strings.ToLower(cntr.DeployComp), beego.AppConfig.String("runmode"),
		unit_name, online.Version)
	jenk_opr := jenkins.JenkOpr{
		BaseUrl: cfunc.GetJenkBaseUrl(),
		JobName: job_name,
		ConfigXml: getJenkXml(online, cntr, job_name),
	}
	cntr_build := CntrBuild{
		Opr: jenk_opr,
		OnlineId: online.Id,
		CntrOnlineId: online_cntr.Id,
		ImageUrl: fmt.Sprintf("%s/%s-library/%s:latest", cfunc.GetHarborUrl(), strings.ToLower(cntr.DeployComp), job_name),
	}
	// dr环境的镜像路径不同
	if beego.AppConfig.String("runmode") == "dr" {
		cntr_build.ImageUrl = fmt.Sprintf("%s/%s-pre/%s:latest", cfunc.GetHarborUrl(), strings.ToLower(cntr.DeployComp), job_name)
	}
	// 更改为执行中
	cntr_build.SaveBuildResult(2, 2, 0, "", "")
	high_conc.JobQueue <- &cntr_build
	c.SetJson(1, "", "标准容器进入jenkins构建队列，请耐心等待！")
}

// @Title 标准容器应用的jenkins构建日志，无权限认证
// @Description 标准容器应用的jenkins构建日志，无权限认证
// @Param	online_id	    query	string	true	"标准容器发布单元的id"
// @Success 200 true or false
// @Failure 403
// @router /cntr/jenk-log [get]
func (c *StdCntrOnlineController) CntrJenkLog() {
	online_id, _ := c.GetInt("online_id")
	// 获取发布单元名称
	type InfoName struct {
		Info string `json:"info"`
	}
	var info_name InfoName
	err := initial.DB.Table("online_all_list a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").
		Select("b.info").Where("a.id = ? and a.is_delete=0", online_id).First(&info_name).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	var online_cntr models.OnlineStdCntr
	err = initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=? and is_delete=0", online_id).First(&online_cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if len(online_cntr.JenkinsName) < 10 {
		ret := map[string]interface{}{
			"unit_info": info_name.Info,
			"log": "jenkins的job名称有误！",
		}
		c.SetJson(0, ret, "jenkins的job名称有误！")
		return
	}

	var jenk_opr jenkins.JenkOpr
	jenk_opr.BaseUrl = cfunc.GetJenkBaseUrl()
	jenk_opr.JobName = online_cntr.JenkinsName
	jenk_opr.Init()
	output, err := jenk_opr.GetBuildLog()
	if err != nil {
		beego.Info(err.Error())
		ret := map[string]interface{}{
			"unit_info": info_name.Info,
			"log": err.Error(),
		}
		c.SetJson(0, ret, err.Error())
		return
	}
	//beego.Info(*output)
	ret := map[string]interface{}{
		"unit_info": info_name.Info,
		"log": output,
	}
	c.SetJson(1, ret, "jenkins日志获取成功！")
}

func GetMcpSubinfo(cntr models.UnitConfCntr) models.UnitConfCntr {
	if cntr.McpConfId > 0 {
		var mcp models.UnitConfMcp
		err := initial.DB.Model(models.UnitConfMcp{}).Where("id=?", cntr.McpConfId).First(&mcp).Error
		if err != nil {
			return cntr
		}
		cntr.AppType = mcp.AppType
		cntr.AppSubType = mcp.AppSubType
		cntr.DeployComp = mcp.DeployComp
		return cntr
	}
	return cntr
}

func getJenkXml(online models.OnlineAllList, cntr models.UnitConfCntr, job_name string) string {
	jenk_xml := jenkins_xml.GetBaseJenkinsXml()
	jenk_xml = jenkins_xml.GetBaseJenkinsXmlAssignNode(cntr.JenkinsNode, jenk_xml)
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.git_http_url }}", cntr.GitUrl, -1)
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.git_sha }}", online.CommitId, -1)
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.before_build_shell }}", cntr.BeforeXml, -1)
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.build_shell }}", cntr.BuildXml, -1)
	repo_name := fmt.Sprintf("%s-library/%s", strings.ToLower(cntr.DeployComp), job_name)
	if beego.AppConfig.String("runmode") == "dr" {
		repo_name = fmt.Sprintf("%s-pre/%s", strings.ToLower(cntr.DeployComp), job_name)
	}
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.docker_repo_name }}", repo_name, -1)
	jenk_xml = strings.Replace(jenk_xml, "{{ xml.after_build_shell }}", cntr.AfterXml, -1)
	return jenk_xml
}


// @Title 标准容器应用的更新，和opr的操作一样，只不过传参不同，重写一次
// @Description 标准容器应用的更新，和opr的操作一样，只不过传参不同，重写一次
// @Param	online_id	    query	string	true	"标准容器发布单元的id"
// @Success 200 true or false
// @Failure 403
// @router /cntr/upgrade [post]
func (c *StdCntrOnlineController) CntrUpgrade() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}

	online_id, _ := c.GetInt("online_id")
	var online models.OnlineAllList
	var online_cntr models.OnlineStdCntr
	var cntr models.UnitConfCntr
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=? and is_delete=0", online_id).First(&online).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.OnlineStdCntr{}).Where("online_id=? and is_delete=0", online_id).First(&online_cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	err = initial.DB.Model(models.UnitConfCntr{}).Where("unit_id=? and is_delete=0", online.UnitId).First(&cntr).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}

	// 构建前判断
	if c.Role == "deploy-single" && !controllers.CheckUnitSingleAuth(online.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的部署权限！")
		return
	}
	if online.IsSuccess == 1 {
		// 什么情况下不允许再次构建，发布成功不允许再次构建
		c.SetJson(0, "", "正在发布中或者已发布完成，不允许再次点击！")
		return
	}
	if online.OnlineTime != "" && beego.AppConfig.String("runmode") == "prd" {
		// 只有生产环境才验证发布时间
		ready := common.ReadyToRelease(online.OnlineTime, initial.DateSepLine)
		if !ready {
			c.SetJson(0, "", fmt.Sprintf("发布时间为%s,请在指定时间进行更新!", online.OnlineTime))
			return
		}
	}
	if online_cntr.JenkinsImage == "" {
		c.SetJson(0, "", "要更新的镜像为空，无法更新！")
		return
	}
	if online_cntr.JenkinsSuccess != 1 {
		c.SetJson(0, "", "构建不成功，不允许更新容器！")
		return
	}

	// 正在更新中的应用不允许再次更新
	var cnt int
	initial.DB.Model(models.OprCntrUpgrade{}).Where("result = 2 and unit_id = ?", online.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}
	var cnt2 int
	initial.DB.Model(models.McpUpgradeList{}).Where("result = 2 and unit_id = ?", online.UnitId).Count(&cnt2)
	if cnt2 > 0 {
		c.SetJson(0, "", "镜像正在更新中，不允许再次点击！")
		return
	}

	// 判断是多容器部署还是之前的部署方式

	if cntr.McpConfId > 0 {
		err := McpUpgrade(online_cntr, cntr, c.UserId)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		if cntr.ServiceName == "" {
			c.SetJson(0, "", "容器服务没有关联，不允许升级容器，请先关联容器服务！")
			return
		}
		err, cass_config := operation.GetCaasConfig(cntr.DeployComp)
		if err != nil {
			beego.Info(err.Error())
			c.SetJson(0, "", err.Error())
			return
		}
		// 初始化连接caas，获取相关数据
		opr := caas.CaasOpr{
			AgentConf: cass_config,
			TeamId: cntr.CaasTeam,
			ClustUuid: cntr.CaasCluster,
			StackName: cntr.CaasStack,
			ServiceName: cntr.ServiceName,
		}
		rel_map := operation.RelMap{
			DataTable: "online_std_cntr",
			DataId: online_cntr.Id,
			DataRowName: "opr_cntr_id",
			Flag: true,
		}
		rel_map_2 := operation.RelMap{
			DataTable: "online_all_list",
			DataId: online.Id,
			DataRowName: "is_success",
			Flag: true,
		}

		cntr_upgrade := operation.CntrUpgradeWithImage{
			Opr: opr,
			UnitId: online.UnitId,
			Image: online_cntr.JenkinsImage,
			Operator: c.UserId,
			CntrId: 0,
			Relation: rel_map,
			RelT: rel_map_2,
			SourceId: online.SourceId,
		}
		high_conc.JobQueue <- &cntr_upgrade
	}

	c.SetJson(1, "", "已成功进入队列，请耐心等待容器更新结果！")
}

func McpUpgrade(online_cntr models.OnlineStdCntr, cntr models.UnitConfCntr, userid string) error {
	err, conf := unit_conf.GetMcpConfById(cntr.McpConfId)
	if err != nil {
		return err
	}

	err, agent_conf := operation.GetCaasConfig(conf.DeployComp)
	if err != nil {
		beego.Error(err.Error())
		return err
	}

	if conf.ContainerType == "istio" {
		err := operation.IstioUpgrade(conf, agent_conf, online_cntr.JenkinsImage, userid, online_cntr.Id)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
		return nil
	}
	if conf.ContainerType == "caas" {
		err := operation.CaasUpgrade(conf, agent_conf, online_cntr.JenkinsImage, userid, online_cntr.Id)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
		return nil
	}
	if conf.ContainerType == "rancher" {
		err := operation.RancherUpgrade(conf, agent_conf, online_cntr.JenkinsImage, userid, online_cntr.Id)
		if err != nil {
			beego.Error(err.Error())
			return err
		}
		return nil
	}
	return errors.New("容器类型不支持，无法升级！")
}