package login

import "models"

type LoginInputData struct {
	UserName     string `json:"username"`
	Password     string `json:"password"`
	VerifyCode   string `json:"verify_code"`
	ImageUUID    string `json:"image_uuid"`
}

// 输入参数
type NucLoginData struct {
	AccountId   string   `json:"accountId"`
	Password    string   `json:"password"`
	VerifyCode  string   `json:"verifyCode"`
	Options     LoginOption   `json:"options"`
}

type LoginOption struct {
	EncodePasswd  string   `json:"encodePasswd"`
	ModuleCode    string   `json:"moduleCode"`
	ImageUUID     string   `json:"imageUUID"`
	AuthType      string   `json:"authType"`
}

type RefreshOption struct {
	Username        string  `json:"username"`
	MobileOrEmail   string  `json:"mobileOrEmail"`
	MessageId       string  `json:"messageId"`
}

type SMSValidateOption struct {
	MessageCode   string   `json:"messageCode"`
	MessageId     string   `json:"messageId"`
}

// 返回参数
type NucBaseRet struct {
	State      bool         `json:"state"`
	Refresh    bool         `json:"refresh"`
	Topologys  interface{} `json:"topologys"`
	Message    string       `json:"message"`
	Data       interface{} `json:"data"`
}

type LoginRet struct {
	Birthday      string    `json:"birthday"`
	AccessToken   string    `json:"accessToken"`
	Type          string    `json:"type"`
	AccountId     string    `json:"accountId"`
	Uid           string    `json:"uid"`
	CmToken       string    `json:"cmToken"`
	Name          string    `json:"name"`
	Nickname      string    `json:"nickname"`
	Id            string	`json:"id"`
	UserType      string    `json:"userType"`
	RefreshToken  string    `json:"refreshToken"`
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

type UserInfo struct {
	models.UserLogin
	Token      string    `json:"token"`
	Role       string    `json:"role"`
}
