package common

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

//判断一个数据是否为空，支持int, float, string, slice, array, map的判断
func Empty(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		if reflect.ValueOf(value).Len() == 0 {
			return true
		} else {
			return false
		}
	}
	return false
}

//判断某一个值是否在列表(支持 slice, array, map)中
func InList(needle interface{}, haystack interface{}) bool {
	//interface{}和interface{}可以进行比较，但是interface{}不可进行遍历
	hayValue := reflect.ValueOf(haystack)
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice, reflect.Array:
		//slice, array类型
		for i := 0; i < hayValue.Len(); i++ {
			if hayValue.Index(i).Interface() == needle {
				return true
			}
		}
	case reflect.Map:
		//map类型
		var keys []reflect.Value = hayValue.MapKeys()
		for i := 0; i < len(keys); i++ {
			if hayValue.MapIndex(keys[i]).Interface() == needle {
				return true
			}
		}
	default:
		return false
	}
	return false
}

//返回某一个值是否在列表位置(支持 slice, array, map) -1为不再列表中
func InListIndex(needle interface{}, haystack interface{}) int {
	//interface{}和interface{}可以进行比较，但是interface{}不可进行遍历
	hayValue := reflect.ValueOf(haystack)
	switch reflect.TypeOf(haystack).Kind() {
	case reflect.Slice, reflect.Array:
		//slice, array类型
		for i := 0; i < hayValue.Len(); i++ {
			if hayValue.Index(i).Interface() == needle {
				return i
			}
		}
	case reflect.Map:
		//map类型
		var keys []reflect.Value = hayValue.MapKeys()
		for i := 0; i < len(keys); i++ {
			if hayValue.MapIndex(keys[i]).Interface() == needle {
				return i
			}
		}
	default:
		return -1
	}
	return -1
}

//string转int
func StrToInt(str string) int {
	intval, _ := strconv.Atoi(str)
	return intval
}

//浮点数四舍五入，并取前几位
func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

//通过interface{}获取字符串
func GetString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

//通过interface{}获取数值型数据
//此获取比较灵活，转换规则如下
//1、如果接收数据为浮点string，则返回浮点数的整数部分，如果是整型string，则返回整数，如果是纯字符串，则返回0
//2、如果接收数据是float型，则返回float的整数部分
//3、如果接收数据是int, int32, int64型，则返回int
func GetInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			fval, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0
			}
			return int(fval)
		}
		return int(n)
	case float32:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

//通过interface{}获取小数型数据
//此获取比较灵活，转换规则如下
//1、如果接收数据为浮点string，则将字符串转换为浮点数
//2、如果接收数据是float型，则返回float数据
//3、如果接收数据是int, int32, int64型，则转义成float类型
//4、返回的数据结果统一为float64
func GetFloat(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	case string:
		result, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return result
	}
	return 0
}

/**
 * 根据path读取文件中的内容，返回字符串
 * 建议使用绝对路径，例如："./schema/search/appoint.json"
 */
func ReadFile(path string) string {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd)
}

func ReadJson(path string) Info {
	jsonStr := ReadFile(path)
	ret := Info{}
	err := json.Unmarshal([]byte(jsonStr), &ret)
	if err != nil {
		panic("文件[" + path + "]的内容不是json格式")
	}
	return ret
}

/**
  判断内网IP
  A  10.0.0.0/8：10.0.0.0～10.255.255.255
  B  172.16.0.0/12：172.16.0.0～172.31.255.255
  C  192.168.0.0/16：192.168.0.0～192.168.255.255
**/
func CheckInternalIp(ip string) bool {
	if ip == "127.0.0.1" {
		return true
	}
	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		return false
	}
	a_from_ip := net.ParseIP("10.0.0.0")
	a_to_ip := net.ParseIP("10.255.255.255")
	b_from_ip := net.ParseIP("172.16.0.0")
	b_to_ip := net.ParseIP("172.31.255.255")
	c_from_ip := net.ParseIP("192.168.0.0")
	c_to_ip := net.ParseIP("192.168.255.255")
	if bytes.Compare(trial, a_from_ip) >= 0 && bytes.Compare(trial, a_to_ip) <= 0 {
		return true
	}
	if bytes.Compare(trial, b_from_ip) >= 0 && bytes.Compare(trial, b_to_ip) <= 0 {
		return true
	}
	if bytes.Compare(trial, c_from_ip) >= 0 && bytes.Compare(trial, c_to_ip) <= 0 {
		return true
	}
	return false
}

