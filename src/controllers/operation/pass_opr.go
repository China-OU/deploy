package operation

import (
    "bytes"
    "controllers"
    "crypto/des"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/jinzhu/gorm"
    "initial"
    "library/common"
    "models"
    "net"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type ManageDBInfoController struct {
    controllers.BaseController
}

func (c *ManageDBInfoController) URLMapping()  {
    c.Mapping("GetDBAccounts", c.GetDBAccounts)
    c.Mapping("NewDBAccount", c.NewDBAccount)
    c.Mapping("DeleteDBPassword", c.DeleteDBPassword)
}

var ask = "V2ludGVySXNDb21l"

// GetDBAccounts 方法
// @Title Get DB account list
// @Description 获取DB账户列表
// @Param keyWord query string false "搜索关键字，支持按主机、库名、用户名搜索"
// @Param corp query string false "按租户筛选"
// @Param dialect query string false "按数据库类型筛选"
// @Param expired query int false "按是否过期筛选"
// @Success 200 true or false
// @Failure 403
// @router /db/account [get]
func (c *ManageDBInfoController) GetDBAccounts() {
    if beego.AppConfig.String("runmode") != "prd" {
        c.SetJson(0, "", "非生产环境不支持该功能！")
        return
    }
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有该页面的操作权限！")
        return
    }
    var accounts []models.DBAccount
    keyWord := c.GetString("keyWord")
    queryDialect := c.GetString("dialect")
    queryCorp := c.GetString("corp")
    page, rows := c.GetPageRows()
    queryStr := "`deleted` = 0 "

    if strings.TrimSpace(keyWord) != "" {
        queryStr += fmt.Sprintf("AND concat(`host`, `schema`, `username`) like '%%%s%%' ", keyWord)
    }
    if strings.TrimSpace(queryCorp) != "" {
        queryStr += fmt.Sprintf("AND `corp` = '%s' ", queryCorp)
    }
    if strings.TrimSpace(queryDialect) != "" {
        queryStr += fmt.Sprintf("AND `dialect` = '%s' ", queryDialect)
    }
    queryExpired := c.Ctx.Input.Query("expired")
    if queryExpired != "" {
        i, err := strconv.Atoi(queryExpired)
        if err != nil {
            c.SetJson(0, "", "expired 参数解析错误！")
            return
        }
        queryStr += fmt.Sprintf("AND `expired` = %d ", i)
    }

    var count int
    if err := initial.DB.Table("conf_db_account").Where(queryStr).Count(&count).Order(" expired, create_time desc").
        Offset((page - 1)*rows).Limit(rows).Find(&accounts).Error; err != nil {
            beego.Error(err)
            c.SetJson(0, "", "数据库查询出错！")
        return
    }

    for k, v := range accounts {
        // 先使用AES解密
        v.EncryptedPWD = common.AesDecrypt(v.EncryptedPWD)
        v.Key = common.AesDecrypt(v.Key)
        // 再使用WEB AES加密
        accounts[k].EncryptedPWD = common.WebPwdEncrypt(v.EncryptedPWD)
        accounts[k].Key = common.WebPwdEncrypt(v.Key)
    }

    resp := map[string]interface{}{
        "count": count,
        "data": accounts,
    }
    c.SetJson(1, resp, "ok")
    return
}

type DBInfo struct {
    Username        string  `json:"username"`
    EncryptedPwd    string  `json:"encrypted_pwd"`
    Key             string  `json:"key"`
}

type DBData struct {
    Corp        string      `json:"corp"`
    Host        string      `json:"host"`
    Port        uint        `json:"port"`
    Schema      string      `json:"schema"`
    Dialect     string      `json:"dialect"`
    Accounts    []DBInfo    `json:"accounts"`
}

