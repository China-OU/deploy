package online

import (
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/httplib"
    "library/cae"
    "library/common"
    "library/host"
    "models"
    "os"
    "strings"
    "time"
)

type VmDeploy struct {
    Conf		models.UnitConfVM
    OnlineAll	models.OnlineAllList
    OnlineVM	models.OnlineStdVM
}

func (dp *VmDeploy) Do() {
    var err error
    var start time.Time
    msg := make([]string, 0)
    successAll := 1
    upgradeStu := 1
    errLog := ""

    defer func() {
       if e := recover() ; e != nil {
           beego.Error("VM Deploy Panic error:",e)
           dataAll := map[string]interface{}{
               "error_log": fmt.Sprint("VM Deploy Panic error:",e),
               "is_success": 0,
           }

           dataVM := map[string]interface{}{
               "upgrade_status": 0,
               "upgrade_duration": 0,
               "upgrade_logs": fmt.Sprint("VM Deploy Panic error:",e),
           }

           err = SaveVmDpRest(dp.OnlineAll, dataAll, dp.OnlineVM, dataVM)
           if err != nil {
               beego.Error(err)
           }
           successAll = 0
       }
       if err = dp.SyncPMSStat(successAll); err != nil {
           beego.Error(err)
       }
    }()

    // 发布
    start = time.Now()
    if dp.Conf.AppType == "app" || dp.Conf.AppType == "web" {
        ch := make(chan int, 1)
        after := time.After(25 * time.Minute)
        run_env := beego.AppConfig.String("runmode")
        if run_env != "prd" {
            after = time.After(10 * time.Minute)
        }

        go func() {
            defer func() {
                if e := recover(); e != nil {
                    beego.Error("VM Deploy Panic error:",e)
                    err = errors.New(fmt.Sprint("VM Deploy Panic error:",e))
                    ch <- 2
                }
            }()
            time.Sleep(1 * time.Second)
            err, msg = dp.Deploy()
            ch <- 1
        }()

        select {
        case c := <- ch:
			if err != nil {
				successAll = 0
				upgradeStu = 0
				errLog = err.Error()
                if c == 2 {
                    msg = append(msg, errLog)
                }
			}
			dataAll := map[string]interface{}{
				"error_log": errLog,
				"is_success": successAll,
			}
			costTime := time.Now().Sub(start).Seconds()
			dataVM := map[string]interface{}{
				"upgrade_status": upgradeStu,
				"upgrade_duration": costTime,
				"upgrade_logs": strings.Join(msg, "\r\n"),
			}

			if err = SaveVmDpRest(dp.OnlineAll, dataAll, dp.OnlineVM, dataVM) ; err != nil {
				beego.Info("升级完成后状态更新失败：", err)
			}
        case <- after:
            successAll = 0
            upgradeStu = 0
			dataAll := map[string]interface{}{
				"error_log": "升级执行超时！",
				"is_success":  0,
			}
			dataVM := map[string]interface{}{
				"upgrade_status": 0,
				"upgrade_logs": "升级执行超时！",
			}
			err = SaveVmDpRest(dp.OnlineAll, dataAll, dp.OnlineVM, dataVM)
			if err != nil {
				beego.Error(err)
			}
        }
    } else {
        s := "不支持的发布类型: " + dp.Conf.AppType
        msg = append(msg, s)
        err = errors.New(s)
    }
}

func (dp *VmDeploy) SyncPMSStat(status int) error {
    if dp.OnlineAll.SourceId != "" && dp.OnlineAll.SourceId != "0" {
        req := httplib.Get(beego.AppConfig.String("pms_baseurl") + "/mdp/release/result")
        req.Header("Authorization", "Basic mdeploy_d8c8680d046b1c60e63657deb3ce6d89")
        req.Header("Content-Type", "application/json")
        req.Param("record_id", dp.OnlineAll.SourceId)
        req.Param("result", common.GetString(status))
        _, err := req.String()
        if err != nil {
            return err
        }
    }
    return nil
}