func CheckIp(ip string) bool {
	addr := strings.Trim(ip, " ")
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

func Md5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

//获取当前日期为当月第几周
func CountWeek(TimeFormat string) int {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", TimeFormat, loc)
	month := t.Month()
	year := t.Year()
	days := 0
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30

		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	week := 1
	for i := 1; i <= days; i++ {
		dayString := GetString(i)
		if i < 10 {
			dayString = "0" + dayString
		}
		dateString := strings.Split(TimeFormat, "-")[0] + "-" + strings.Split(TimeFormat, "-")[1] + "-" + dayString + " 18:30:50"
		t1, _ := time.ParseInLocation("2006-01-02 15:04:05", dateString, loc)
		if t.YearDay() > t1.YearDay() {
			if t1.Weekday().String() == "Sunday" {
				week++
			}
		}

	}

	return week
}
func GetWeekday(TimeFormat string) string {
	loc, _ := time.LoadLocation("Local")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", TimeFormat, loc)
	if t.Weekday().String() == "Monday" {
		return "星期一"
	}
	if t.Weekday().String() == "Tuesday" {
		return "星期二"
	}
	if t.Weekday().String() == "Wednesday" {
		return "星期三"
	}
	if t.Weekday().String() == "Thursday" {
		return "星期四"
	}
	if t.Weekday().String() == "Friday" {
		return "星期五"
	}
	if t.Weekday().String() == "Saturday" {
		return "星期六"
	}
	if t.Weekday().String() == "Sunday" {
		return "星期日"
	}
	return ""

}
func SubString(str string, begin, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

func CheckNil(note string) string {
	if note == "<nil>" || note == "null" || note == "nil" {
		return ""
	}
	return note
}

func ReadyToRelease(release_time, date_sep string) bool {
	now_str := time.Now().Format("15:04")
	// 4点是每天的分隔点
	if now_str < date_sep {
		return true
	}
	if now_str >= release_time {
		return true
	} else {
		return false
	}
}

func Post(url string, headers map[string]string, data []byte) (res []byte, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	if !strings.HasPrefix(resp.Status, "20") {
		res, err = ioutil.ReadAll(resp.Body)
		fmt.Println(string(res))
		return res, errors.New("请求失败")
	}
	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}
	return res, nil
}

func RegexpMatched(reStr, str string) (matched bool, err error) {
	matched,err = regexp.Match(reStr, []byte(str))
	return
}

// 生成固定长度随机字符串
func GenRandString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// 校验一个字符串是否为合法的IPv4
func IsValidIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip.To4() == nil {
		return false
	}
	return true
}

// 获取查询周期内的每一天
func GetDay(sStr, eStr string) (dayList []time.Time, err error) {
	var (
		s time.Time
		e time.Time
	)

	if s, err = time.ParseInLocation("2006-01-02 15:04:05", sStr, time.Local) ; err != nil {
		return nil,err
	}

	if e, err = time.ParseInLocation("2006-01-02 15:04:05", eStr, time.Local) ; err != nil {
		return nil,err
	}

	dayList = append(dayList,s)

	for {
		s = s.Add(24 * time.Hour)
		if sub := e.Sub(s).Seconds() ; sub <= 0 {
			break
		}else {
			dayList = append(dayList,s)
		}
	}

	return dayList, nil
}

