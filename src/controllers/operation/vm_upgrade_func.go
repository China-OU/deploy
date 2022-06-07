package operation

import (
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "library/cae"
    "library/common"
    "library/host"
    "models"
    "net/url"
    "os"
    "strings"
    "time"
)

type UnitVMUpgrade struct {
    OnlineVM        models.OnlineStdVM
    ConfigVM        models.UnitConfVM
    OldVersion      models.OprVMVersion
    NewVersion      models.OprVMVersion
    UpgradeRecord   models.OprVMUpgrade
    TempArtifact    string              // 部署包暂存路径
    Logs            []string
    Operator        string
}

// 主机命令预处理
func (u *UnitVMUpgrade) PreProcessingCmd() {
    u.ConfigVM.CMDPre       = preProcessingCmd(u.ConfigVM.CMDPre)
    u.ConfigVM.CMDRear      = preProcessingCmd(u.ConfigVM.CMDRear)
    u.ConfigVM.CMDStop      = preProcessingCmd(u.ConfigVM.CMDStop)
    u.ConfigVM.CMDStartup   = preProcessingCmd(u.ConfigVM.CMDStartup)
}

func preProcessingCmd(cmd string) string {
    if strings.HasPrefix(cmd, "source /etc/profile") || strings.HasPrefix(cmd, "source ~/.bashrc") {
        return cmd
    }
    return "source /etc/profile && source ~/.bashrc && " + cmd
}

// 创建临时工作目录
func (u *UnitVMUpgrade) MkWorkspace() error {
    ws := "/tmp/" + common.GenRandString(8)
    mkdirCmd := fmt.Sprintf("mkdir -m 777 -p %s", ws)
    u.Logs = append(u.Logs, "创建临时工作目录", "工作目录创建命令：" + mkdirCmd)
    // 创建本地临时目录
    if _, err := os.Stat(ws); os.IsNotExist(err) {
        if _, err := common.RunShellCMD(mkdirCmd); err != nil {
            u.Logs = append(u.Logs, "本地工作目录创建失败：" + err.Error())
            return err
        }
        u.Logs = append(u.Logs, "本地工作目录创建成功：" + ws)
    }

    // 创建远端工作目录
    hosts := strings.Split(u.ConfigVM.Hosts, ";")
    for _, h := range hosts {
        err, log := cae.ExecCmd(mkdirCmd, "/tmp", u.ConfigVM.AppUser, h)
        if err != nil {
            u.Logs = append(u.Logs, log...)
            u.Logs = append(u.Logs, h + ": 工作目录创建失败")
            return err
        }
        u.Logs = append(u.Logs, h + ": 工作目录创建成功")
    }

    // 更新发布单元配置中的临时目录，后续流程要用到
    u.ConfigVM.AppTempPath = ws
    return nil
}

// 下载部署包
func (u *UnitVMUpgrade) DownloadArtifact() error {
    u.Logs = append(u.Logs, "获取应用部署包：" + u.OnlineVM.ArtifactURL)
    err := checkURL(u.OnlineVM.ArtifactURL)
    if err != nil {
        u.Logs = append(u.Logs, "下载链接校验失败：" + err.Error())
        return err
    }
    u.TempArtifact = u.ConfigVM.AppTempPath + "/" + u.ConfigVM.Artifact
    // 下载
    if strings.Contains(u.OnlineVM.ArtifactURL, "pan.cmrh.com") {
        // 预处理pan.cmrh.com下载链接
        if !strings.HasSuffix(u.OnlineVM.ArtifactURL, "/download") {
            if strings.HasSuffix(u.OnlineVM.ArtifactURL, "/") {
                u.OnlineVM.ArtifactURL += "download"
            } else {
                u.OnlineVM.ArtifactURL += "/download"
            }
        }
        err := common.DownloadWithHttp(u.OnlineVM.ArtifactURL, u.TempArtifact, 10 * time.Minute, nil)
        if err != nil {
            u.Logs = append(u.Logs, "应用部署包获取失败：" + err.Error())
            return err
        }
    } else {
        err := common.DownloadWithCurl(u.OnlineVM.ArtifactURL, u.TempArtifact, "ops", common.AesDecrypt("53cce73712461546be79a85f31f88ce5"))
        if err != nil {
            u.Logs = append(u.Logs, "应用部署包获取失败：" + err.Error())
            return err
        }
    }
    return nil
}

