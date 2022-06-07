package online

import (
	"time"
	"github.com/astaxie/beego"
	"initial"
	"models"
	"library/common"
	"path"
	"strings"
	"library/database"
	"errors"
	"fmt"
)

type DBAgentOpr struct {
	Online        models.OnlineAllList
	OnlineDetail  models.OnlineDbList
	Conf          models.UnitConfDb
	UnitInfo      models.UnitConfList
	Operator      string
	Err           error
	FreSha        string
}

func (c *DBAgentOpr) PullGitDirAct() bool {
	// 生成当前记录的目录结构, 例如：/app/db_file/oracle/amldb/2944  后面接 /aml/ddl(dml/pkg)
	base_dir := beego.AppConfig.String("file_base_dir")
	unit_en := strings.Replace(strings.ToLower(c.UnitInfo.Unit), "_", "-", -1)
	unit_dir := path.Join(base_dir, c.Conf.Type, unit_en, common.GetString(c.OnlineDetail.OnlineId) )

	// 调用agent接口，拉取文件，返回文件名和文件路径
	agent_opr := database.DBAgentOpr{
		DeployComp: c.Conf.DeployComp,
	}
	if !agent_opr.GetAgentInfo() {
		c.Err = errors.New("agent信息获取失败！")
		return false
	}
	fmap, err := agent_opr.PullGitDir(unit_dir, c.Conf.Type, c.Conf.GitUrl, c.Online.Branch, c.Online.CommitId)
	if err != nil {
		c.Err = err
		return false
	}

	// 调用agent接口，逐个读取文件内容，补齐其它参数
	for k, v := range *fmap {
		if strings.Contains(k, "pkg") == false {
			for _, t := range v {
				if c.CheckDDLInsert(k, t) {
					continue
				}
				content, err := agent_opr.GetDDLSqlInfo(t)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
				err = c.SaveDBLog(k, t, *content)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
			}
		} else {
			for _, t := range v {
				t_arr := strings.Split(t, ".")
				if len(t_arr) != 3 {
					c.Err = errors.New("yml文件中pkg命名为：prx.file_name.sql，命名有误！")
					return false
				}
				if c.CheckPkgInsert(k, t) {
					continue
				}
				content, err := agent_opr.GetPkgSqlInfo(path.Join(unit_dir, GetDbName(c.Conf.GitUrl), "pkg"), t_arr[1]+".sql")
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
				err = c.SavePkgLog(k, t, *content)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
			}
		}
	}

	// 删除失效记录，调用agent接口，删除远端文件，只保留最后10个文件夹
	//  select * from online_db_list where is_dir_clear in (0, 2) order by id asc  > data
	var data []models.OnlineDbList
	err = initial.DB.Table("online_db_list a").Select("a.*").
		Joins("LEFT JOIN online_all_list b on a.online_id = b.id").
		Where("a.is_delete=0 AND b.is_delete=0 AND b.unit_id=? AND a.is_dir_clear=0", c.Online.UnitId).
		Order("b.id asc").Find(&data).Error
	if err != nil {
		beego.Error(err.Error())
		c.Err = err
		return false
	}
	if len(data) > 9 {
		for _, v := range data[0:len(data)-10] {
			rm_dir := path.Join(base_dir, c.Conf.Type, unit_en, common.GetString(v.OnlineId) )
			err := agent_opr.RmAgentDir(rm_dir)
			if err != nil {
				beego.Error(err.Error())
				continue
			}
			c.ClearGitDir(v.OnlineId)
		}
	}
	return true
}

