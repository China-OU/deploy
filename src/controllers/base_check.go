package controllers

import (
	"models"
	"initial"
	"strings"
	"fmt"
	"library/datalist"
	"time"
	"encoding/json"
	"math/rand"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

// 数据库校验token，先走这个
func DBCheckToken(auth_md5, xauth string) (int, string, *datalist.UserInfo) {
	var token models.UserToken
	var cnt int
	initial.DB.Where("token_md5 = ? and expire > ?", auth_md5, time.Now().Format("2006-01-02 15:04:05")).First(&token).Count(&cnt)
	var info datalist.UserInfo
	if cnt > 0 {
		var user_info models.UserLogin
		var cnt1 int
		initial.DB.Model(models.UserLogin{}).Where("userid = ?", token.UserId).Count(&cnt1).First(&user_info)
		if cnt1 == 0 {
			beego.Info("该用户在用户表中不存在!")
			return 401, "没有登录!", nil
		}
		info.UserLogin = user_info
	} else {
		// 只有校验后端nuc接口，才需要录入token和用户
		flag, user_info := NucCheckToken(auth_md5, xauth)
		if flag == false {
			return 401, "没有登录!", nil
		}
		info.UserLogin = user_info
	}
	info.Token = xauth
	info.Role = GetLoginUserRole(info.Userid)
	login_info_byte, _ := json.Marshal(info)
	initial.GetCache.Put(auth_md5, string(login_info_byte), 5*time.Minute)
	return 1, "校验成功!", &info
}

// nuc校验token
func NucCheckToken(auth_md5, auth string) (bool, models.UserLogin) {
	nuc_base_url := beego.AppConfig.String("nuc_base_url")
	req := httplib.Post(nuc_base_url + "LoginApp/v1/checkToken")
	req.Header("Content-Type", "application/json")
	req.Header("Authorization", auth)
	req.Header("CallerModule", "CMFT_UAD")
	check_body := map[string]string{
		"moduleCode": "CMFT_UAD",
		"tag":"NO_MENU",
	}
	data, _ := json.Marshal(check_body)
	req.Body(data)
	rs, err := req.String()
	//beego.Info(rs)
	if err != nil {
		beego.Error(err.Error())
		return false, models.UserLogin{}
	}
	var check_data datalist.NucBaseRet
	err = json.Unmarshal([]byte(rs), &check_data)
	if err != nil {
		beego.Error(err.Error())
		return false, models.UserLogin{}
	}
	if check_data.State == false {
		beego.Info(check_data.Message)
		return false, models.UserLogin{}
	}
	var info datalist.CheckRet
	check_byte, _ := json.Marshal(check_data.Data)
	err = json.Unmarshal(check_byte, &info)
	if err != nil {
		beego.Error(err.Error())
		return false, models.UserLogin{}
	}
	var user models.UserLogin
	user.Userid = info.AccountId
	user.UserName = info.Name
	user.Company = info.BusinessUnitName
	InsertToken(user, auth_md5, auth)
	go InsertUsername(user)
	return true, user
}

func InsertUsername(user models.UserLogin) {
	rand.Seed(time.Now().Unix())
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	var cnt int
	initial.DB.Model(models.UserLogin{}).Where("userid=?", user.Userid).Count(&cnt)
	if cnt == 0 {
		tx := initial.DB.Begin()
		user.InsertTime = time.Now().Format(initial.DatetimeFormat)
		err := tx.Create(&user).Error
		if err != nil {
			tx.Rollback()
			beego.Error(err.Error())
			return
		}
		tx.Commit()
	}
}

func InsertToken(user models.UserLogin, auth_md5, auth string) {
	tx := initial.DB.Begin()
	var token models.UserToken
	token.UserId = user.Userid
	token.TokenMd5 = auth_md5
	token.Email = user.Userid + "@cmft.com"
	token.Expire = time.Now().Add(4 * time.Hour).Format(initial.DatetimeFormat)
	// 安全加固30分钟，后续注释
	now_date := time.Now().Format(initial.DateFormat)
	if beego.AppConfig.String("runmode") == "prd" && now_date > "20200626" && now_date < "20200720" {
		token.Expire = time.Now().Add(30 * time.Minute).Format(initial.DatetimeFormat)
	}
	// 安全加固30分钟，后续注释
	var info datalist.UserInfo
	info.UserLogin = user
	info.Token = auth
	info_json, _ := json.Marshal(info)
	token.Info = string(info_json)
	err := tx.Create(&token).Error
	if err != nil {
		tx.Rollback()
		beego.Error(err.Error())
	}
	tx.Commit()
}



func GetLoginUserRole(userid string) string {
	// 分五种角色，super-admin, admin, deploy-global, deploy-single, guest
	var data []models.UserRole
	err := initial.DB.Model(models.UserRole{}).Where("username=? and is_delete=0", userid).Find(&data).Error
	if err != nil {
		beego.Error(err.Error())
		return "guest"
	}
	if len(data) == 0 {
		// 非生产环境控制到发布单元权限级别，生产环境无 deploy-single 权限
		if beego.AppConfig.String("runmode") != "prd" {
			flag := CheckSingleAuth(userid)
			if flag {
				return "deploy-single"
			} else {
				return "guest"
			}
		} else {
			return "guest"
		}
	}
	// 有数据记录，直接返回
	var role_arr []string
	for _, v := range data {
		role_arr = append(role_arr, v.Role)
	}
	return strings.Join(role_arr, ",")
}

func CheckSingleAuth(userid string) bool {
	var cnt int
	cond := fmt.Sprintf("(leader='%s' or developer like '%%%s%%' or test like '%%%s%%') and is_offline=0",
		userid, userid+",", userid+",")
	initial.DB.Model(models.UnitConfList{}).Where(cond).Count(&cnt)
	if cnt == 0 {
		return false
	}
	return true
}

func CheckUnitSingleAuth(unit_id int, userid string) bool {
	var cnt int
	cond := fmt.Sprintf("(leader='%s' or developer like '%%%s%%' or test like '%%%s%%') and is_offline=0 and id=%d",
		userid, userid+",", userid+",", unit_id)
	initial.DB.Model(models.UnitConfList{}).Where(cond).Count(&cnt)
	if cnt == 0 {
		return false
	}
	return true
}

func CheckUnitLeaderAuth(unit_id int, userid string) bool {
	var cnt int
	cond := fmt.Sprintf("leader='%s' and is_offline=0 and id=%d", userid, unit_id)
	initial.DB.Model(models.UnitConfList{}).Where(cond).Count(&cnt)
	if cnt == 0 {
		return false
	}
	return true
}