// 校验部署包
func (u *UnitVMUpgrade) CheckArtifact() error {
    u.Logs = append(u.Logs, "部署包文件类型检查")
    filetype, subtype, err := common.FileType(u.TempArtifact)
    if err != nil {
        errMsg := "文件类型检测失败：" + err.Error()
        u.Logs = append(u.Logs, errMsg)
        return errors.New(errMsg)
    }
    switch strings.ToLower(u.ConfigVM.DeployType) {
    case "jar":
        if subtype != "zip" {
            errMsg := "下载的文件不是 Jar 包：" + filetype + "/" + subtype
            u.Logs = append(u.Logs, errMsg)
            return errors.New(errMsg)
        }
    case "war":
        if subtype != "zip" {
            errMsg := "下载的文件不是 War 包：" + filetype + "/" + subtype
            u.Logs = append(u.Logs, errMsg)
            return errors.New(errMsg)
        }
    case "py2":
        if subtype != "zip" && subtype != "gz" {
            errMsg := "下载的文件不是 zip 或 gz 包：" + filetype + "/" + subtype
            u.Logs = append(u.Logs, errMsg)
            return errors.New(errMsg)
        }
    case "py3":
        if subtype != "zip" && subtype != "gz" {
            errMsg := "下载的文件不是 zip 或 gz 包：" + filetype + "/" + subtype
            u.Logs = append(u.Logs, errMsg)
            return errors.New(errMsg)
        }
    case "ng":
        if subtype != "zip" && subtype != "gz" {
            errMsg := "下载的文件不是 zip 或 gz 包：" + filetype + "/" + subtype
            u.Logs = append(u.Logs, errMsg)
            return errors.New(errMsg)
        }
    default:
        errMsg := "应用类型错误，当前仅支持：jar / war / py2 / py3 / ng"
        u.Logs = append(u.Logs, errMsg)
        return errors.New(errMsg)
    }
    return nil
}

// 上传部署包
func (u *UnitVMUpgrade) UploadArtifact() error {
    u.Logs = append(u.Logs, "推送部署包")
    hosts := strings.Split(u.ConfigVM.Hosts, ";")
    for _, h := range hosts{
        err, _ := cae.TransFile(u.TempArtifact, u.ConfigVM.AppTempPath, h)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, h + ": 部署包推送失败")
            return err
        }
    }
    // 更新文件属主&权限
    u.Logs = append(u.Logs, "更新部署包属主&权限")
    chCmd := fmt.Sprintf("chown -R %s. %s && chmod 644 %s", u.ConfigVM.AppUser, u.ConfigVM.AppTempPath, u.TempArtifact)
    u.Logs = append(u.Logs, "更新权限命令：" + chCmd)
    for _, h := range hosts {
        err, log := cae.ExecCmd(chCmd, u.ConfigVM.AppTempPath, "root", h)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, h + ": 部署包权限修改失败：" + cae.TruncCaeOut(log, 500))
            return err
        }
    }
    return nil
}