func (c *DBAgentOpr) FreshGitDir() bool {
	// 生成当前记录的目录结构, 例如：/app/db_file/oracle/amldb/2944/aml
	base_dir := beego.AppConfig.String("file_base_dir")
	unit_en := strings.Replace(strings.ToLower(c.UnitInfo.Unit), "_", "-", -1)
	git_dir := path.Join(base_dir, c.Conf.Type, unit_en, common.GetString(c.OnlineDetail.OnlineId), GetDbName(c.Conf.GitUrl))

	// 调用agent接口，拉取文件，返回文件名和文件路径
	agent_opr := database.DBAgentOpr{
		DeployComp: c.Conf.DeployComp,
	}
	if !agent_opr.GetAgentInfo() {
		c.Err = errors.New("agent信息获取失败！")
		return false
	}
	fmap, err := agent_opr.FreshGitDirFunc(git_dir, c.Conf.Type, c.FreSha)
	if err != nil {
		c.Err = err
		return false
	}

	// 调用agent接口，逐个读取文件内容，如果存在，更新记录的sha值，如果不存在，新增;
	for k, v := range *fmap {
		if strings.Contains(k, "pkg") == false {
			for _, t := range v {
				if c.CheckDDLInsert(k, t) {
					// 更新sha值
					c.UpdateDDLSha(k, t)
					continue
				}
				content, err := agent_opr.GetDDLSqlInfo(t)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
				err = c.SaveDBLog(k, t, *content)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
			}
		} else {
			for _, t := range v {
				t_arr := strings.Split(t, ".")
				if len(t_arr) != 3 {
					c.Err = errors.New("yml文件中pkg命名为：prx.file_name.sql，命名有误！")
					return false
				}
				if c.CheckPkgInsert(k, t) {
					c.UpdatePkgSha(k, t)
					continue
				}
				content, err := agent_opr.GetPkgSqlInfo(path.Join(git_dir, "pkg"), t_arr[1]+".sql")
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
				err = c.SavePkgLog(k, t, *content)
				if err != nil {
					beego.Error(err.Error())
					c.Err = err
					return false
				}
			}
		}
	}

	return true
}

func (c *DBAgentOpr) UpdateDirInfo(umap map[string]interface{}) {
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineDbList{}).Where("id=?", c.OnlineDetail.Id).Updates(umap).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

// 更新sha值时更新外部的sha值
func (c *DBAgentOpr) UpdateAllListSha(short string) bool {
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineAllList{}).Where("id=?", c.Online.Id).Updates(map[string]interface{}{
		"commit_id": c.FreSha, "short_commit_id": short}).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}

