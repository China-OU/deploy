package ext

import (
	"controllers"
	"encoding/json"
	"strings"
	"models"
	"initial"
	"github.com/astaxie/beego"
	"library/common"
	"time"
	"fmt"
	"library/cae"
	"library/datasession"
	"github.com/astaxie/beego/httplib"
)

type ExtCpdsFuncController struct {
	controllers.BaseUrlAuthController
}

/* 提供给cpds相关接口，包括但不限于以下功能：
 * harbor镜像同步，多容器平台升级，虚机应用升级，数据库部署等
 * 实现全流程自助部署，将部署平台作为一个后台服务
 */
func (c *ExtCpdsFuncController) URLMapping() {
	c.Mapping("HarborImageSync", c.HarborImageSync)
	c.Mapping("HarborImageSyncPoll", c.HarborImageSyncPoll)
	c.Mapping("CntrServiceUpgrade", c.CntrServiceUpgrade)
	c.Mapping("CntrServicePoll", c.CntrServicePoll)
	c.Mapping("CntrServiceDetail", c.CntrServiceDetail)
}

// @Title 镜像同步通用外部接口，接口内部会做权限校验
// @Description 镜像同步通用外部接口，接口内部会做权限校验
// @Param	body	body	ext.ImageSyncInput 	true	"body形式的数据，发布单元id名和镜像"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 {object} {}
// @Failure 403
// @router /cpds/image/sync [post]
func (c *ExtCpdsFuncController) HarborImageSync() {
	var input ImageSyncInput
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &input)
	if err != nil {
		beego.Error(err.Error())
		c.SetJson(0, "", err.Error())
		return
	}
	if strings.TrimSpace(input.Image) == "" || strings.TrimSpace(input.RecordId) == "" {
		c.SetJson(0, "", "输入参数不能为空！")
		return
	}

	if c.Role != "admin" {
		// 如果是受限用户，需要做权限认定，非cpds用户需要做限制，白名单还是什么策略，后续可以补充。
	}

	// 并发数不能超过5个
	var syn_count int
	var c_count int
	initial.DB.Model(models.HarborSync{}).Where("result=2 and is_delete=0").Count(&syn_count)
	if syn_count > 4 {
		c.SetJson(0, "", "最大只允许5个镜像同时同步，请稍后再点！")
		return
	}
	initial.DB.Model(models.HarborSync{}).Where("source_id=? and is_delete=0 and result in (1, 2)", input.RecordId).Count(&c_count)
	if c_count > 4 {
		c.SetJson(0, "", "当前镜像已同步成功或正在同步，请稍后再试！")
		return
	}

	// 录入数据库，校验在子函数做
	var is models.HarborSync
	is.ImageUrl = input.Image
	is.Result = 2
	is.Message = ""
	is.CostTime = 0
	is.ApplyPerson = c.Username
	is.Operator = c.Username
	is.InsertTime = time.Now().Format(initial.DatetimeFormat)
	is.IsDelete = 0
	is.SourceId = input.RecordId
	tx := initial.DB.Begin()
	err = tx.Create(&is).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		c.SetJson(0, "", err.Error())
		return
	}
	tx.Commit()
	go syncExecUnit(input.Image, is.Id)
	c.SetJson(1, "", "镜像同步已进入队列，请耐心等待执行结果！")
}

type ImageSyncInput struct {
	Image       string  `json:"image"`
	RecordId    string  `json:"record_id"'`          // 唯一字符串类型
}

func syncExecUnit(image string, record_id int) {
	// 校验测试环境harbor镜像是否存在
	req := httplib.Get("http://100.70.42.52/mdeploy/v1/ext/harbor/check")
	req.Header("Authorization", "mdeploy_IpFhvFjiQpV65PjIUywc3VHDjC0Wo9EM")
	req.Param("image", image)
	ret, err := req.String()
	if err != nil {
		beego.Error(err.Error())
		UpdateSyncResult(0, 0, record_id, err.Error())
		return
	}
	img_check := make(map[string]interface{})
	err = json.Unmarshal([]byte(ret), &img_check)
	if err != nil {
		UpdateSyncResult(0, 0, record_id, err.Error())
		return
	}
	if common.GetString(img_check["code"]) != "1" {
		UpdateSyncResult(0, 0, record_id, "harbor-uat不存在该镜像！")
		return
	}

	// harbor登录session校验
	exec_user := "deployop"
	exec_host := "100.65.169.42"
	user := "deployop"
	pwd := common.AesDecrypt("ce81cb87a0092b4399dc9037bf8bd0d0")
	flag, err := datasession.HarborLoginCheck(exec_user, exec_host, user, pwd)
	if !flag {
		UpdateSyncResult(0, 0, record_id, err.Error())
		return
	}

	now := time.Now()
	path := "image-" + common.GetString(time.Now().UnixNano())
	sync_cmd := fmt.Sprintf("source /etc/profile && bash harbor-image-sync.sh %s %s", path, image)
	err, sync_msg := cae.ExecCmd(sync_cmd, "/home/deployop", exec_user, exec_host, map[string]interface{}{"timeout": 600})
	if err != nil {
		beego.Error(sync_msg)
		cost := time.Now().Sub(now).Seconds()
		sync_msg_str := strings.Join(sync_msg, "\n")
		n_msg := err.Error() + "\n" + sync_msg_str
		UpdateSyncResult(0, common.GetInt(cost), record_id, n_msg)
		return
	}

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
	UpdateSyncResult(sync_result, common.GetInt(cost), record_id, sync_msg_str)
}

func UpdateSyncResult(result, cost, id int, msg string)  {
	update_map := map[string]interface{}{
		"result": result,
		"message": msg,
		"cost_time": cost,
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

// @Title 轮询接口，查询执行结果
// @Description 轮询接口，查询执行结果
// @Param	record_list	query	string	true	"记录列表，比如 aaabb,ccc,dd,ee"
// @Param	ak	query	string	true	"用户名"
// @Param	ts	query	string	true	"时间戳"
// @Param	sn	query	string	true	"加密串"
// @Param	debug	query	string	true	"调试模式"
// @Success 200 true or false
// @Failure 403
// @router /cpds/image-sync/poll [get]
func (c *ExtCpdsFuncController) HarborImageSyncPoll() {
	record_list := c.GetString("record_list")
	r_list := strings.Split(record_list, ",")

	type PollRet struct {
		RecordId    string    `json:"record_id"`
		Result      int       `json:"result"`
		Msg         string    `json:"msg"`
		Cost        int       `json:"cost"`
	}

	var ret []PollRet
	for _, v := range r_list {
		record_id := strings.Trim(v, " ")
		var per PollRet
		per.RecordId = record_id
		var harbor_sync models.HarborSync
		err := initial.DB.Model(models.HarborSync{}).Where("source_id=?", record_id).Order("field (result, 1, 2, 10, 0)").
			First(&harbor_sync).Error
		if err != nil {
			beego.Error(err.Error())
			per.Msg = err.Error()
			per.Result = 100
			ret = append(ret, per)
			continue
		}
		per.Result = harbor_sync.Result
		per.Msg = harbor_sync.Message
		per.Cost = harbor_sync.CostTime
		ret = append(ret, per)
	}
	c.SetJson(1, ret, "结果查询成功！")
}
