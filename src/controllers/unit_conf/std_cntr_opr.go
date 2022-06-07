package unit_conf

import (
	"controllers"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/git"
	"models"
	"strings"
	"time"
)

// 标准容器发布单元录入和修改
// @Title 标准容器发布单元录入和修改
// @Description 标准容器发布单元录入和修改。从发布单元列表选取数据，同时作相关信息确认和维护
// @Param	body	body	models.UnitConfCntr	true	"body形式的数据，涉及密码要加密"
// @Success 200 true or false
// @Failure 403
// @router /cntr/edit [post]
func (c *StdCntrConfController) CntrEdit() {
	if c.Role == "guest" {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	if beego.AppConfig.String("runmode") == "prd" && strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "生产环境权限收缩，您没有权限操作！")
		return
	}
	var cntr models.UnitConfCntr
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &cntr)
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if c.Role == "deploy-single" && !controllers.CheckUnitLeaderAuth(cntr.UnitId, c.UserId) {
		c.SetJson(0, "", "您没有此发布单元的编辑权限，请联系发布单元负责人进行操作！")
		return
	}
	// 校验
	flag, msg := CntrCheck(cntr)
	if flag == false {
		c.SetJson(0, "", msg)
		return
	}
	var cnt int
	initial.DB.Model(models.UnitConfCntr{}).Where("id != ? and unit_id = ? and is_delete = 0", cntr.Id, cntr.UnitId).Count(&cnt)
	if cnt > 0 {
		c.SetJson(0, "", "此发布单元已经创建，不能重复创建。如果需要创建备份发布单元，请在主列表页录入新发布单元！")
		return
	}

	// 录入或者更新
	tx := initial.DB.Begin()
	input_id := cntr.Id
	if cntr.Id > 0 {
		// 只更新五个字段
		update_map := map[string]interface{}{
			"git_id": cntr.GitId,
			"git_unit": cntr.GitUnit,
			"git_url": cntr.GitUrl,
			"mcp_conf_id": cntr.McpConfId,
		}
		err = tx.Model(models.UnitConfCntr{}).Where("id=?", cntr.Id).Updates(update_map).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	} else {
		cntr.InsertTime = time.Now().Format(initial.DatetimeFormat)
		cntr.JenkinsNode = ""
		cntr.IsConfirm = 0
		cntr.CpdsFlag = 0
		err = tx.Create(&cntr).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	tx.Commit()
	ret_msg := "标准容器发布单元信息新增成功！"
	if input_id > 0 {
		ret_msg = "信息维护成功！"
	}
	c.SetJson(1, "", ret_msg)
}

func CntrCheck(cntr models.UnitConfCntr) (bool, string) {
	// git信息确认
	git_info := git.SearchByGitId(cntr.GitId)
	if cntr.GitUnit != git_info.PathWithNamespace {
		return false, "发布单元路径不对！"
	}
	if cntr.GitUrl != git_info.HTTPURLToRepo {
		return false, "发布单元url不对！"
	}
	// 检查发布单元是否正确
	unit_flag := CheckBasicInfo(fmt.Sprintf("id = %d", cntr.UnitId))
	if !unit_flag {
		return false, "基表中没有该发布单元！"
	}
	if cntr.McpConfId <= 0 {
		return false, "配置需要关联多容器配置！"
	}
	// 确认发布单元是否匹配
	err, mcp := GetMcpConfById(cntr.McpConfId)
	if err != nil {
		return false, "多容器配置不存在"
	}
	if mcp.UnitId != cntr.UnitId {
		return false, "构建配置的发布单元和容器配置的发布单元不一样，请保持一致！"
	}
	return true, ""
}

func CheckBasicInfo(cond string) bool {
	// 不作租户校验，因为租户下面有子网
	if strings.HasPrefix(cond, "dumd_comp_en") {
		return true
	}
	var cnt int
	initial.DB.Model(models.UnitConfList{}).Where("is_offline=0 and " + cond).Count(&cnt)
	if cnt > 0 {
		return true
	} else {
		return false
	}
}
