package cae

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/thedevsaddam/gojsonq"
	"library/common"
	"strings"
	"time"
)

// 远程执行命令，返回命令的原始输出内容
func ExecCmdOutput(cmd, path, user, host string) (err error, msg []string, out map[string]interface{}) {
	err, token := getToken()
	if err != nil {
		beego.Error(err)
		return err, []string{"获取cae token失败！"}, nil
	}
	cv := CaeInitVar()
	hosts := strings.Split(host, ";")
	for _, h := range hosts {
		if !common.IsValidIP(h) {
			return errors.New("IP地址不合法"), []string{"IP地址不合法"}, nil
		}
	}
	timeout := 150
	reqData := make(map[string]interface{})
	reqData["dst_ips"] = hosts
	reqData["remote_comand"] = fmt.Sprintf("/usr/bin/env bash -c \"%s\"", cmd)
	reqData["remote_path"] = path
	reqData["run_user"] = user
	reqData["timeout"] = timeout
	reqData["is_async"] = false
	reqDataBytes, _ := json.Marshal(reqData)

	req := httplib.Post(cv.BaseUrl + cv.CmdApi)
	req.Header("Content-Type", "application/json")
	req.Header("X-Auth-Token", token)
	req.SetTimeout(time.Duration(timeout) * time.Second, time.Duration(timeout) * time.Second)
	// 重试三次
	out = make(map[string]interface{})
	flag := false
	for i := 0; i < 3; i++ {
		req.Body(reqDataBytes)
		resp, err := req.Bytes()
		if err != nil {
			beego.Error(fmt.Sprintf("第 %d 次远程执行命令调用失败，命令：%s，远程主机：%s", i + 1, cmd, host), err)
			msg = append(msg, fmt.Sprintf("第%d次接口调用失败。错误为：%s", i+1, err.Error()))
			continue
		}
		var res map[string]interface{}
		err = json.Unmarshal(resp, &res)
		if err != nil {
			beego.Error("远程执行命令接口数据解析失败：", err)
			msg = append(msg, fmt.Sprintf("第%d次接口调用失败。CAE返回数据解析失败！", i + 1))
			continue
		}
		if res == nil {
			beego.Error("远程执行命令返回的数据为空")
			msg = append(msg, fmt.Sprintf("第%d次接口调用失败。远程执行命令返回的数据为空！", i + 1))
			continue
		}

		jsonQ := gojsonq.New(gojsonq.SetSeparator("_")).FromInterface(res)
		env := []string{"test", "prod"}
		for _, e := range env {
			// gojsonq 在调用一次 Find() 后，内部会记录当前位置，下次调用将会继续
			// 如需从头开始查找，需手动调用 gojsonq.Reset()
			queryEnv := fmt.Sprintf("data_[0]_%s", e)
			envData := jsonQ.Find(queryEnv)
			jsonQ.Reset()
			if envData != nil {
				for _, h := range hosts {
					queryHost := fmt.Sprintf("data_[0]_%s_%s", e, h)
					hostData := jsonQ.Find(queryHost)
					jsonQ.Reset()
					out[h] = hostData
				}
				flag = true
				break
			}
		}
		if flag {
			msg = append(msg, fmt.Sprintf("命令执行成功。 命令为：%s", cmd))
			return nil, msg, out
		}
	}
	err = errors.New(fmt.Sprintf("CAE三次调用失败，执行命令：%s，远程主机：%s", cmd, host))
	msg = append(msg, fmt.Sprintf("CAE三次调用失败，执行命令：%s，远程主机：%s", cmd, host))
	return err, msg, out
}

// 适用于单台主机的远程命令执行，直接输出命令stdout
func ExecCmdSingleHost(cmd, path, user, host string) (error, []string, string) {
	if !common.IsValidIP(host) {
		return errors.New("IP地址不合法"), []string{}, ""
	}
	err, msg, output := ExecCmdOutput(cmd, path, user, host)
	if err != nil {
		beego.Error(err.Error())
		return err, msg, ""
	}
	outBytes, err := json.Marshal(output)
	if err != nil {
		beego.Error(err.Error())
		return err, msg, ""
	}
	outJSON := gojsonq.New(gojsonq.SetSeparator("_")).FromString(string(outBytes))
	stdout := fmt.Sprintf("%s", outJSON.Find(host + "_return"))
	outJSON.Reset()
	stat := fmt.Sprintf("%s", outJSON.Find(host + "_status"))
	if stat != "ok" {
		return errors.New("远程命令执行失败"), msg, stdout
	}
	return nil, msg, stdout
}

