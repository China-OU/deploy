package operation

import (
	"time"
	"library/caas"
	"github.com/astaxie/beego"
	"initial"
	"models"
	"library/common"
	"github.com/astaxie/beego/httplib"
)

// cntr操作的进程
type CntrUpgradeWithImage struct {
	Opr      caas.CaasOpr   // 基础配置
	UnitId   int            // 更新的发布单元
	Image    string         // 更新后的镜像
	Operator string         // 操作人员
	CntrId   int            // 操作记录ID
	Relation RelMap         // 关联记录
	RelT     RelMap         // 关联表2
	SourceId string         // 外部来源id
}

type RelMap struct {
	DataTable   string
	DataId      int
	DataRowName string
	Flag        bool
}

type RetHttp struct {
	DataId    string
	Flag      bool
}

func (self *CntrUpgradeWithImage) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Cntr Upgrade Panic error:", err)
		}
	}()
	timeout := time.After(20 * time.Minute)
	run_env := beego.AppConfig.String("runmode")
	if run_env != "prd" {
		// 测试环境缩容器更新短超时时间
		timeout = time.After(8 * time.Minute)
	}
	result_ch := make(chan bool, 1)
	go func() {
		result := self.UpgradeImage()
		result_ch <- result
	}()
	select {
	case <-result_ch:
		beego.Info("执行完成")
		self.RelResult()
	case <-timeout:
		beego.Info("执行超时")
		self.SaveExecResult(false, run_env + "环境执行超时，容器状态异常，请上caas平台查看", 20 * 60)
		self.RelResult()
	}
}

func (self *CntrUpgradeWithImage) UpgradeImage() bool {
	err, detail := self.Opr.GetServiceDetail()
	if err != nil {
		// 数据库没记录，无法记录到数据库中
		beego.Error(err.Error())
		return false
	}
	if detail.Image == "" {
		beego.Error("容器当前镜像无法获取！")
		return false
	}
	id := self.InsertRecord(detail.Image)
	if id == 0 {
		beego.Error(err.Error())
		return false
	}
	self.CntrId = id
	// 关联外部表
	self.RelRecord()
	// 判断是镜像是否同名
	if self.Image == detail.Image {
		self.SameImage()
	} else {
		self.DiffImage()
	}
	return true
}

func (self *CntrUpgradeWithImage) SameImage() {
	// 删除实例重新创建，过程较长
	err, instance_list := self.Opr.GetInstanceList()
	if err != nil {
		beego.Info(err.Error())
		self.SaveExecResult(false, err.Error(), 0)
		return
	}
	if len(instance_list) == 0 {
		self.SaveExecResult(false, "无法获取pod数据", 0)
		return
	}
	start := time.Now()
	beego.Info(start.Format(initial.DatetimeFormat))
	for _, v := range instance_list {
		beego.Info(v.Name)
		err := self.Opr.DelCaasInstance(v.Name)
		if err != nil {
			beego.Info(err.Error())
			msg := err.Error()
			if msg == "" {
				msg = "pod删除失败"
			}
			self.SaveExecResult(false, msg, 0)
			return
		}
		// 加入延时，避免直接拿到结果；第一次长一点
		time.Sleep(40 * time.Second)
		ec := 0
		for {
			ec += 1
			if ec > 50 {
				// 设置执行次数
				cost_time := time.Now().Sub(start).Seconds()
				self.SaveExecResult(false, "执行超时，容器状态异常，请上容器平台检查！", common.GetInt(cost_time))
				return
			}
			err, detail := self.Opr.GetServiceDetail()
			if err != nil {
				beego.Info(err.Error())
				self.SaveExecResult(false, err.Error(), 0)
				return
			}
			if detail.State != "active" {
				// 同名镜像有时候会出现upgraded状态
				if detail.State == "upgraded" {
					self.Opr.FinishUpgradeService()
					time.Sleep(20 * time.Second)
				}
				beego.Info("还未升级完成，请等待20秒。。。")
				time.Sleep(20 * time.Second)
			} else {
				break
			}
		}
	}
	beego.Info(time.Now().Format(initial.DatetimeFormat))
	beego.Info(self.Opr.ServiceName + "同镜像名更新成功！")
	cost_time := time.Now().Sub(start).Seconds()
	self.SaveExecResult(true, self.Opr.ServiceName + "同镜像名更新成功！", common.GetInt(cost_time))
	return
}

