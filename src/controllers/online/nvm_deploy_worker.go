package online

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"initial"
	"library/cae"
	"library/common"
	"models"
	"os"
	"path"
	"strings"
	"time"
)

type NvmDeploy struct {
	Conf        models.UnitConfNvm
	Online      models.OnlineAllList
	Nvm         models.OnlineNvm
	TmpDir      string
	ErrLog      string
	ShellLog    string
}

func (c *NvmDeploy) Do() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Nvm Deploy Panic error:", err)
		}
	}()
	timeout := time.After(15 * time.Minute)
	run_env := beego.AppConfig.String("runmode")
	if run_env != "prd" {
		// 测试环境缩短超时时间
		timeout = time.After(8 * time.Minute)
	}
	result_ch := make(chan bool, 1)
	go func() {
		c.DeployProcessing()
		err, msg := c.NvmDeployFunc()
		if err != nil {
			c.ErrLog = err.Error() + "\n" + msg
			result_ch <- false
		} else {
			c.ErrLog = msg
			result_ch <- true
		}
	}()
	select {
	case result := <-result_ch:
		beego.Info("执行完成")
		c.rmTmpDir()
		c.SaveExecResult(result, "")
	case <-timeout:
		beego.Info("执行超时")
		c.rmTmpDir()
		c.SaveExecResult(false, "timeout")
	}
}

/* 第一步：生成临时目录，下载应用包；执行SHA256校验；
 * 第二步：（生成shell脚本）执行shell脚本;
 */
func (c *NvmDeploy) NvmDeployFunc() (error, string) {
	// 创建临时工作目录
	tmp_dir_path, err, msg := c.mkTmpDir()
	if err != nil {
		beego.Error(err.Error())
		return err, msg
	}
	c.TmpDir = tmp_dir_path
	// 下载应用包
	file_path := path.Join(tmp_dir_path, c.Nvm.FileName)
	download_cmd := fmt.Sprintf("curl -o %s -k %s", file_path, c.Nvm.FileAddr)
	out, err := common.RunShellCMD(download_cmd)
	if err != nil {
		beego.Error(err.Error())
		return err, out
	}
	// 下载包sha256校验
	check_cmd := fmt.Sprintf("sha256sum %s", file_path)
	out, err = common.RunShellCMD(check_cmd)
	beego.Info(out)
	if err != nil {
		beego.Error(err.Error())
		return err, out
	}
	if !strings.Contains(out, strings.ToLower(c.Nvm.Sha256)) {
		beego.Error(out)
		return errors.New("下载包的sha256值和录入的不一致，请校验"), ""
	}
	// 将下载的包设置为可读
	if err := os.Chmod(file_path, 0644); err != nil {
		return err, "本地的包设置为可读模式失败"
	}

	host_arr := strings.Split(c.Nvm.Host, ";")
	for _, v := range host_arr {
		if strings.TrimSpace(v) == "" {
			continue
		}
		// 传包
		err, log := cae.TransFile(file_path, c.TmpDir, v)
		if err != nil {
			beego.Error(err.Error())
			return err, strings.Join(log, "\n")
		}
		// 生成shell脚本，此步可选
		if c.Conf.PathOrGene == "generate" {
			beego.Info("此处是根据数据库的脚本内容，在对应的shell_path生成shell文件")
		}
		// 串行执行shell脚本
		cmd_deploy := fmt.Sprintf("source /etc/profile && source ~/.bashrc && sh deploy.sh %s", file_path)
		err, log, output := cae.ExecCmdSingleHost(cmd_deploy, path.Dir(c.Conf.ShellPath), c.Conf.AppUser, v)
		shell_log := ""
		for _, k := range log {
			shell_log += fmt.Sprintf("%s: %s\n", v, k)
		}
		shell_log += fmt.Sprintf("%s: 部署脚本输出结果为\n\n", v)
		if output == "" {
			output = "执行脚本报错，无输出！"
		}
		shell_log += output + "\n\n"
		c.ShellLog += shell_log
		if err != nil {
			beego.Error(err.Error())
			return err, "shell脚本执行失败"
		}
	}
	return nil, "非标虚机发布成功"
}

func (c *NvmDeploy) DeployProcessing() {
	err := initial.DB.Model(models.OnlineAllList{}).Where("id=?", c.Online.Id).Update("is_success", 2).Error
	if err != nil {
		beego.Error(err.Error())
		return
	}
}

func (c *NvmDeploy) SaveExecResult(result bool, rtype string) {
	int_result := 0
	if result {
		int_result = 1
	}
	error_log := c.ErrLog
	shell_log := c.ShellLog
	if rtype == "timeout" {
		error_log = "执行超时。\n" + c.ErrLog
		shell_log = "执行超时。\n" + c.ShellLog
	}
	update_all := map[string]interface{}{
		"is_processing": 0,
		"is_success": int_result,
		"error_log": error_log,
	}
	update_nvm := map[string]interface{}{
		"shell_log": common.TextPrefixString(shell_log),
	}
	tx := initial.DB.Begin()
	err := tx.Model(models.OnlineAllList{}).Where("id=?", c.Online.Id).Updates(update_all).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	err = tx.Model(models.OnlineNvm{}).Where("id=?", c.Nvm.ID).Updates(update_nvm).Error
	if err != nil {
		beego.Error(err.Error())
		tx.Rollback()
		return
	}
	tx.Commit()
}

//  子函数
func (c *NvmDeploy) mkTmpDir() (string, error, string) {
	rs := common.GenRandString(8)
	tmp_dir := "/tmp/" + rs
	mkdir_cmd := fmt.Sprintf("mkdir -m 777 -p %s", tmp_dir)
	_, err := os.Stat(tmp_dir)
	if os.IsNotExist(err) {
		out, err := common.RunShellCMD(mkdir_cmd)
		if err != nil {
			return "", err, out
		}
	}

	hosts := strings.Split(c.Nvm.Host, ";")
	for _, v := range hosts {
		if strings.TrimSpace(v) == "" {
			continue
		}
		err, log := cae.ExecCmd(mkdir_cmd, "/tmp", c.Conf.AppUser, v)
		if err != nil {
			msg := fmt.Sprintf("主机%s创建临时目录失败：\n", v)
			msg += strings.Join(log, "\n")
			return tmp_dir, err, msg
		}
	}
	return tmp_dir, nil, "临时目录创建成功"
}

func (c *NvmDeploy) rmTmpDir() {
	if strings.HasPrefix(c.TmpDir, "/tmp/") == false {
		beego.Error("临时目录删除有误，临时目录异常")
		return
	}
	// 删除本地临时目录
	rmdir_cmd := fmt.Sprintf("rm -rf %s", c.TmpDir)
	out, err := common.RunShellCMD(rmdir_cmd)
	if err != nil {
		beego.Error(err.Error(), out)
	}
	// 删除远端临时目录
	hosts := strings.Split(c.Nvm.Host, ";")
	for _, v := range hosts {
		err, log := cae.ExecCmd(rmdir_cmd, "/tmp", c.Conf.AppUser, v)
		if err != nil {
			beego.Error(err.Error(), strings.Join(log, ", "))
		}
	}
}