func (dp *VmDeploy) Deploy() (err error, msg []string){
    var log []string

    // 创建本地 & 应用主机临时目录
    msg = append(msg, "1. 创建临时工作目录")
    localWS, remoteWS, err, log := mkWorkDir(&dp.Conf)
    if err != nil {
        beego.Error(err)
        msg = append(msg, log...)
        msg = append(msg, "发布失败！" + err.Error())
        return err, msg
    }
    if len(localWS) == 1 || len(remoteWS) == 1 {
        msg = append(msg, "临时工作目录错误！ " + localWS + " / " + remoteWS)
        return err, msg
    }
    rmLocalWS := fmt.Sprintf("rm -rf %s", localWS)

    // 获取应用部署包
    msg = append(msg, "\n2. 获取应用部署包")
    localFile := localWS + "/" + dp.Conf.Artifact
    runMode := beego.AppConfig.String("runmode")
    jksUser, jksPwd := common.JenkinsUserDev, common.AesDecrypt(common.JenkinsPwdDev)
    if runMode == "prd" || runMode == "dr" {
        jksUser = common.JenkinsUserPrd
        jksPwd = common.AesDecrypt(common.JenkinsPwdPrd)
    }
    if err := common.DownloadWithCurl(dp.OnlineVM.ArtifactURL, localFile, jksUser, jksPwd); err != nil {
        beego.Error()
        msg = append(msg, "应用部署包下载失败！" + err.Error() + "\n发布失败！")
        _, _= common.RunShellCMD(rmLocalWS)
        return err, msg
    }
    msg = append(msg, "部署包下载成功，本地文件: " + localFile)

    // 部署包校验
    msg = append(msg, "\n3. 校验部署包文件完整性")
    if err := GetFileType(localFile, dp.Conf.DeployType); err != nil {
        beego.Error(err)
        msg = append(msg, "部署包校验失败！" + err.Error() + "\n发布失败！")
        _, _= common.RunShellCMD(rmLocalWS)
        return err, msg
    }
    msg = append(msg, "应用部署包校验通过")

    // 数据格式化 & 命令预处理
    dp.Conf.DeployType = strings.ToLower(strings.TrimSpace(dp.Conf.DeployType))
    dp.Conf.CMDStop = cmdPrefixParse(dp.Conf.CMDStop)
    dp.Conf.CMDStartup = cmdPrefixParse(dp.Conf.CMDStartup)
    dp.Conf.AppTempPath = remoteWS
    sj := strings.Split(dp.OnlineVM.JenkinsName, "-")
    sn := sj[2:len(sj) - 1]
    unitName := strings.Join(sn, "-")
    checkProcCmd := "ps -ef | grep -v grep | grep -E '<proc>'"
    switch dp.Conf.DeployType {
    case "jar":
        checkProcCmd = strings.Replace(checkProcCmd, "<proc>", unitName + "|" + dp.Conf.Artifact, -1)
    case "war":
        checkProcCmd = strings.Replace(checkProcCmd, "<proc>", "tomcat|" + unitName + "|" + dp.Conf.Artifact, -1)
    case "py2":
        checkProcCmd = strings.Replace(checkProcCmd, "<proc>", "python27", -1)
    case "py3":
        checkProcCmd = strings.Replace(checkProcCmd, "<proc>", "python36", -1)
    case "ng":
        checkProcCmd = strings.Replace(checkProcCmd, "<proc>", "nginx", -1)
    default:
        msg = append(msg, "不支持的部署类型: " + dp.Conf.DeployType)
        return errors.New("部署类型错误"), msg
    }

    // 串行更新
    msg = append(msg, "\n4. 执行发布")
    hosts := strings.Split(dp.Conf.Hosts, ";")
    for _, h := range hosts {
        err, log = checkServerDir(h, &dp.Conf)
        if err != nil {
            beego.Error(err)
            msg = append(msg, log...)
            _, _ = common.RunShellCMD(rmLocalWS)
            return err, msg
        }

        // 部署包推送
        msg = append(msg, h + ": 推送应用部署包...")
        err, log = cae.TransFile(localFile, dp.Conf.AppTempPath, h)
        if err != nil {
            beego.Error(err)
            msg = append(msg, h + ": 部署包推送失败！\n发布失败！")
            msg = append(msg, log...)
            _, _= common.RunShellCMD(rmLocalWS)
            return err, msg
        }

        // 执行前置命令
        if dp.Conf.CMDPre != "" {
            dp.Conf.CMDPre = cmdPrefixParse(dp.Conf.CMDPre)
            msg = append(msg, h + ": 执行前置命令...")
            msg = append(msg, h + ": 前置命令 " + dp.Conf.CMDPre)
            err, log = cae.ExecCmd(dp.Conf.CMDPre, dp.Conf.AppPath, dp.Conf.AppUser, h)
            msg = append(msg, h + ": 命令执行日志: " + cae.TruncCaeOut(log, 500))
            if err != nil {
                beego.Error(log)
                msg = append(msg, h + ": 前置命令执行失败: " + log[0] + "\n发布失败！")
                _, _= common.RunShellCMD(rmLocalWS)
                return err, msg
            }
            msg = append(msg, h + ": 前置命令执行成功")
        }

        if dp.Conf.NeedReboot == 1 {
            msg = append(msg, h + ": 停应用...")
            msg = append(msg, h + ": 停止命令 " + dp.Conf.CMDStop)
            err, log = cae.ExecCmd(dp.Conf.CMDStop, dp.Conf.AppPath, dp.Conf.AppUser, h)
            msg = append(msg, h + ": 命令执行日志: " + cae.TruncCaeOut(log, 500))
            if err != nil {
                beego.Error(log)
                msg = append(msg, h + ": 停止命令执行失败: " + log[0] + "\n发布失败！")
                _, _= common.RunShellCMD(rmLocalWS)
                return errors.New("停止命令执行失败！"), msg
            }
        }

        // 分类更新
        switch dp.Conf.DeployType {
        case "jar":
            err, log = upgradeSingleFile(h, &dp.Conf)

        case "war":
            err, log = upgradeSingleFile(h, &dp.Conf)

        case "py2":
            err, log = upgradeArchive(h, &dp.Conf)

        case "py3":
            err, log = upgradeArchive(h, &dp.Conf)

        case "ng":
            err, log = upgradeArchive(h, &dp.Conf)

        default:
            s := fmt.Sprintf("不支持的部署类型: %s", dp.Conf.DeployType)
            log = append(log, s)
            err = errors.New(s)
        }
        msg = append(msg, log...)
        if err != nil {
            msg = append(msg, "发布失败！")
            _, _= common.RunShellCMD(rmLocalWS)
            return err, msg
        }

        if dp.Conf.NeedReboot == 1 {
            // 启动应用
            msg = append(msg, h + ": 启动应用...")
            msg = append(msg, h + ": 启动命令 " + dp.Conf.CMDStartup)
            err, log = cae.ExecCmd(dp.Conf.CMDStartup, dp.Conf.AppPath, dp.Conf.AppUser, h)
            msg = append(msg, h + ": 命令执行日志: " + cae.TruncCaeOut(log, 500))
            if err != nil {
                beego.Error(log)
                msg = append(msg, h + ": 启动命令执行失败: " + log[0] + "\n发布失败！")
                _, _= common.RunShellCMD(rmLocalWS)
                return err, msg
            }

            // 检查应用进程，当前仅针对java应用检查
            if dp.Conf.DeployType == "jar" || dp.Conf.DeployType == "war" {
                msg = append(msg, h + ": 检查应用进程")
                msg = append(msg, h + ": 检查命令 " + checkProcCmd)
                err, _ = cae.ExecCmd(checkProcCmd, dp.Conf.AppTempPath, dp.Conf.AppUser, h)
                if err != nil {
                    msg = append(msg, h + ": 应用未在运行，启动失败！")
                    _, _ = common.RunShellCMD(rmLocalWS)
                    return errors.New("应用启动失败！"), msg
                }
            }
        }

        // 执行后置命令
        if dp.Conf.CMDRear != "" {
            dp.Conf.CMDRear = cmdPrefixParse(dp.Conf.CMDRear)
            msg = append(msg, h + ": 执行后置命令...")
            err, log = cae.ExecCmd(dp.Conf.CMDRear, dp.Conf.AppPath, dp.Conf.AppUser, h)
            msg = append(msg, h + ": 命令执行日志: " + cae.TruncCaeOut(log, 500))
            if err != nil {
                beego.Error(log)
                msg = append(msg, h + ": 后置命令执行失败: " + log[0] + "\n发布失败！")
                _, _= common.RunShellCMD(rmLocalWS)
                return err, msg
            }
        }
    }

    // 备份
    msg = append(msg, "\n5. 备份新版本应用")
    _, log = backupNewVersion(unitName, &dp.Conf)
    msg = append(msg, log...)

    // 清理临时目录
    msg = append(msg, "\n6. 清理临时文件")
    msg = append(msg, rmWorkDir(localWS, remoteWS, &dp.Conf)...)
    return nil, msg
}