func (self *CntrUpgradeWithImage) DiffImage() {
	start := time.Now()
	err := self.Opr.UpgradeService(self.Image)
	if err != nil {
		beego.Info(err.Error())
		self.SaveExecResult(false, err.Error(), 0)
		return
	}
	// 升级是瞬时操作，要判断是否升级完成
	time.Sleep(40 * time.Second)
	ec := 0
	for {
		ec += 1
		if ec > 50 {
			// 设置执行次数
			cost_time := time.Now().Sub(start).Seconds()
			self.SaveExecResult(false, "执行超时，容器状态异常，请上容器平台检查！", common.GetInt(cost_time))
			return
		}
		beego.Info("还未升级完成，请等待20秒。。。")
		time.Sleep(20 * time.Second)
		err, detail := self.Opr.GetServiceDetail()
		if err != nil {
			beego.Info(err.Error())
			self.SaveExecResult(false, err.Error(), 0)
			return
		}
		if detail.State == "upgraded" {
			break
		}
	}
	err = self.Opr.FinishUpgradeService()
	if err != nil {
		beego.Info(err.Error())
		self.SaveExecResult(false, err.Error(), 0)
		return
	}
	time.Sleep(5 * time.Second)
	cost_time := time.Now().Sub(start).Seconds()
	self.SaveExecResult(true, self.Opr.ServiceName + "镜像更新成功！", common.GetInt(cost_time))
	return
}

func (self *CntrUpgradeWithImage) SaveExecResult(result bool, msg string, cost int) {
	int_result := 0
	if result {
		int_result = 1
	}
	update_map := map[string]interface{}{
		"result": int_result,
		"message": msg,
		"cost_time": cost,
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.OprCntrUpgrade{}).Where("id=?", self.CntrId).Updates(update_map).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (self *CntrUpgradeWithImage) InsertRecord(old_image string) int {
	var cntr models.OprCntrUpgrade
	cntr.UnitId = self.UnitId
	cntr.OldImage = old_image
	cntr.NewImage = self.Image
	cntr.Result = 2
	cntr.Operator = self.Operator
	now := time.Now()
	today := now.Format(initial.DateFormat)
	if now.Hour() < 4 {
		today = now.AddDate(0, 0, -1).Format(initial.DateFormat)
	}
	cntr.OnlineDate = today
	cntr.CostTime = 0
	cntr.InsertTime = now.Format(initial.DatetimeFormat)
	cntr.SourceId = self.SourceId
	tx := initial.DB.Begin()
	err := tx.Create(&cntr).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return 0
	}
	tx.Commit()
	return cntr.Id
}

func (self *CntrUpgradeWithImage) RelRecord() {
	if self.Relation.Flag == false {
		// 未关联外部系统
		return
	}
	if self.Relation.DataTable == "" || self.Relation.DataRowName == "" || self.Relation.DataId <= 0 || self.CntrId <=0 {
		beego.Error("数据有误，无法进行外部关联！")
		return
	}

	tx := initial.DB.Begin()
	err := tx.Table(self.Relation.DataTable).Where("id=?", self.Relation.DataId).Update(self.Relation.DataRowName,
		self.CntrId).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

func (self *CntrUpgradeWithImage) RelResult() {
	if self.RelT.Flag == false {
		// 未关联外部系统
		return
	}
	if self.RelT.DataTable == "" || self.RelT.DataRowName == "" || self.RelT.DataId <= 0 || self.CntrId <=0 {
		beego.Error("数据有误，无法进行外部关联！")
		return
	}

	tx := initial.DB.Begin()
	var opr models.OprCntrUpgrade
	err := tx.Model(models.OprCntrUpgrade{}).Where("id=?", self.CntrId).First(&opr).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	update_map := map[string]interface{}{
		self.RelT.DataRowName: opr.Result,
		"error_log": opr.Message,
	}
	err = tx.Table(self.RelT.DataTable).Where("id=?", self.RelT.DataId).Updates(update_map).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()

	// 将结果回传给pms
	if opr.SourceId != "" && opr.SourceId != "0" {
		req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/release/result")
		req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
		req.Header("Content-Type", "application/json")
		req.Param("record_id", opr.SourceId)
		req.Param("result", common.GetString(opr.Result))
		_, err := req.String()
		if err != nil {
			beego.Error(err.Error())
		}
	}
}
