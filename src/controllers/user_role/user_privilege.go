package user_role

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"initial"
	"models"
)

// @Title GetUserPrivilege
// @Description 个人中心==>我的权限
// @Param	title	query	string	false	"下拉选择，开发单元负责人: leader, 开发/测试：dev_test"
// @Param	unit	query	string	false	"发布单元英文名，支持模糊搜索"
// @Param	page	query	string	true	"页数 , 每页固定10条数据"
// @Success 200 {object} []user_role.MyPrivilege
// @Failure 403
// @router /role/my-privilege [get]
func (c *RoleController) GetUserPrivilege() {
	mp := make([]*MyPrivilege, 0)
	count := 0
	// test
	// c.Role = "deploy-single"
	// c.UserId = "xiaojp001"

	switch c.Role {
	case "super-admin":
		mp = append(mp, &MyPrivilege{
			UnitInfo: "全部发布单元",
			AppType: "--",
			UnitRole: "--",
			Privilege: "所有权限",
		})
	case "admin":
		mp = append(mp, &MyPrivilege{
			UnitInfo: "全部发布单元",
			AppType: "--",
			UnitRole: "--",
			Privilege: "所有权限",
		})
	case "deploy-global":
		mp = append(mp, &MyPrivilege{
			UnitInfo: "全部发布单元",
			AppType: "--",
			UnitRole: "--",
			Privilege: "部署",
		})
	case "guest":
		mp = append(mp, &MyPrivilege{
			UnitInfo: "--",
			AppType: "--",
			UnitRole: "--",
			Privilege: "--",
		})
	case "deploy-single":
		title := c.GetString("title")
		unit := c.GetString("unit")
		page, err  := c.GetInt("page")
		if err != nil {
			c.SetJson(0, "", err.Error())
			return
		}

		if mp, count, err = SinglePrivilege(c.UserId, title, unit, page) ; err != nil {
			c.SetJson(0, "", err.Error())
			return
		}
	default:
		c.SetJson(0, "", "无法识别的角色！")
		return
	}

	data := map[string]interface{} {
		"user_info": fmt.Sprintf("%s(%s)", c.Username, c.UserId),
		"role": c.Role,
		"unit_cnt": count,
		"data": mp,
	}
	c.SetJson(1, data, "数据获取成功")
}

func SinglePrivilege(userID, title, unit string,  page int ) (mp[]*MyPrivilege, count int, err error) { // 所有发布单元
	cond := ""
	UnitRole := ""
	pr := ""
	if title == "" {
		cond = fmt.Sprintf("(b.leader='%s' OR b.developer LIKE '%%%s%%' OR b.test LIKE '%%%s%%') AND a.is_delete = 0", userID, userID+",", userID+",")
	}else if title == "leader" {
		cond = fmt.Sprintf("b.leader='%s' AND a.is_delete = 0", userID)
		UnitRole = "发布单元负责人"
		pr = "配置 部署 初始化"
	}else if title == "dev_test"{
		cond = fmt.Sprintf("(b.developer like '%%%s%%' OR b.test like '%%%s%%') AND a.is_delete = 0", userID+",", userID+",")
		UnitRole = "开发/测试"
		pr = "部署"
	}

	if unit != "" {
		cond = fmt.Sprintf("%s AND b.unit  LIKE '%%%s%%'", cond, unit)
	}

	cntrArr := make([]models.UnitConfList, 0)
	vmArr := make([]models.UnitConfList, 0)
	dbArr := make([]models.UnitConfList, 0)

	if err = initial.DB.Table("unit_conf_cntr a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").Select("b.*").Where(cond).Find(&cntrArr).Error ; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, 0, err
		}
	}
	if err = initial.DB.Table("unit_conf_vm a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").Select("b.*").Where(cond).Find(&vmArr ).Error ; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, 0, err
		}
	}
	if err = initial.DB.Table("unit_conf_db a").Joins("LEFT JOIN unit_conf_list b ON a.unit_id = b.id").Select("b.*").Where(cond).Find(&dbArr).Error ; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, 0, err
		}
	}


	c := len(cntrArr)
	v := len(vmArr)
	d := len(dbArr)

	cntrArr = append(cntrArr, vmArr...)
	cntrArr = append(cntrArr, dbArr...)

	for k, v1 := range cntrArr {
		appType := ""
		if k < c {
			appType = "容器部署"
		}
		if k >= c && k < c + v {
			appType = "虚机部署"
		}
		if k >= c+v && k < c+v+d {
			appType = "数据库部署"
		}

		if title == "" {
			if v1.Leader == userID {
				UnitRole = "发布单元负责人"
				pr = "配置 部署 初始化"
			}else {
				UnitRole = "开发/测试"
				pr = "部署"
			}
		}
		mp = append(mp, &MyPrivilege{
			UnitInfo: v1.Info,
			AppType: appType,
			UnitRole: UnitRole,
			Privilege: pr,
		})
	}

	count = len(mp)	// 16

	//if count <= 10 {
	//	return mp, count, nil
	//}
	//
	//if count < 10 * page {
	//	return nil, count, errors.New("请检查指定的分页！")
	//}
	//
	//start := 0
	//end := 0
	//if page > 0 {  // 2
	//
	//	start = (page - 1) * 10
	//	end = page * 10
	//	if end > count {
	//		end = count // [10:16] 第11-16个
	//	}
	//}

	return mp, count, nil
}

type MyPrivilege struct {
	UnitInfo string `json:"unit_info"`
	AppType string `json:"app_type"`
	UnitRole string `json:"unit_role"`
	Privilege string `json:"privilege"`
}
