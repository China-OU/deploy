package tasks

import (
	"github.com/astaxie/beego"
	"library/common"
	"models"
	"runtime"
	"time"
	"initial"
	"controllers/caas_conf"
	"strings"
)

// 同步容器平台数据信息
func CassDataSync() error {
	//获取panic
	defer func() {
		if panic_err := recover(); panic_err != nil {
			var buf []byte = make([]byte, 1024)
			c := runtime.Stack(buf, false)
			beego.Error("CassDataSync 获取容器信息错误:", panic_err)
			beego.Error("CassDataSync 获取容器信息详细信息:", string(buf[0:c]))
		}
	}()

	beego.Info("获取容器信息 Start." + time.Now().Format("2006-01-02 15:04:05"))
	sync_data()
	beego.Info("获取容器信息 End." + time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

func sync_data() error {
	var ca_conf []models.CaasConf
	cond := "is_delete=0"
	err := initial.DB.Model(models.CaasConf{}).Where(cond).Find(&ca_conf).Error
	if err != nil {
		beego.Error(err.Error())
		return nil
	}
	// 每周日，清除垃圾数据，先is_delete=1，数据同步过来后，置为0
	if time.Now().Weekday().String() == "Sunday" {
		tx := initial.DB.Begin()
		err := tx.Model(models.CaasConfDetail{}).Update("is_delete", 1).Error
		if err != nil {
			tx.Rollback()
			beego.Error(err.Error())
			return nil
		}
		tx.Commit()
	}

	for _, v := range ca_conf {
		// 不同的网络区域，不同的agent，可以并发，后续再处理
		caas_id := v.Id
		// 获取team的列表
		team_list, err := caas_conf.GetCaasTeamList(v.AgentIp, v.AgentPort)
		if err != nil {
			beego.Error(err.Error())
			beego.Error("team获取错误，ip和端口为：", v.AgentIp, v.AgentPort)
			continue
		}

		// 获取cluster的列表
		clust_list, err := caas_conf.GetCaasClustList(v.AgentIp, v.AgentPort)
		if err != nil {
			beego.Error(err.Error())
			beego.Error("cluster获取错误，ip和端口为：", v.AgentIp, v.AgentPort)
			continue
		}

		// 获取stack列表
		for _, i := range team_list {
			// 去掉压测数据
			if i.Name == "pressbasedata" || strings.Contains(i.Name, "_INNER") == true {
				continue
			}
			for _, j := range clust_list {
				stack_list, err := caas_conf.GetCaasStackList(common.GetString(i.Id), j.Uuid, v.AgentIp, v.AgentPort)
				if err != nil {
					beego.Error(err.Error())
					continue
				}
				for _, k := range stack_list {
					// 录入单元
					service_list, err := caas_conf.GetCaasServiceList(common.GetString(i.Id), j.Uuid, k.Name, v.AgentIp, v.AgentPort)
					if err != nil {
						beego.Error(err.Error())
						continue
					}
					for _, t := range service_list {
						err := caas_conf.InsertOrUpdateCaasDetail(caas_id, i, j, k, t)
						if err != nil {
							beego.Error(err.Error())
							continue
						}
						time.Sleep(10 * time.Nanosecond)
					}
				}
			}
		}

		// 配置表同步更新时间
		tx := initial.DB.Begin()
		err = tx.Model(models.CaasConf{}).Where("id=?", v.Id).Update("detail_sync_time",
			time.Now().Format(initial.DatetimeFormat)).Error
		if err != nil {
			beego.Error(err.Error())
			tx.Rollback()
			return nil
		}
		tx.Commit()

	}
	return nil
}