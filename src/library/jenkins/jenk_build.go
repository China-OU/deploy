package jenkins

import (
	"github.com/bndr/gojenkins"
	"encoding/json"
	"github.com/astaxie/beego"
	"errors"
	"library/common"
	"strings"
)

type JenkOpr struct {
	BaseUrl      string
	JobName      string
	ConfigXml    string
	Jenkins      *gojenkins.Jenkins
}

func (c *JenkOpr) Init() {
	defer func() {
		if err := recover(); err != nil {
			beego.Error("Cntr Jenkins Build Panic error:", err)
		}
	}()
	username := ""
	pwd := ""
	runMode := beego.AppConfig.String("runmode")
	if runMode == "dev" || runMode == "di" || runMode == "st" {
 		username = common.JenkinsUserDev
		pwd = common.AesDecrypt(common.JenkinsPwdDev)
	} else if runMode == "prd" || runMode == "dr" {
		username = common.JenkinsUserPrd
		pwd = common.AesDecrypt(common.JenkinsPwdPrd)
	} else {
	//
	}
	c.Jenkins, _ = gojenkins.CreateJenkins(nil, c.BaseUrl, username, pwd).Init()
}

func (c *JenkOpr) CreateJob() error {
	job, err := c.Jenkins.CreateJob(c.ConfigXml, c.JobName)
	if err != nil {
		return err
	}
	ret_byte, _ := json.Marshal(job)
	beego.Info(string(ret_byte))
	return nil
}

func (c *JenkOpr) BuildJob() error {
	job_id, err := c.Jenkins.BuildJob(c.JobName)
	if err != nil {
		return err
	}
	beego.Info(job_id)
	return nil
}

func (c *JenkOpr) DeleteJob() error {
	tag, err := c.Jenkins.DeleteJob(c.JobName)
	if err != nil {
		beego.Error(err.Error())
		if strings.Contains(err.Error(), "404") {
			// 没有构建过的项目无法删除
			return nil
		}
		return err
	}
	if tag == false {
		return errors.New(c.JobName + "删除失败！")
	}
	return nil
}

func (c *JenkOpr) GetLastBuild() (*gojenkins.Build, error) {
	job, err := c.Jenkins.GetJob(c.JobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetLastBuild()
	if err != nil {
		return nil, err
	}
	return build, nil
}

func (c *JenkOpr) GetBuildLog() (*string, error) {
	job, err := c.Jenkins.GetJob(c.JobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetLastBuild()
	if err != nil {
		return nil, err
	}
	resp, err := build.GetConsoleOutputFromIndex(0)
	return &resp.Content, nil
}