// NewDBAccount 方法
// @Title Insert or Update DB account
// @Description 新增&更新DB账户信息
// @Param body body DBData true "body传入DB账户信息"
// @Success 200 models.DBAccount
// @Failure 403
// @router /db/account [post]
func (c *ManageDBInfoController) NewDBAccount() {
    if beego.AppConfig.String("runmode") != "prd" {
        c.SetJson(0, "", "非生产环境不支持该功能！")
        return
    }
    if strings.Contains(c.Role, "admin") == false {
        c.SetJson(0, "", "您没有操作权限！")
        return
    }
    var data DBData
    if err := json.Unmarshal(c.Ctx.Input.RequestBody, &data); err != nil {
        beego.Error(err)
        c.SetJson(0, "", "数据解析失败！")
        return
    }

    // 逐条处理
    var account models.DBAccount
    var accounts []models.DBAccount
    account.Corp = data.Corp
    account.Host = data.Host
    account.Port = data.Port
    account.Schema = data.Schema
    account.Dialect = data.Dialect
    account.ExpireTime = time.Now().Add(90 * 24 * time.Hour)
    account.Expired = 0
    account.CreateTime = time.Now().Format("2006-01-02 15:04:05")

    // 表单中重复用户检查
    user := make(map[string]string)
    for _, u := range data.Accounts {
        if _, ok := user[u.Username]; ok {
            c.SetJson(0, "", "账户列表中存在重复用户！ " + u.Username)
            return
        }
        user[u.Username] = u.EncryptedPwd
    }

    salt, _ := base64.StdEncoding.DecodeString(ask)
    for _, u := range data.Accounts {
        account.Username = u.Username
        account.EncryptedPWD = u.EncryptedPwd
        account.Key = u.Key

        // 数据校验
        if err := dbInfoCheck(&account); err != nil {
            beego.Error(err)
            c.SetJson(0, "", err.Error())
            return
        }
        // 数据库中重复记录检查
        queryStr := fmt.Sprintf("`corp` = '%s' " +
            "AND `host` = '%s' " +
            "AND `port` = '%d' " +
            "AND `schema` = '%s' " +
            "AND `username` = '%s' " +
            "AND `expired` = 0 " +
            "AND `deleted` = 0",
            account.Corp, account.Host, account.Port, account.Schema, account.Username)
        var repeat int
        if err := initial.DB.Model(&account).Where(queryStr).Count(&repeat).Error; err != nil {
            beego.Error(err)
            c.SetJson(0, "", "查询时出错！ " + err.Error())
            return
        }
        if repeat != 0 {
            msg := fmt.Sprintf("用户 %s 相同记录已存在！", account.Username)
            c.SetJson(0, "", msg)
            return
        }
        // 前端密文解密
        pwdStr := common.WebPwdDecrypt(account.EncryptedPWD)
        keyStr := common.WebPwdDecrypt(account.Key)
        realKey := keyStr + string(salt)
        if len(realKey) < 16 || len(realKey) > 18 {
            msg := fmt.Sprintf("用户 %s 密钥长度错误！", account.Username)
            c.SetJson(0, "", msg)
            return
        }

        // 解出明文
        plainStr, err := decrypt(pwdStr, []byte(keyParser(realKey)))
        if err != nil {
            msg := fmt.Sprintf("用户 %s 解密失败，请核对密文和密钥！", account.Username)
            c.SetJson(0, "", msg)
            return
        }

        // 对明文密码 & key 加密
        account.EncryptedPWD = common.AesEncrypt(plainStr)
        account.Key = common.AesEncrypt(keyStr)

        accounts = append(accounts, account)
    }

    // 一次性入库，Gorm v1.9暂不支持批量创建，数据校验通过后逐条处理
    tx := initial.DB.Begin()
    for _, v := range accounts {
        if err := tx.Create(&v).Error; err != nil {
            tx.Rollback()
            beego.Error()
            c.SetJson(0, "", "数据插入失败！ " + err.Error())
            return
        }
    }
    tx.Commit()
    c.SetJson(1, "", "录入成功！")
    return
}