func mkWorkDir(vmConf *models.UnitConfVM) (localDir, remoteDir string, err error, msg []string) {
    rs := common.GenRandString(8)
    localDir = "/tmp/" + rs
    remoteDir = vmConf.AppTempPath + "/" + rs
    mkLDirCmd := fmt.Sprintf("mkdir -m 777 -p %s", localDir)
    mkRDirCmd := fmt.Sprintf("mkdir -m 777 -p %s", remoteDir)
    if _, err := os.Stat(localDir); os.IsNotExist(err) {
        msg = append(msg, "本地临时目录创建命令 " + mkLDirCmd)
        if _, err = common.RunShellCMD(mkLDirCmd); err != nil {
            msg = append(msg, "本地临时目录创建失败！" + err.Error())
            return "", "", err, msg
        }
        msg = append(msg, "本地临时目录创建成功，" + localDir)
    }

    hosts := strings.Split(vmConf.Hosts, ";")
    for _, h := range hosts {
        msg = append(msg, h + ": 临时目录创建命令 " + mkRDirCmd)
        err, log := cae.ExecCmd(mkRDirCmd, "/tmp", vmConf.AppUser, h)
        if err != nil {
            msg = append(msg, log...)
            msg = append(msg, h + ": 临时目录创建失败！")
            return localDir, "", err, msg
        }
        msg = append(msg, h + ": 临时目录创建成功.")
    }
    return localDir, remoteDir, nil, msg
}