// 校验应用目录
// checkMode: dir 只检查目录，file 只检查应用文件，full 都检查
func (u *UnitVMUpgrade) CheckServerDir(host, checkMode string) error {
    var checkCmd string
    if checkMode == "full" || checkMode == "dir" {
        u.Logs = append(u.Logs, host + ": 应用目录校验")
        checkCmd = fmt.Sprintf("[[ -d %s ]] && [[ -d %s ]]", u.ConfigVM.AppPath, u.ConfigVM.AppBackupPath)
        u.Logs = append(u.Logs, host + ": 目录校验命令：" + checkCmd)
        err, log := cae.ExecCmd(checkCmd, "/tmp", u.ConfigVM.AppUser, host)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, host + ": 目录校验失败：" + cae.TruncCaeOut(log, 500))
            return err
        }
    }
    if checkMode == "full" || checkMode == "file" {
        u.Logs = append(u.Logs, host + ": 应用文件校验")
        if strings.HasPrefix(u.ConfigVM.DeployType, "py") {
            checkCmd = fmt.Sprintf("ls %s | grep -E '*\\.sh|*\\.py' && ls %s | grep requirements.txt && ls %s | grep Packages",
                u.ConfigVM.AppPath, u.ConfigVM.AppPath, u.ConfigVM.AppPath)
        }
        if u.ConfigVM.DeployType == "ng" {
            checkCmd = fmt.Sprintf("ls %s | grep -E '*\\.html|*\\.htm|*\\.js'", u.ConfigVM.AppPath)
        }
        if u.ConfigVM.DeployType == "jar" || u.ConfigVM.DeployType == "war" {
            checkCmd = fmt.Sprintf("ls %s | grep -E '%s'", u.ConfigVM.AppPath, u.ConfigVM.Artifact)
        }
        u.Logs = append(u.Logs, host + ": 应用文件校验命令：" + checkCmd)
        err, _ := cae.ExecCmd(checkCmd, "/tmp", u.ConfigVM.AppUser, host)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, host + ": 配置的应用目录下没有找到应用文件，请检查目录或发布的文件是否正确！")
            return err
        }
    }
    return nil
}

// 执行前置命令
func (u *UnitVMUpgrade) ExecPreCmd(host string) error {
    u.Logs = append(u.Logs, "前置命令：" + u.ConfigVM.CMDPre)
    err, log := cae.ExecCmd(u.ConfigVM.CMDPre, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
    u.Logs = append(u.Logs, host + ": 命令执行日志：" + cae.TruncCaeOut(log, 500))
    if err != nil {
        beego.Error(log)
        u.Logs = append(u.Logs, host + ": 前置命令执行失败")
        return err
    }
    return nil
}

// 停应用
func (u *UnitVMUpgrade) StopApp(host string) error {
    if u.ConfigVM.CMDStop == "" {
        errMsg := "应用停止命令不能为空"
        u.Logs = append(u.Logs, errMsg)
        return errors.New(errMsg)
    }
    u.Logs = append(u.Logs, host + ": 停止应用命令：" + u.ConfigVM.CMDStop)
    err, log := cae.ExecCmd(u.ConfigVM.CMDStop, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
    u.Logs = append(u.Logs, host + ": 命令执行日志：" + cae.TruncCaeOut(log, 500))
    if err != nil {
        beego.Error(log)
        u.Logs = append(u.Logs, host + ": 停止应用命令执行失败")
        return err
    }
    return nil
}

// 更新应用文件
func (u *UnitVMUpgrade) UpgradeArtifact(host string) error {
    u.Logs = append(u.Logs, host + ": 更新应用文件")
    // 单文件更新
    if u.ConfigVM.DeployType == "jar" || u.ConfigVM.DeployType == "war" {
        replaceCmd := fmt.Sprintf("mv -f %s %s", u.TempArtifact, u.ConfigVM.AppPath)
        u.Logs = append(u.Logs, host + ": 文件更新命令：" + replaceCmd)
        err, log := cae.ExecCmd(replaceCmd, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, host + ": 文件更新失败：" + cae.TruncCaeOut(log, 500))
            return err
        }
    }
    // 归档包更新
    if strings.HasPrefix(u.ConfigVM.DeployType, "py") || u.ConfigVM.DeployType == "ng" {
        // 删除旧版文件
        if u.ConfigVM.AppPath == "" {
            return errors.New(host + ": 配置的应用目录为空")
        }
        rmCmd := fmt.Sprintf("rm -rf %s/*", u.ConfigVM.AppPath)
        u.Logs = append(u.Logs, host + ": 旧版应用删除命令：" + rmCmd)
        err, log := cae.ExecCmd(rmCmd, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, host + ": 删除命令执行失败：" + cae.TruncCaeOut(log, 500))
            return err
        }
        // 解压新版文件
        var unpackCmd string
        if strings.HasSuffix(u.TempArtifact, ".tar.gz") {
            unpackCmd = fmt.Sprintf("tar xvf %s -C %s > /dev/null", u.TempArtifact, u.ConfigVM.AppPath)
        } else if strings.HasSuffix(u.TempArtifact, ".zip") {
            unpackCmd = fmt.Sprintf("unzip %s -d %s > /dev/null", u.TempArtifact, u.ConfigVM.AppPath)
        } else {
            msg := "部署包格式错误，暂仅支持 .tar.gz 和 .zip 格式的归档包"
            u.Logs = append(u.Logs, msg)
            return errors.New(msg)
        }
        u.Logs = append(u.Logs, host + ": 解压命令：" + unpackCmd)
        err, log = cae.ExecCmd(unpackCmd, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, host + ": 解压命令执行失败：" + cae.TruncCaeOut(log, 500))
            return err
        }
    }
    // 更新后应用文件检查
    err := u.CheckServerDir(host, "file")
    if err != nil {
        return err
    }
    return nil
}