// 检查开始/结束日期是否合法,并格式化为"2006-01-02 15:04:05"格式
func FormatTime(qStr, sStr, eStr string) (start, end string, err error) {
	var (
		sTime time.Time
		eTime time.Time
		nowD int
		nowM int
		nowY int
	)

	now := time.Now()
	nowD = now.Day()
	nowM = int(now.Month())
	nowY = now.Year()

	if sStr == "" && eStr == ""  && qStr == "" {
		sTime = time.Date(nowY, time.Month(nowM - 1), nowD, 4, 0, 0, 0, time.Local)
		eTime = time.Date(nowY, time.Month(nowM), nowD + 1 , 4, 0, 0, 0, time.Local)
		return sTime.Format("2006-01-02 15:04:05"), eTime.Format("2006-01-02 15:04:05"), nil
	}

	if qStr == "" {
		if sStr == "" || eStr == "" {
			return "", "", errors.New("开始或结束时间未指定！")
		}
		if sTime, err = time.ParseInLocation("2006-01-02",sStr,time.Local) ; err != nil {
			return "", "", err
		}
		if eTime, err = time.ParseInLocation("2006-01-02",eStr,time.Local) ; err != nil {
			return "", "", err
		}
		if sTime.Sub(eTime).Seconds() > 0 {
			return "", "", errors.New("开始时间大于结束时间，请重新选择！")
		}
		if eTime.Add(24 * time.Hour).Sub(sTime).Hours() / 24 > 366 {
			return "", "", errors.New("查询周期内总时间不允许超过366天，请重新选择！")
		}
		sTime = sTime.Add(4 * time.Hour)
		eTime = eTime.Add(28 * time.Hour)
	}else {
		if sStr != "" || eStr != "" {
			return "", "", errors.New("使用快速查询时，开始或结束请置空！")
		}
		eTime = time.Date(nowY, time.Month(nowM), nowD + 1 , 4, 0, 0, 0, time.Local)
		switch qStr {
		case "1":
			sTime = time.Date(nowY, time.Month(nowM - 1), nowD, 4, 0, 0, 0, time.Local)
		case "3":
			sTime = time.Date(nowY, time.Month(nowM - 3), nowD, 4, 0, 0, 0, time.Local)
		case "6":
			sTime = time.Date(nowY, time.Month(nowM - 6), nowD, 4, 0, 0, 0, time.Local)
		default:
			return "", "", errors.New("快速查询只能选择最近一个月/三个月/六个月！")
		}
	}
	sStr = sTime.Format("2006-01-02 15:04:05")
	eStr = eTime.Format("2006-01-02 15:04:05")

	return sStr, eStr, nil
}

// 执行shell命令
func RunShellCMD(bashCMD string) (out string, err error) {
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	cmd := exec.Command("sh", "-c", bashCMD)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	beego.Info("本地执行的命令为：",cmd)
	if err =  cmd.Run() ; err != nil {
		fmt.Println(stdErr.String())
		return "", errors.New(strings.Trim(stdErr.String(),"\n"))
	}
	return strings.Trim(stdOut.String(),"\n"), nil
}

// vm 上线日期和时间校验是否合法;日期2006-01-02格式，日期为08:08格式
func CheckOnlineDate(onlineDate, onlineTime, now string) (string, string, error) {
	now_day := strings.Replace(now[0:10], "-", "", -1)
	now_time := now[11:16]
	if now_time < "04:00" {
		now_day = time.Now().AddDate(0, 0, -1).Format("20060102")
	}
	if onlineDate != "" {
		if _, err := time.Parse("20060102", onlineDate) ; err != nil {
			return  "", "", errors.New("上线日期格式不正确请检查！")
		}
		now_day = onlineDate
	}
	if onlineTime != "" {
		if _, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("2020-02-21 %s:18", onlineTime)) ; err != nil {
			return  "", "", errors.New("上线时间格式不正确请检查！")
		}
		now_time = onlineTime
	}
	return now_day, now_time, nil
}

// 去除string中的重复字符
func RemoveRepeatedElement(str string) (strNew string) {
	strArr := make([]string, 0, len(str))
	for i := 0; i < len(str); i++ {
		strArr = append(strArr, str[i:i+1])
	}
	for i := 0; i < len(strArr); i++ {
		if strArr[i] == "" {
			continue
		}
		if i == len(strArr)-1 {
			break
		}
		for k, v := range strArr[i+1:] {
			if strArr[i] == v {
				strArr[i+k+1] = ""
			}
		}
	}
	new := ""
	for _, v := range strArr {
		if v == " " || v == "" {
			continue
		}
		new += fmt.Sprint(v)
	}
	return new
}

func TextPrefixString(content string) string {
	runes := []rune(content)
	// 小于65535即可，取65000
	if len(runes) > 60000 {
		return string(runes[0:60000])
	}
	return content
}

func TextSuffixString(content string) string {
	runes := []rune(content)
	// 小于65535即可，取65000
	if len(runes) > 60000 {
		return string(runes[len(runes)-60000:])
	}
	return content
}