func rmWorkDir(localDir, remoteDir string, vmConf *models.UnitConfVM) (msg []string) {
    if !host.IsWSValid(localDir) {
        msg = append(msg, "本地目录 " + localDir + "  不合法，只允许 /app/ 和 /tmp/ 目录下的操作。")
        return
    }
    if !host.IsWSValid(remoteDir) {
        msg = append(msg, "本地目录 " + remoteDir + "  不合法，只允许 /app/ 和 /tmp/ 目录下的操作。")
        return
    }
    rmLDirCmd := fmt.Sprintf("rm -rf %s", localDir)
    rmRDirCmd := fmt.Sprintf("rm -rf %s", remoteDir)
    msg = append(msg, "本地临时目录删除命令 " + rmLDirCmd)
    if _, err := common.RunShellCMD(rmLDirCmd); err != nil {
        msg = append(msg, "本地临时目录删除失败！请手动处理")
    }
    hosts := strings.Split(vmConf.Hosts, ";")
    for _, h := range hosts {
        msg= append(msg, h + ": 临时目录删除命令 " + rmRDirCmd)
        err, log := cae.ExecCmd(rmRDirCmd, vmConf.AppTempPath, vmConf.AppUser, h)
        if err != nil {
            msg = append(msg, log...)
            msg = append(msg, h + ": 临时目录删除失败！请手动处理" + err.Error())
        }
    }
    msg = append(msg, "临时文件清理完成\n发布成功")
    return msg
}

