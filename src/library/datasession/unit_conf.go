package datasession

import (
	"library/common"
	"time"
	"initial"
	"library/cae"
	"strings"
	"fmt"
	"github.com/astaxie/beego"
	"errors"
)

// 存储发布单元同步间隔时间
func PmstUnitSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_pms_unit") || common.GetString(initial.GetCache.Get("sync_pms_unit")) == "" {
		initial.GetCache.Put("sync_pms_unit", time.Now().Format(initial.DatetimeFormat), 500*time.Second)
		flag = 0
	}
	return initial.GetCache.Get("sync_pms_unit"), flag
}

// caas全量数据同步间隔时间
func CaasServiceSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_caas_service") || common.GetString(initial.GetCache.Get("sync_caas_service")) == "" {
		initial.GetCache.Put("sync_caas_service", time.Now().Format(initial.DatetimeFormat), 30*time.Minute)
		flag = 0
	}
	return initial.GetCache.Get("sync_caas_service"), flag
}

// caas单租户数据同步间隔时间
func CaasSingleSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_caas_single_comp") || common.GetString(initial.GetCache.Get("sync_caas_single_comp")) == "" {
		initial.GetCache.Put("sync_caas_single_comp", time.Now().Format(initial.DatetimeFormat), 5*time.Minute)
		flag = 0
	}
	return initial.GetCache.Get("sync_caas_single_comp"), flag
}

// 拉取用户列表间隔时间
func PmstUserSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_pms_user") || common.GetString(initial.GetCache.Get("sync_pms_user")) == "" {
		initial.GetCache.Put("sync_pms_user", time.Now().Format(initial.DatetimeFormat), 600*time.Second)
		flag = 0
	}
	return initial.GetCache.Get("sync_pms_user"), flag
}

// 拉取每天版本列表
func ReleaseRecordSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_release_record") || common.GetString(initial.GetCache.Get("sync_release_record")) == "" {
		initial.GetCache.Put("sync_release_record", time.Now().Format(initial.DatetimeFormat), 100*time.Second)
		flag = 0
	}
	return initial.GetCache.Get("sync_release_record"), flag
}
// caas路由服务配置数据同步间隔时间
func CaasRouteSyncTime() (interface{}, int) {
	flag := 1
	if !initial.GetCache.IsExist("sync_caas_route") || common.GetString(initial.GetCache.Get("sync_caas_route")) == "" {
		initial.GetCache.Put("sync_caas_route", time.Now().Format(initial.DatetimeFormat), 30*time.Minute)
		flag = 0
	}
	return initial.GetCache.Get("sync_caas_route"), flag
}

// harbor镜像同步登录标记
func HarborLoginCheck(exec_user, exec_host, user, pwd string) (bool, error) {
	if !initial.GetCache.IsExist("harbor_login_check") || common.GetString(initial.GetCache.Get("harbor_login_check")) == "" {
		sync_cmd := fmt.Sprintf("source /etc/profile && bash harbor-login.sh %s %s", user, pwd)
		err, sync_msg := cae.ExecCmd(sync_cmd, "/home/deployop", exec_user, exec_host, map[string]interface{}{"timeout": 60})
		if err != nil {
			beego.Error(sync_msg)
			sync_msg_str := strings.Join(sync_msg, "\n")
			n_msg := err.Error() + "\n" + sync_msg_str
			return false, errors.New(n_msg)
		}
		initial.GetCache.Put("harbor_login_check", "true", 4*time.Hour)
	}
	return true, nil
}