// UpdateDBPassword 方法
// @Title Update DB encrypted password
// @Description 更新DB账户信息
// @Param body body models.DBAccount true "带ID传入待更新的数据"
// @Success 200 true or false
// @Failure 403
// @router /db/account [put]
func (c *ManageDBInfoController) UpdateDbPassword() {
    if beego.AppConfig.String("runmode") != "prd" {
        c.SetJson(0, "", "非生产环境不支持该功能！")
        return
    }
    if !strings.Contains(c.Role, "admin") {
        c.SetJson(0, "", "您没有操作权限！")
        return
    }
    var dbInfo models.DBAccount
    var exist models.DBAccount
    if err := json.Unmarshal(c.Ctx.Input.RequestBody, &dbInfo); err != nil {
        c.SetJson(0, "", "数据解析失败！")
        return
    }
    // 数据校验
    if err := dbInfoCheck(&dbInfo); err != nil {
        c.SetJson(0, "", err.Error())
        return
    }

    err := initial.DB.Model(&exist).Where("`deleted` = 0 AND `id` = ?", dbInfo.ID).First(&exist).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.SetJson(0, "", "记录不存在！")
            return
        } else {
            c.SetJson(0, "", "查询数据时出错！")
            return
        }
    }

    // 第一次 WebDecrypt 得到原始数据
    pwdStr := common.WebPwdDecrypt(dbInfo.EncryptedPWD)
    keyStr := common.WebPwdDecrypt(dbInfo.Key)
    salt, _ := base64.StdEncoding.DecodeString(ask)
    realKey := keyStr + string(salt)
    // 直接点击确定
    if common.WebPwdDecrypt(pwdStr) == common.AesDecrypt(exist.EncryptedPWD) &&
        common.WebPwdDecrypt(keyStr) == common.AesDecrypt(exist.Key) {
        c.SetJson(0, "", "数据相同，无需更新！")
        return
    }
    if len(realKey) < 16 || len(realKey) > 18 {
        c.SetJson(0, "", "密钥长度错误！")
        return
    }
    plainPwd, err := decrypt(pwdStr, []byte(keyParser(realKey)))
    if err != nil {
        c.SetJson(0, "", "解密失败，请核对密文和密钥！")
        return
    }
    // 使用相同密文串
    if common.AesEncrypt(plainPwd) == exist.EncryptedPWD &&
        common.AesEncrypt(keyStr) == exist.Key {
        c.SetJson(0, "", "数据相同，无需更新！")
        return
    }
    // 使用不同密文串
    updates := map[string]interface{}{
        "encrypted_pwd": common.AesEncrypt(plainPwd),
        "key": common.AesEncrypt(keyStr),
        "update_time": time.Now(),
    }

    tx := initial.DB.Begin()
    if err := tx.Model(&dbInfo).Where("`id` = ?", exist.ID).Updates(updates).Error; err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", "更新失败！" + err.Error())
        return
    }
    tx.Commit()
    c.SetJson(1, "", "更新成功！")
    return

}

func dbInfoCheck(data *models.DBAccount) (err error) {
    if strings.TrimSpace(data.Corp) == "" {
        return errors.New("租户信息不能为空！")
    }
    mStr := "^[A-Z]{1,}$"
    if m, _ := regexp.Match(mStr, []byte(data.Corp)); !m {
        return errors.New("租户名只能是大写英文！")
    }
    if strings.TrimSpace(data.Host) == "" {
        return errors.New("主机不能为空！")
    }
    // 主机格式校验，只能是IP或cmftdc.cn域名
    ip := net.ParseIP(data.Host)
    if ip.To4() == nil {
        mStr :="^[a-zA-Z.]+\\.cmftdc.cn"
        if m, _ := regexp.Match(mStr, []byte(data.Host)); !m {
            return errors.New("主机只能是IP或*.cmftdc.cn域名！")
        }
    }
    data.Dialect = strings.ToLower(strings.TrimSpace(data.Dialect))
    if data.Dialect == "" {
        return errors.New("数据库类型不能为空！")
    }
    mStr = "(^(mysql|oracle|postgresql|redis|mongodb|sqlserver))"
    if m, _ := regexp.Match(mStr, []byte(data.Dialect)); !m {
        return errors.New("数据库类型错误！")
    }
    if strings.TrimSpace(data.Schema) == "" {
        return errors.New("库名不能为空！")
    }
    mStr = "^[0-9a-zA-Z_]{1,}$"
    if m, _ := regexp.Match(mStr, []byte(data.Schema)); !m {
        return errors.New("库名只允许包含英文和数字！")
    }
    if data.Port < 1024 || data.Port > 65536 {
        return errors.New("端口值错误！1025~65535")
    }
    if strings.TrimSpace(data.Username) == "" {
        return errors.New("DB用户名不能为空！")
    }
    if strings.TrimSpace(data.EncryptedPWD) == "" {
        return errors.New("密码不能为空！")
    }
    if strings.TrimSpace(data.Key) == "" {
        return errors.New("密钥不能为空！")
    }
    return nil
}

