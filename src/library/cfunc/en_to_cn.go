package cfunc

import (
	"initial"
	"models"
	"strings"
)

func GetUserCnName(en_name string) string {
	user_cn := ""
	en_arr := strings.Split(en_name, ",")
	for _, v := range en_arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		var user models.UserLogin
		err := initial.DB.Model(models.UserLogin{}).Where("userid = ?", strings.TrimSpace(v)).First(&user).Error
		if err == nil {
			user_cn += user.UserName + ", "
		} else {
			user_cn += strings.TrimSpace(v) + ", "
		}
	}
	return strings.TrimRight(user_cn, ", ")
}

func GetCompCnName(en_name string) string {
	comp_map :=map[string]string{
		"AMC": "招商平安资产管理",
		"CMC": "招商资本",
		"CMCL": "招商局通商融资租赁有限公司",
		"CMEW": "招商局公路网络科技控股股份有限公司",
		"CMFH": "招商金融集团",
		"CMFT": "招商金融科技",
		"CMHD": "招商海达",
		"CMHK": "招商局集团",
		"CMI": "招商局保险有限公司",
		"CMP": "招商港口",
		"CMQHW": "招商前海置业",
		"CMRH": "招商仁和人寿",
		"CMSK": "招商蛇口",
		"CMVC": "招商创投",
		"CSC": "长航集团",
		"CMES": "招商轮船",
		"CMHT": "招商海通",
		"CMIH": "招商局工业集团有限公司",
		"CMPO": "招商积余",
	}
	_, ok := comp_map[en_name]
	if !ok {
		return ""
	}
	return comp_map[en_name]
}

func GetTypeCnName(en_name string) string {
	type_map :=map[string]string{
		"app": "应用",
		"web": "前端",
		"db": "数据库",
		"other": "其它",
	}
	_, ok := type_map[en_name]
	if !ok {
		return ""
	}
	return type_map[en_name]
}

func GetTeamCnName(team_id, company string) string {
	var data []models.CaasConfDetail
	err := initial.DB.Table("conf_caas_detail a").Select("a.*").Joins("left join conf_caas b" +
		" ON a.caas_id = b.id").Where("b.deploy_comp = ? and a.team_id = ?", company, team_id).Limit(1).
		Find(&data).Error
	if err != nil {
		return team_id
	}
	if len(data) > 0 {
		return data[0].TeamName
	} else {
		return team_id
	}

}

func GetClusterCnName(team_id, uuid, company string) string {
	var data []models.CaasConfDetail
	err := initial.DB.Table("conf_caas_detail a").Select("a.*").Joins("left join conf_caas b " +
		"ON a.caas_id = b.id").Where("b.deploy_comp = ? and a.team_id = ? AND cluster_uuid = ?", company,
		team_id, uuid).Limit(1).Find(&data).Error
	if err != nil {
		return uuid
	}
	if len(data) > 0 {
		return data[0].ClusterName
	} else {
		return uuid
	}
}