func (c *DBAgentOpr) ClearGitDir(id int) {
	update_map := map[string]interface{}{
		"is_dir_clear": 1,
	}
	update_map2 := map[string]interface{}{
		"file_content": "文件已删除",
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineDbList{}).Where("online_id=?", id).Updates(update_map).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	err = tx.Model(models.OnlineDbLog{}).Where("online_id=?", id).Updates(update_map2).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

// 用于ddl、dml和trig的录入
func (c *DBAgentOpr) SaveDBLog(stype, path, content string) error {
	// 校验
	path_arr := strings.Split(path, "/")
	file_name := path_arr[len(path_arr)-1]
	if strings.Contains(file_name, " ") {
		return errors.New(file_name + " 文件名命名有空格！")
	}
	if len(strings.Split(file_name, ".")) > 2 {
		return errors.New(file_name + " 有两个.号！")
	}
	prx_user := ""
	command := "source "+path
	if c.Conf.Type == "oracle" {
		prx_user_t, err := GetPrxUser(file_name)
		if err != nil {
			return err
		}
		command = fmt.Sprintf("sh execute_sql.sh %s %s %s", c.Conf.Dbname, path, prx_user_t)
		prx_user = prx_user_t
	}
	if c.Conf.Type == "pgsql" {
		command = fmt.Sprintf("psql postgresql://%s@%s:%d/%s --file=%s --single-transaction",
			c.Conf.Username, c.Conf.Host, c.Conf.Port, c.Conf.Dbname, path)
	}

	var log models.OnlineDbLog
	log.OnlineId = c.Online.Id
	log.FileName = file_name
	log.FileSha = c.Online.CommitId
	if len(c.FreSha) > 6 {
		// 更新sha值时的录入，特殊情况
		log.FileSha = c.FreSha
	}
	log.FilePath = path
	log.FileContent = common.TextPrefixString(content)
	log.IsSuccess = 10
	log.Message = ""
	log.SqlType = stype
	log.ProxyUser = prx_user
	log.Command = command
	log.InsertTime = time.Now().Format(initial.DatetimeFormat)
	log.IsDelete = 0
	log.Operator = c.Operator

	tx := initial.DB.Begin()
	err := tx.Create(&log).Error
	if err != nil {
		beego.Info(log.FileContent)
		beego.Error(err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// 用于pkg的录入
func (c *DBAgentOpr) SavePkgLog(stype, item string, ret_map map[string]string) error {
	if strings.Contains(item, " ") {
		return errors.New(item + " yml命名有空格！")
	}
	itme_arr := strings.Split(item, ".")
	var log models.OnlineDbLog
	log.OnlineId = c.Online.Id
	log.FileName = itme_arr[1]+".sql"
	log.FileSha = c.Online.CommitId
	if len(c.FreSha) > 6 {
		// 更新sha值时的录入，特殊情况
		log.FileSha = c.FreSha
	}
	log.FilePath = ret_map["path"]
	log.FileContent = common.TextPrefixString(ret_map["content"])
	log.IsSuccess = 10
	log.Message = ""
	log.SqlType = stype
	log.ProxyUser = itme_arr[0]
	log.Command = fmt.Sprintf("sh execute_sql.sh %s %s %s", c.Conf.Dbname, ret_map["path"], log.ProxyUser)
	log.InsertTime = time.Now().Format(initial.DatetimeFormat)
	log.IsDelete = 0
	log.Operator = c.Operator

	tx := initial.DB.Begin()
	err := tx.Create(&log).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (c *DBAgentOpr) CheckDDLInsert(stype, name string) bool {
	var cnt int
	err := initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and file_path=? and sql_type=? and is_delete=0",
		c.Online.Id, name, stype).Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		return false
	}
	if cnt > 0 {
		return true
	}
	return false
}

// 更新sha值字段，仅限于更新时用
func (c *DBAgentOpr) UpdateDDLSha(stype, name string) bool {
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineDbLog{}).Where("online_id=? and file_path=? and sql_type=? and is_delete=0",
		c.Online.Id, name, stype).Update("file_sha", c.FreSha).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}

func (c *DBAgentOpr) CheckPkgInsert(stype, name string) bool {
	name_arr := strings.Split(name, ".")
	var cnt int
	err := initial.DB.Model(models.OnlineDbLog{}).Where("online_id=? and file_name=? and sql_type=? and proxy_user=? and is_delete=0",
		c.Online.Id, name_arr[1]+"."+name_arr[2], stype, name_arr[0]).Count(&cnt).Error
	if err != nil {
		beego.Error(err.Error())
		return false
	}
	if cnt > 0 {
		return true
	}
	return false
}

// 更新sha值字段，仅限于更新时用
func (c *DBAgentOpr) UpdatePkgSha(stype, name string) bool {
	name_arr := strings.Split(name, ".")
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineDbLog{}).Where("online_id=? and file_name=? and sql_type=? and proxy_user=? and is_delete=0",
		c.Online.Id, name_arr[1]+"."+name_arr[2], stype, name_arr[0]).Update("file_sha", c.FreSha).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}

func GetPrxUser(name string) (string, error) {
	// vecloud_zsj 为差旅云代理用户
	var SPECIAL_PRX_USER = []string{"I", "EXDATA", "LEPUS", "ORACLE", "C", "BIBAS", "MSTR", "VECLOUD"}
	sname := strings.Replace(name, ".sql", "", -1)
	name_arr := strings.Split(sname, "_")
	if len(name_arr) < 3 {
		return "", errors.New("命名至少有三段，序号_类型_代理用户_xxx.sql")
	}
	if common.InList(strings.ToUpper(name_arr[2]), SPECIAL_PRX_USER) {
		if len(name_arr) < 4 {
			return "", errors.New("命名中有特殊代理用户，必须要有四段命名！")
		}
		return name_arr[2]+"_"+name_arr[3], nil
	} else {
		return name_arr[2], nil
	}
}

func GetDbName(url string) string {
	url_arr := strings.Split(url, "/")
	last := url_arr[len(url_arr)-1]
	return strings.Split(last, ".")[0]
}