// DeleteDBPassword 方法
// @Title Delete DB encrypted password
// @Description 删除DB账户信息
// @Param account_id query int true "传入待删除的记录ID"
// @Success 200 models.DBAccount
// @Failure 403
// @router /db/account/:account_id [delete]
func (c *ManageDBInfoController) DeleteDBPassword()  {
    if beego.AppConfig.String("runmode") != "prd" {
        c.SetJson(0, "", "非生产环境不支持该功能！")
        return
    }
    if strings.Contains(c.Role, "admin") == false {
        c.SetJson(0, "", "您没有操作权限！")
        return
    }
    accountID, err := c.GetInt("account_id")
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "参数获取失败！" + err.Error())
        return
    }
    if accountID == 0 {
        c.SetJson(0, "", "参数不能为0！")
        return
    }

    var count int
    err = initial.DB.Model(models.DBAccount{}).
        Where("`deleted` = 0 AND `id` = ?", accountID).Count(&count).Error
    if err != nil {
        beego.Error(err)
        c.SetJson(0, "", "查询时出错！ " + err.Error())
        return
    }
    if count == 0 {
        c.SetJson(0, "", "没有查询到该记录！")
        return
    }
    updates := map[string]interface{}{
        "key": common.AesEncrypt("********"),
        "encrypted_pwd": common.AesEncrypt("********"),
        "deleted": 1,
    }
    tx := initial.DB.Begin()
    err = tx.Table("conf_db_account").Where("`id` = ?", accountID).Updates(updates).Error
    if err != nil {
        tx.Rollback()
        beego.Error(err)
        c.SetJson(0, "", "删除失败！ " + err.Error())
        return
    }
    tx.Commit()
    c.SetJson(1, "", "删除成功！")
    return
}

func keyParser(str string) string {
    key := base64.StdEncoding.EncodeToString([]byte(str))
    if len(key) < 24 {
        key += "************************"
    }
    key = key[0:24]
    return key
}

func zeroPadding(cipherText []byte, blockSize int) []byte {
    padding := blockSize - len(cipherText)%blockSize
    padText := bytes.Repeat([]byte{0}, padding)
    return append(cipherText, padText...)
}

func zeroUnPadding(origData []byte) []byte {
    return bytes.TrimFunc(origData,
        func(r rune) bool {
            return r == rune(0)
        })
}

func encrypt(text string, key []byte) (string, error) {
    src := []byte(text)
    block, err := des.NewTripleDESCipher(key)
    if err != nil {
        return "", err
    }
    bs := block.BlockSize()
    src = zeroPadding(src, bs)
    if len(src)%bs != 0 {
        return "", errors.New("need a multiple of the blockSize")
    }
    out := make([]byte, len(src))
    dst := out
    for len(src) > 0 {
        block.Encrypt(dst, src[:bs])
        src = src[bs:]
        dst = dst[bs:]
    }
    return hex.EncodeToString(out), nil
}

func decrypt(encrypted string , key []byte) (string, error) {
    src, err := hex.DecodeString(encrypted)
    if err != nil {
        return "", err
    }
    block, err := des.NewTripleDESCipher(key)
    if err != nil {
        return "", err
    }
    out := make([]byte, len(src))
    dst := out
    bs := block.BlockSize()
    if len(src)%bs != 0 {
        return "", errors.New("crypto/cipher: input not full blocks")
    }
    for len(src) > 0 {
        block.Decrypt(dst, src[:bs])
        src = src[bs:]
        dst = dst[bs:]
    }
    out = zeroUnPadding(out)
    plain := string(out)
    if ok := common.IsUTF8([]byte(plain)); !ok {
        return "", errors.New("the cipher does not match the key")
    }
    return plain, nil
}