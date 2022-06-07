package harbor

import (
	"strings"
	"controllers"
	"initial"
	"models"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego"
	"encoding/json"
	"library/common"
	"library/cae"
	"fmt"
	"time"
	"library/datasession"
)

type HarborOprController struct {
	controllers.BaseController
}

func (c *HarborOprController) URLMapping() {
	c.Mapping("ImageSync", c.ImageSync)
	c.Mapping("ImageAdd", c.ImageAdd)
	c.Mapping("ImageList", c.ImageList)
	c.Mapping("ImageDel", c.ImageDel)
}

// @Title ImageSync
// @Description 镜像同步，从harbor.uat.cmft.com同步镜像到harbor.cmft.com
// @Param	harbor_id	query	string	true	"同步id"
// @Success 200 {object} {}
// @Failure 403
// @router /image/sync [post]
func (c *HarborOprController) ImageSync() {
	// 同步方案如下，目前选方案一，理由是安全性更好，不需要频繁操作harbor的配置。
	// 方案一：在生产机器，docker pull image, docker tag xx, docker push image，类似于git操作，安全但是稳定性不好。
	// 方案二：使用纯api，修改复制策略，触发复制进行同步。
	if strings.Contains(c.Role, "guest") == true {
		c.SetJson(0, "", "您没有权限操作！")
		return
	}
	now := time.Now()
	harbor_id := c.GetString("harbor_id")
	var hs models.HarborSync
	err := initial.DB.Model(models.HarborSync{}).Where("id=? and is_delete=0", harbor_id).First(&hs).Error
	if err != nil {
		c.SetJson(0, "", err.Error())
		return
	}
	if hs.Result == 1 {
		c.SetJson(0, "", "已同步成功，不允许再次同步！")
		return
	}
	if hs.Result == 2 {
		c.SetJson(0, "", "镜像正在同步，不允许再次点击！")
		return
	}

	// 并发数不能超过5个
	var syn_count int
	initial.DB.Model(models.HarborSync{}).Where("result=2 and is_delete=0").Count(&syn_count)
	if syn_count > 4 {
		c.SetJson(0, "", "最大只允许5个镜像同时同步，请稍后再点！")
		return
	}

	// 校验测试环境harbor镜像是否存在
	req := httplib.Get("http://100.70.42.52/mdeploy/v1/ext/harbor/check")
	req.Header("Authorization", "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM")
	req.Param("image", hs.ImageUrl)
	ret, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	img_check := make(map[string]interface{})
	err = json.Unmarshal([]byte(ret), &img_check)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if common.GetString(img_check["code"]) != "1" {
		c.SetJson(0, "", "harbor-uat不存在该镜像！")
		return
	}

	UpdateSyncResult(2, 0, harbor_id, "", c.UserId)
	// cae同步镜像
	exec_user := "deployop"
	exec_host := "100.65.169.42"
	path := "image-" + common.GetString(time.Now().UnixNano())
	user := "deployop"
	pwd := common.AesDecrypt("ce81cb87a0092b4399dc9037bf8bd0d0")
	// harbor登录session校验
	flag, err := datasession.HarborLoginCheck(exec_user, exec_host, user, pwd)
	if !flag {
		c.SetJson(0, "", err.Error())
		UpdateSyncResult(0, 0, harbor_id, "", c.UserId)
		return
	}
	// 镜像同步
	sync_cmd := fmt.Sprintf("source /etc/profile && bash harbor-image-sync.sh %s %s", path, hs.ImageUrl)
	err, sync_msg := cae.ExecCmd(sync_cmd, "/home/deployop", exec_user, exec_host, map[string]interface{}{"timeout": 600})
	if err != nil {
		beego.Error(sync_msg)
		cost := time.Now().Sub(now).Seconds()
		sync_msg_str := strings.Join(sync_msg, "\n")
		n_msg := err.Error() + "\n" + sync_msg_str
		UpdateSyncResult(0, common.GetInt(cost), harbor_id, n_msg, c.UserId)
		c.SetJson(0, "", err.Error())
		return
	}
	beego.Info(strings.Join(sync_msg, "\n"))

	// 更新数据库中字段
	sync_result := 1
	sync_msg_str := strings.Join(sync_msg, "\n")
	// 隐藏用户名和密码
	sync_msg_str = strings.Replace(sync_msg_str, fmt.Sprintf("%s %s", user, pwd), "user pwd", -1)
	if strings.Contains(sync_msg_str, "Error response from daemon") {
		sync_result = 0
	}
	if strings.Contains(sync_msg_str, "Not logged in to harbor.cmft.com") {
		sync_result = 0
	}
	cost := time.Now().Sub(now).Seconds()
	UpdateSyncResult(sync_result, common.GetInt(cost), harbor_id, sync_msg_str, c.UserId)
	c.SetJson(sync_result, "", "harbor镜像同步完成！")
}

func UpdateSyncResult(result, cost int, id, msg, opr string)  {
	update_map := map[string]interface{}{
		"result": result,
		"message": msg,
		"cost_time": cost,
		"operator": opr,
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.HarborSync{}).Where("id=?", id) .Updates(update_map).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}
