package unit_conf

import (
	"strings"
	"models"
	"initial"
	"fmt"
	"github.com/jinzhu/gorm"
)

// @Description 多容器平台配置删除
// @Param	id	query	string	true	"数据列的id"
// @Param	container_type	query	string	true	"容器平台类型"
// @Success 200 true or false
// @Failure 403
// @router /mcp/del [post]
func (c *MultiContainerConfController) McpConfDel() {
	if strings.Contains(c.Role, "admin") == false {
		c.SetJson(0, "", "您没有权限操作，请联系管理员进行删除！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	container_type := c.GetString("container_type")

	var conf models.UnitConfMcp
	err = initial.DB.Model(models.UnitConfMcp{}).Where("id=?", id).First(&conf).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	tx := initial.DB.Begin()
	err = tx.Model(models.UnitConfMcp{}).Where("id=?", id).Update("is_delete", 1).Error
	if err != nil {
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	if container_type == "istio" {
		err = tx.Model(models.McpConfIstio{}).Where("mcp_id=?", id).Update("is_delete", 1).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	if container_type == "caas" {
		err = tx.Model(models.McpConfCaas{}).Where("mcp_id=?", id).Update("is_delete", 1).Error
		if err != nil {
			tx.Rollback()
			c.SetJson(0, "", err.Error())
			return
		}
	}
	//if container_type == "openshift" {
	//
	//}

	tx.Commit()
	c.SetJson(1, "", "多容器平台配置删除成功！")
}

// @Description 获取发布单元详细配置
// @Param	unit_id	query	string	true	"发布单元id"
// @Success 200 {object} models.UnitConfMcp
// @Failure 403
// @router /mcp/detail [get]
func (c *MultiContainerConfController) McpConfDetail() {
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	unit_id, _ := c.GetInt("unit_id")
	if unit_id == 0 {
		c.SetJson(0, "", "请选中列表中的发布单元！")
		return
	}
	var mcp models.UnitConfMcp
	err := initial.DB.Model(models.UnitConfMcp{}).Where("unit_id=? and is_delete=0", unit_id).First(&mcp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.SetJson(0, "", "请先在<多容器配置>页面，配置容器服务位置信息！")
			return
		}
		c.SetJson(0, "", err.Error())
		return
	}

	type Ret struct {
		McpConfId    int     `json:"mcp_conf_id"`
		Detail       string  `json:"detail"`
	}
	var ret Ret
	ret.McpConfId = mcp.Id
	if mcp.ContainerType == "caas" {
		err, caas := GetCaasConfDetail(mcp.Id)
		if err != nil {
			ret.Detail = err.Error()
		} else {
			ret.Detail = fmt.Sprintf("项目团队：%s \r\n 容器集群：%s \r\n 容器堆栈：%s \r\n 容器服务名：%s",
				caas.TeamName, caas.ClusterName, caas.StackName, caas.ServiceName)
		}
	}

	if mcp.ContainerType == "rancher" {
		err, rancher := GetRancherConfDetail(mcp.Id)
		if err != nil {
			ret.Detail = err.Error()
		} else {
			ret.Detail = fmt.Sprintf("项目名：%s \r\n 容器堆栈：%s \r\n 容器服务名：%s",
				rancher.ProjectName, rancher.StackName, rancher.ServiceName)
		}
	}

	if mcp.ContainerType == "istio" {
		err, istio := GetIstioConfDetail(mcp.Id)
		if err != nil {
			ret.Detail = err.Error()
		} else {
			ret.Detail = fmt.Sprintf("命名空间：%s \r\n 服务名称：%s",
				istio.Namespace, istio.Deployment)
		}
	}
	c.SetJson(1, ret, "数据获取成功！")
}
