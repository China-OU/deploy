package datalist

import "models"

type LoginInputData struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type LoginData struct {
	Code string `json:"code"`
	Message string `json:"message"`
	Data LoginDetail `json:"data"`
}

type LoginDetail struct {
	UserId string `json:"userId"`
	Token string `json:"token"`
	UmDTO UserDetail `json:"umDTO"`

}

type UserDetail struct {
	UserId string `json:"userId"`
	UnId string `json:"unId"`
	UserName string `json:"userName"`
	Users []map[string]string `json:"users"`
}

type UserInfo struct {
	models.UserLogin
	Token      string    `json:"token"`
	Role       string    `json:"role"`
}

// 返回参数
type NucBaseRet struct {
	State      bool         `json:"state"`
	Refresh    bool         `json:"refresh"`
	Topologys  interface{} `json:"topologys"`
	Message    string       `json:"message"`
	Data       interface{} `json:"data"`
}

type CheckRet struct {
	BusinessUnitName    string    `json:"businessUnitName"`
	BusinessUnit        string    `json:"businessUnit"`
	Token               string    `json:"token"`
	AccountId           string    `json:"accountId"`
	Uid                 string    `json:"uid"`
	Name                string    `json:"name"`
	Nickname            string    `json:"nickname"`
}