func upgradeSingleFile(ip string, vmConf *models.UnitConfVM) (err error, msg []string) {
    // 针对war包，删除应用目录下所有内容
    if strings.ToLower(vmConf.DeployType) == "war" {
        rmCmd := fmt.Sprintf("rm -rf %s/*", vmConf.AppPath)
        msg = append(msg, ip + ": 旧版文件删除命令 " + rmCmd)
        err, log := cae.ExecCmd(rmCmd, vmConf.AppTempPath, vmConf.AppUser, ip)
        if err != nil {
            msg = append(msg, ip + ": 旧版本应用文件删除失败！" + err.Error())
            msg = append(msg, log...)
            return err, msg
        }
    }
    // 更新应用文件
    msg = append(msg, ip + ": 更新应用文件...")
    dpFile := vmConf.AppTempPath + "/" + vmConf.Artifact
    replaceCmd := fmt.Sprintf("mv -f %s %s", dpFile, vmConf.AppPath)
    msg = append(msg, ip + ": 更新命令 " + replaceCmd)
    err, log := cae.ExecCmd(replaceCmd, vmConf.AppTempPath, vmConf.AppUser, ip)
    if err != nil {
        msg = append(msg, ip + ": 应用文件更新失败！" + err.Error())
        msg = append(msg, log...)
        return err, msg
    }
    // 文件权限修复
    chCmd := fmt.Sprintf("chown -R %s. %s", vmConf.AppUser, vmConf.AppPath)
    msg = append(msg, ip + ": 文件权限修复命令 " + chCmd)
    err, log = cae.ExecCmd(chCmd, vmConf.AppTempPath, "root", ip)
    if err != nil {
        msg = append(msg, ip + ": 应用文件权限修复失败！")
        msg = append(msg, log...)
        return err, msg
    }
    return nil, msg
}

func upgradeArchive(ip string, vmConf *models.UnitConfVM) (err error, msg []string) {
    msg = append(msg, ip + ": 更新应用文件...")
    unpackCmd := ""

    // 删除旧版本文件
    rmCmd := fmt.Sprintf("rm -rf %s/*", vmConf.AppPath)
    msg = append(msg, ip + ": 旧版文件删除命令 " + rmCmd)
    err, log := cae.ExecCmd(rmCmd, vmConf.AppTempPath, vmConf.AppUser, ip)
    if err != nil {
        msg = append(msg, ip + ": 旧版本应用文件删除失败！" + err.Error())
        msg = append(msg, log...)
        return err, msg
    }

    // 解压新版本应用文件到应用目录
    if strings.HasSuffix(vmConf.Artifact, ".tar.gz") {
        unpackCmd = fmt.Sprintf("tar xzf %s -C %s > /dev/null", vmConf.Artifact, vmConf.AppPath)
    }else if strings.HasSuffix(vmConf.Artifact, ".zip") {
        unpackCmd = fmt.Sprintf("unzip %s -d %s > /dev/null", vmConf.Artifact, vmConf.AppPath)
    }else {
        s := "文件格式错误，目前仅支持tar.gz和zip文件！"
        msg = append(msg, s)
        return  errors.New(s), msg
    }

    msg = append(msg, ip + ": 解压命令 " + unpackCmd)
    err, log = cae.ExecCmd(unpackCmd, vmConf.AppTempPath, vmConf.AppUser, ip)
    if err != nil {
        msg = append(msg, ip + ": 部署包解压失败！" + err.Error())
        msg = append(msg, log...)
        return err, msg
    }
    err, log = checkServerDir(ip, vmConf)
    if err != nil {
        msg = append(msg, cae.TruncCaeOut(log, 200))
        return err, msg
    }

    // 文件权限修复
    chCmd := fmt.Sprintf("chown -R %s. %s", vmConf.AppUser, vmConf.AppPath)
    msg = append(msg, ip + ": 文件权限修复命令 " + chCmd)
    err, log = cae.ExecCmd(chCmd, vmConf.AppTempPath, "root", ip)
    if err != nil {
        msg = append(msg, ip + ": 应用文件权限修复失败！")
        msg = append(msg, log...)
        return err, msg
    }

    // 依赖包预先下载到项目本地，发布时连同代码一同部署到服务器，然后服务器上从本地目录刷新依赖
    if vmConf.AppType == "web" && strings.Contains(vmConf.DeployType, "py") {
        msg = append(msg, ip + ": 当配置项中应用类型选择为web时，不需要刷新python依赖，直接进行更新。此时依赖刷新需要上虚机手动刷新！")
        return nil, msg
    }
    if vmConf.DeployType == "py2" || vmConf.DeployType == "py3" {
        refPipCmd := "source /etc/profile && source ~/.bashrc && <pipVer> install -r requirements.txt --no-index --find-links=Packages/"
        if vmConf.DeployType == "py2" {
            refPipCmd = strings.Replace(refPipCmd, "<pipVer>", "pip27", -1)
        }
        if vmConf.DeployType == "py3" {
            refPipCmd = strings.Replace(refPipCmd, "<pipVer>", "pip36", -1)
        }
        msg = append(msg, ip + ": 刷新 pip 依赖")
        msg = append(msg, ip + ": 刷新命令 " + refPipCmd)
        err, log := cae.ExecCmd(refPipCmd, vmConf.AppPath, vmConf.AppUser, ip)
        msg = append(msg, cae.TruncCaeOut(log, 200))
        if err != nil {
            msg = append(msg, ip + ": 依赖刷新失败！")
            return errors.New("python依赖刷新失败"), msg
        }
    }

    return nil, msg
}

