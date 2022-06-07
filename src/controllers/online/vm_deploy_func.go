package online

import (
	"errors"
	"github.com/astaxie/beego"
	"initial"
	"library/common"
	"models"
	"strings"
)

func SaveVmDpRest(onlAll models.OnlineAllList, dataAll map[string]interface{}, onlVm models.OnlineStdVM, dataVm map[string]interface{}) error{
	if _, ok := dataVm["upgrade_logs"]; ok {
		logs, ok := dataVm["upgrade_logs"].(string)
		if !ok {
			return errors.New("字段 'upgrade_logs' 数据类型不是 string")
		}
		//if len(logs) > 4999 {
		//	dataVm["upgrade_logs"] = logs[:4999]
		//}
		dataVm["upgrade_logs"] = common.TextPrefixString(logs)
	}
	tx := initial.DB.Begin()
	err1 := tx.Model(onlAll).Updates(dataAll).Error
	if err1 != nil {
		beego.Error(err1.Error())
		tx.Rollback()
		return err1
	}
	err2 := tx.Model(onlVm).Updates(dataVm).Error
	if err2 != nil {
		beego.Error(err2.Error())
		tx.Rollback()
		return err2
	}
	tx.Commit()
	return nil
}

func cmdPrefixParse(cmd string) string {
	if strings.HasPrefix(cmd, "source /etc/profile") || strings.HasPrefix(cmd, "source ~/.bashrc") {
		return cmd
	}
	return "source /etc/profile && source ~/.bashrc && " + cmd
}

// 判断应用包的文件类型是否合法
func GetFileType(filePath, appType string) error {
	subtypes := []string{"zip", "gz", "x-gzip", "x-tgz", "x-archive", "x-tar", "x-bzip2", "gzip", "x-gtar", "tar+gzip", "java-archive"}
	filetype, subtype, err := common.FileType(filePath)
	if err != nil {
		return errors.New("文件类型检测失败：" + err.Error())
	}
	if !common.InList(subtype, subtypes) {
		return errors.New("不支持的部署包文件类型：" + filetype + "/" + subtype)
	}
	return nil
}