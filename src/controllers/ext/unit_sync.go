package ext

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"initial"
	"library/harbor"
	"models"
	"strings"
)

type MultiEnvConnController struct {
	// 多环境通信，用于生产同步di/st的信息
	beego.Controller
}

func (c *MultiEnvConnController) URLMapping() {
	c.Mapping("Recieve", c.Recieve)
	c.Mapping("HarborImgCheck", c.HarborImgCheck)
	c.Mapping("UnitOnlineQuery", c.UnitOnlineQuery)
	c.Mapping("DbConnCheck", c.DbConnCheck)
}

// @Title 获取发布单元信息，同步到测试环境
// @Description 从prd的部署平台获取发布单元信息，同步到测试环境
// @Param	body	body	models.UnitConfCntr	true	"body形式的数据"
// @Success 200 true or false
// @Failure 403
// @router /unit/sync [post]
func (c *MultiEnvConnController) Recieve() {
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!"}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!"}
		c.ServeJSON()
		return
	}

	var ulist []models.UnitConfList
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &ulist)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": err.Error(), "data": ""}
		c.ServeJSON()
		return
	}

	tx := initial.DB.Begin()
	for _, v := range ulist {
		cnt := 0
		var exist models.UnitConfList
		tx.Model(models.UnitConfList{}).Where("unit = ? and is_offline=0", v.Unit).Count(&cnt).First(&exist)
		if cnt > 0 {
			v.Id = exist.Id
			err = tx.Model(models.UnitConfList{}).Where("unit=? and is_offline=0", v.Unit).Update(v).Error
			if err != nil {
				tx.Rollback()
				beego.Error("基础信息修改错误", err.Error())
				c.Data["json"] = map[string]interface{}{"code": 0, "msg": err.Error(), "data": ""}
				c.ServeJSON()
				return
			}
		} else {
			// id要重新赋值
			v.Id = 0
			err = tx.Create(&v).Error
			if err != nil {
				tx.Rollback()
				beego.Error("基础信息录入错误", err.Error())
				c.Data["json"] = map[string]interface{}{"code": 0, "msg": err.Error(), "data": ""}
				c.ServeJSON()
				return
			}
		}
	}
	tx.Commit()
	c.Data["json"] = map[string]interface{}{"code": 1, "msg": "发布单元同步成功！", "data": ""}
	c.ServeJSON()
}

// @Title 检查镜像是否存在
// @Description 检查镜像是否存在
// @Param	image	query	string	true	"镜像地址"
// @Success 200 true or false
// @Failure 403
// @router /harbor/check [get]
func (c *MultiEnvConnController) HarborImgCheck() {
	header := c.Ctx.Request.Header
	auth := ""
	if header["Authorization"] != nil && len(header["Authorization"]) > 0 {
		auth = header["Authorization"][0]
	} else {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "没有header!"}
		c.ServeJSON()
		return
	}
	if strings.Replace(auth, "Basic ", "", -1) != "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM" {
		c.Data["json"] = map[string]interface{}{"code": 0, "message": "header校验失败!"}
		c.ServeJSON()
		return
	}

	image := c.GetString("image")
	err := harbor.HarborCheckImage(image)
	if err != nil {
		beego.Info(err.Error())
		c.Data["json"] = map[string]interface{}{"code": 0, "msg": err.Error(), "data": ""}
		c.ServeJSON()
		return
	}
	c.Data["json"] = map[string]interface{}{"code": 1, "msg": "镜像存在！", "data": ""}
	c.ServeJSON()
}