// 起应用
func (u *UnitVMUpgrade) StartApp(host string) error {
    if u.ConfigVM.CMDStartup == "" {
        errMsg := "应用启动命令不能为空"
        u.Logs = append(u.Logs, errMsg)
        return errors.New(errMsg)
    }
    u.Logs = append(u.Logs, host + ": 启动应用命令：" + u.ConfigVM.CMDStartup)
    err, log := cae.ExecCmd(u.ConfigVM.CMDStartup, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
    u.Logs = append(u.Logs, host + ": 命令执行日志：" + cae.TruncCaeOut(log, 500))
    if err != nil {
        beego.Error(log)
        u.Logs = append(u.Logs, host + ": 启动应用命令执行失败")
        return err
    }
    return nil
}

// 执行后置命令
func (u *UnitVMUpgrade) ExecRearCmd(host string) error {
    u.Logs = append(u.Logs, "后置命令：" + u.ConfigVM.CMDRear)
    err, log := cae.ExecCmd(u.ConfigVM.CMDRear, u.ConfigVM.AppPath, u.ConfigVM.AppUser, host)
    u.Logs = append(u.Logs, host + ": 命令执行日志：" + cae.TruncCaeOut(log, 500))
    if err != nil {
        beego.Error(log)
        u.Logs = append(u.Logs, host + ": 后置命令执行失败")
        return err
    }
    return nil
}

// 清理工作现场
func (u *UnitVMUpgrade) CleanWS() {
    u.Logs = append(u.Logs, "清理临时工作目录")
    if !host.IsWSValid(u.ConfigVM.AppTempPath) {
        u.Logs = append(u.Logs, "临时工作目录不合法，只允许 /app/ 和 /tmp/ 下的操作。")
        return
    }
    cleanCmd := fmt.Sprintf("rm -rf %s", u.ConfigVM.AppTempPath)
    u.Logs = append(u.Logs, "本地临时目录删除命令：" + cleanCmd)
    if _, err := common.RunShellCMD(cleanCmd); err != nil {
        u.Logs = append(u.Logs, "本地工作目录删除失败，请手动处理")
        u.Logs = append(u.Logs, err.Error())
    }
    u.Logs = append(u.Logs, "本地临时目录删除成功")

    hosts := strings.Split(u.ConfigVM.Hosts, ";")
    for _, h := range hosts {
        u.Logs = append(u.Logs, h +": 临时目录删除命令:" + cleanCmd)
        err, _ := cae.ExecCmd(cleanCmd, "/tmp", "root", h)
        if err != nil {
            beego.Error(err)
            u.Logs = append(u.Logs, h + ": 删除命令执行失败", err.Error())
        }
        u.Logs = append(u.Logs, h + ": 临时目录删除成功")
    }
}

// 下载链接检查
func checkURL(link string) error {
    if !strings.HasPrefix(link, "http") && !strings.HasPrefix(link, "https") {
        return errors.New("当前仅支持http|https站点")
    }
    if err := common.URLCheck(link); err != nil {
        return errors.New("站点URL不正确")
    }
    var validHost = []string{"pan.cmrh.com", "jenkins-di1.sit.cmrh.com", "100.69.218.115"}
    u, err := url.Parse(link)
    if err != nil {
        return err
    }
    hostname := u.Hostname()
    if !common.InList(hostname, validHost) {
        return errors.New("暂不支持该下载站点")
    }
    return  nil
}