func backupNewVersion(appName string, vmConf *models.UnitConfVM) (bakFile string, msg []string) {
    bakCmd := ""
    timeNow := strings.Split(time.Now().Format(time.RFC3339), "+")[0]
    bakFile = ""
    if vmConf.DeployType == "jar" || vmConf.DeployType == "war" {
        bakFile = fmt.Sprintf("%s/%s.%s.%s", vmConf.AppBackupPath, appName, vmConf.DeployType, timeNow)
        bakCmd = fmt.Sprintf("cp %s/%s %s", vmConf.AppPath, vmConf.Artifact, bakFile)
    }
    if vmConf.DeployType == "py2" || vmConf.DeployType == "py3" || vmConf.DeployType == "ng" {
        bakFile = fmt.Sprintf("%s/%s.tar.gz.%s", vmConf.AppBackupPath, appName, timeNow)
        bakCmd = fmt.Sprintf("tar czf %s -C %s . --exclude logs --exclude *.log", bakFile, vmConf.AppPath)
    }
    hosts := strings.Split(vmConf.Hosts, ";")
    for _, h := range hosts {
        msg = append(msg, h + ": 备份命令 " + bakCmd)
        err, log := cae.ExecCmd(bakCmd, vmConf.AppTempPath, vmConf.AppUser, h)
        if err != nil {
            msg = append(msg, h+": 新版本应用文件备份失败！" + err.Error())
            msg = append(msg, log...)
        }
        msg = append(msg, h+": 备份成功，备份文件：" + bakFile)
    }
    return
}

func checkServerDir(ip string, vmConf *models.UnitConfVM) (err error, msg []string) {
    msg = append(msg, ip + ": 应用目录/文件校验")
    checkCmd := fmt.Sprintf("[[ -d %s ]] && [[ -d %s ]]", vmConf.AppPath, vmConf.AppBackupPath)
    msg = append(msg, ip + ": 目录校验命令 " + checkCmd)
    err, _ = cae.ExecCmd(checkCmd, vmConf.AppPath, vmConf.AppUser, ip)
    if err != nil {
        msg = append(msg, ip + ": 应用目录或备份目录不存在！")
        return err, msg
    }

    if vmConf.DeployType == "py2" || vmConf.DeployType == "py3" {
        checkCmd = fmt.Sprintf("ls %s | grep -E '*\\.sh|*\\.py' && ls %s | grep requirements.txt && ls %s | grep Packages", vmConf.AppPath, vmConf.AppPath, vmConf.AppPath)
    }
    if vmConf.DeployType == "ng" {
        checkCmd = fmt.Sprintf("ls %s | grep -E '*\\.html|*\\.htm|*\\.js'", vmConf.AppPath)
    }
    if vmConf.DeployType == "jar" || vmConf.DeployType == "war" {
        checkCmd = fmt.Sprintf("ls %s | grep -E '%s'", vmConf.AppPath, vmConf.Artifact)
    }
    msg = append(msg, ip + ": 应用文件校验命令 " + checkCmd)
    err, _ = cae.ExecCmd(checkCmd, vmConf.AppPath, vmConf.AppUser, ip)
    if err != nil {
        msg = append(msg, ip + ": 配置的应用目录下没有找到应用文件，请检查目录或发布的文件是否正确！")
        return err, msg
    }

    return nil, msg
}