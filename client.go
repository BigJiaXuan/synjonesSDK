package synjonesSDK

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BigJiaXuan/synjonesSDK/internal"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Conf struct {
	URL         string //tsm 地址
	AppKey      string
	Des3Key     string
	SvcPkcs8Key string
}

// OpenAcc 虚拟卡开户结构体
type OpenAcc struct {
	Sno        string
	Name       string
	Sex        string
	IdNo       string
	Phone      string
	SchoolCode string
	DeptCode   string
	PidCode    string
	InDate     string
	ExpDate    string
	CardType   string
	Photo      string
}

type Client interface {
	// Send 通用发送tsm请求方法
	Send(ctx context.Context, token, request, method string) (code, resp string, err error)
	// GetAccessToken 获取tsm access_token
	GetAccessToken(ctx context.Context) (string, error)
	// UnFrozenCard 校园卡解冻
	UnFrozenCard(ctx context.Context, accessToken string, account int64) error
	// OpenAccount 虚拟卡开户
	OpenAccount(ctx context.Context, accessToken, sno, name, sex, idNo, phone, schoolCode, deptCode, pidCode, inDate, expDate, cardType, photo string) error
	// GetBarCode 获取一卡通二维码
	GetBarCode(ctx context.Context, accessToken, account, payType, payAcc string) (string, error)
	// OpenAccountV2 虚拟卡开户
	OpenAccountV2(ctx context.Context, accessToken string, openAcc OpenAcc) (sno, account string, err error)
}
type client struct {
	conf    *Conf
	encrypt internal.EncryptUtil
}

func NewClient(conf *Conf) Client {
	encrypt := internal.NewEncryptUtilImpl(conf.Des3Key)
	return &client{conf: conf, encrypt: encrypt}
}

// GetBarCode
//
//	@Description: 获取一卡通二维码
//	@receiver r
//	@param ctx
//	@param accessToken
//	@param account 一卡通账号
//	@param payType 支付方式 1 校园卡账户 2 绑定银行卡 3 自定义银行卡
//	@param payAcc paytype为1 值为 ### 卡账户 其他为电子账户类型 paytype为2 可以空 paytype为3 可以是银行卡号
//	@return string
//	@return error
func (r *client) GetBarCode(ctx context.Context, accessToken, account, payType, payAcc string) (string, error) {
	method := "synjones.onecard.barcode.get"
	request := fmt.Sprintf("{\n\"barcode_get\": {\n\"account\": \"%s\","+
		" \"paytype\": \"%s\", "+
		"\"payacc\": \"%s\"\n} }", account, payType, payAcc)
	_, resp, err := r.Send(ctx, accessToken, request, method)
	if err != nil {
		return "", err
	}
	type Response struct {
		BarcodeGet struct {
			Retcode string `json:"retcode"`
			Errmsg  string `json:"errmsg"`
			Account string `json:"account"`
			Paytype string `json:"paytype"`
			Payacc  string `json:"payacc"`
			Barcode string `json:"barcode"`
			Expires string `json:"expires"`
		} `json:"barcode_get"`
	}
	var response Response
	err = json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return "", err
	}
	if response.BarcodeGet.Retcode != "0" {
		return "", errors.New(response.BarcodeGet.Errmsg)
	}
	return response.BarcodeGet.Barcode, nil
}

// OpenAccount
//
//	@Description:虚拟卡开户
//	@receiver r
//	@param ctx
//	@param accessToken
//	@param sno 学工号
//	@param name 姓名
//	@param sex 性别 1男 2女 9 未知
//	@param idNo 身份证号
//	@param phone 手机号
//	@param schoolCode 校区代码
//	@param deptCode 部门代码
//	@param cardType 卡类型 800 正式卡 801 临时卡
//	@param pidCode 身份代码
//	@param inDate 入校日期
//	@param expDate 失效日期
//	@param photo base64照片
//	@return error
func (r *client) OpenAccount(ctx context.Context, accessToken, sno, name, sex, idNo, phone, schoolCode, deptCode, pidCode, inDate, expDate, cardType, photo string) error {
	method := "synjones.onecard.open.acc"
	request := fmt.Sprintf("{\n\"open_acc\": {\n\"sno\": \"%s\","+
		"\n\"name\": \"%s\","+
		"\n\"sex\": \"%s\","+
		"\n\"idno\": \"%s\","+
		"\n\"phone\": \"%s\","+
		"\n\"schoolcode\": \"%s\","+
		"\n\"depcode\": \"%s\","+
		"\n\"born\": \" \","+
		"\n\"pidcode\": \"%s\","+
		"\n \"email\": \"\","+
		"\n\"indate\": \"%s\","+
		"\n\"expdate\": \"%s\","+
		"\n\"cardtype\": \"%s\","+
		"\n\"photo_image\": \"\"\n}\n}",
		sno, name, sex, idNo, phone, schoolCode, deptCode, pidCode, inDate, expDate, cardType)
	_, resp, err := r.Send(ctx, accessToken, request, method)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	type Response struct {
		OpenAcc struct {
			Retcode string `json:"retcode"`
			Errmsg  string `json:"errmsg"`
			Sno     string `json:"sno"`
			Account string `json:"account"`
		} `json:"open_acc"`
	}
	response := Response{}
	err = json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return err
	}
	if response.OpenAcc.Retcode != "0" {
		return errors.New(response.OpenAcc.Errmsg)
	}
	return nil
}

func (r *client) OpenAccountV2(ctx context.Context, accessToken string, openAcc OpenAcc) (sno, account string, err error) {
	method := "synjones.onecard.open.acc"
	request := fmt.Sprintf("{\n\"open_acc\": {\n\"sno\": \"%s\","+
		"\n\"name\": \"%s\","+
		"\n\"sex\": \"%s\","+
		"\n\"idno\": \"%s\","+
		"\n\"phone\": \"%s\","+
		"\n\"schoolcode\": \"%s\","+
		"\n\"depcode\": \"%s\","+
		"\n\"born\": \" \","+
		"\n\"pidcode\": \"%s\","+
		"\n \"email\": \"\","+
		"\n\"indate\": \"%s\","+
		"\n\"expdate\": \"%s\","+
		"\n\"cardtype\": \"%s\","+
		"\n\"photo_image\": \"\"\n}\n}",
		openAcc.Sno, openAcc.Name, openAcc.Sex, openAcc.IdNo, openAcc.Phone, openAcc.SchoolCode, openAcc.DeptCode, openAcc.PidCode, openAcc.InDate, openAcc.ExpDate, openAcc.CardType)
	_, resp, err := r.Send(ctx, accessToken, request, method)
	if err != nil {
		return "", "", err
	}
	fmt.Println(resp)
	type Response struct {
		OpenAcc struct {
			Retcode string `json:"retcode"`
			Errmsg  string `json:"errmsg"`
			Sno     string `json:"sno"`
			Account string `json:"account"`
		} `json:"open_acc"`
	}
	response := Response{}
	err = json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return "", "", err
	}
	if response.OpenAcc.Retcode != "0" {
		return "", "", errors.New(response.OpenAcc.Errmsg)
	}
	return response.OpenAcc.Sno, response.OpenAcc.Account, nil
}

// UnFrozenCard 校园卡解冻
func (r *client) UnFrozenCard(ctx context.Context, accessToken string, account int64) error {
	method := "synjones.onecard.unfrozen.card"
	request := fmt.Sprintf("{\n\"unfrozen_card\": {\n\"account\": \"%d\" }\n}", account)
	_, resp, err := r.Send(ctx, accessToken, request, method)
	if err != nil {
		return err
	}
	type Response struct {
		UnforzenCard struct {
			Account string `json:"account"`
			Retcode string `json:"retcode"`
			Errmsg  string `json:"errmsg"`
		} `json:"unforzen_card"`
	}
	var res Response
	err = json.Unmarshal([]byte(resp), &res)
	if err != nil {
		return err
	}
	if res.UnforzenCard.Retcode != "0" {
		return errors.New(res.UnforzenCard.Errmsg)
	}
	return nil
}

// GetAccessToken 获取access_token
func (r *client) GetAccessToken(ctx context.Context) (string, error) {
	accessToken := "0000000000000000000000000000000000000000000000000000000000000000" +
		"0000000000000000000000000000000000000000000000000000000000000000"
	request := "{\"authorize_access_token\": {}}"
	method := "synjones.authorize.access_token"
	_, resp, err := r.Send(ctx, accessToken, request, method)
	if err != nil {
		return "", err
	}
	type Response struct {
		AuthorizeAccessToken struct {
			Retcode     string `json:"retcode"`
			Errmsg      string `json:"errmsg"`
			AccessToken string `json:"access_token"`
			ExpiresIn   string `json:"expires_in"`
		} `json:"authorize_access_token"`
	}
	var res Response
	err = json.Unmarshal([]byte(resp), &res)
	if err != nil {
		return "", err
	}
	// 当retcode非0时，返回错误
	if res.AuthorizeAccessToken.Retcode != "0" {
		return "", errors.New(res.AuthorizeAccessToken.Errmsg)
	}
	return res.AuthorizeAccessToken.AccessToken, nil
}

func (r *client) Send(ctx context.Context, token, request, method string) (code, resp string, err error) {
	// 1.对request参数签名
	requested := r.encrypt.SignRequest(request)

	// 2.对sign排序
	sign := r.sortSign(token, requested, method)
	// 3.给sign签名
	signed := r.encrypt.SignatureSign(sign, r.conf.SvcPkcs8Key)
	// 准备http请求体
	req := map[string]string{
		"method":       method,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"format":       "json",
		"app_key":      r.conf.AppKey,
		"access_token": token,
		"v":            "2.0",
		"sign_method":  "rsa",
		"request":      requested,
		"sign":         signed,
	}
	formData := url.Values{}
	for key, value := range req {
		formData.Set(key, value)
	}
	// 发送http请求
	response, err := http.Post(r.conf.URL, "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", "", err
	}

	// 关闭响应体
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)
	// 读取返回
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}
	// 准备对body进行解密
	code, resp = r.decodeResponse(string(body))
	// 判断当code非0时，返回错误信息
	if code != "0" {
		intCode, _ := strconv.Atoi(code)
		resp = ErrMsg(int64(intCode))
	}
	return code, resp, nil
}

// sortSign
//
//	@Description: 对sign参数按从小到大排序，
//	@receiver r
//	@param token
//	@param method
//	@param request
//	@return string access_tokenxxxapp_keyxxxformatxxxmethodxxxrequestxxxtimestampxxxvxxx
func (r *client) sortSign(token, request, method string) string {
	// 声明一个sign结构体
	sign := map[string]string{
		"method":       method,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"format":       "json",
		"app_key":      r.conf.AppKey,
		"access_token": token,
		"v":            "2.0",
		"sign_method":  "rsa",
		"request":      request,
	}

	var keys []string
	for key := range sign {
		keys = append(keys, key)
	}
	// 排序key
	sort.Strings(keys)
	sortedString := ""
	for _, key := range keys {
		sortedString += fmt.Sprintf("%s%s", key, sign[key])
	}
	return sortedString
}

// DecodeResponse
//
//	@Description: 对tsm返回的请求解码
//	@receiver r
//	@param response
//	@return string
func (r *client) decodeResponse(response string) (errcode, request string) {
	// errcode=0&request=pXMP7aN%2BJapLT9iRIeBJ5%2FNpuRpXyZ3KSjx0Zge5R507u4ufhtiERiN0UygGUqhuxPuFSfHMkHEH5peNDx%2FN7na0%2FBJcTwBla0awncdyrMUStKyIgd8cShugF1HATPzyCnYIb4MmwrEarWBGP9tj8HuLnMQtNHDQlpkPLkcvRr3MctHDEJRhT4rp7ewdS%2BZw1HJVd2%2FdyZGeqc%2BgUUgB%2FZojwaKu3z7w9Ib9DOgdPWoVCC%2BzVlwJfVx2GLtr74j%2B5F4o6GUcIa4C5rZjywqbWjJNFtwMWY0cjjC8ANMcDCD9y7ajFaBQCBHlqg9iGW4k&sign=XgQz5fjo7xykIsHy6dZ8ajcCfGOkBPdli0kby%2FuxGB8haVarOLFKe9scFBMbL1vZm25u4vIi1Ju3%2FkO2oP1v0Rol54S%2BS%2FnHTcYuS6fdBQfA%2BI9lrX3kln0aypTxSUkyNyIESVfljcesu4TnVQvfKgbd0e8N2jnaQEfO0Ngk%2Fd%2FMvdR6dnrQp2hbf%2BPuG%2FgKvIN8ZCalKSAmGKi4eDt%2BQ5vRRFOlGTAF4yysDPhB8qPwVf8TnNZ73vgK%2FoFOt6oVkzWHziGCsUZoPYkVIan3O47eSeoYW3J5kEoiRYuHGVeENUWmBWIUsFSHGcOgZRajjKLjtrToM49ADaJGmGZ%2BKQ%3D%3D
	// 对response进行url解码
	decodeResponse, err := url.PathUnescape(response)
	if err != nil {
		return "", ""
	}
	// 解析参数
	query, err := url.ParseQuery(decodeResponse)
	if err != nil {
		return "", ""
	}
	errorCode := query.Get("errcode")
	req := strings.ReplaceAll(query.Get("request"), " ", "+")
	// 对request进行解码操作
	req = r.encrypt.DecryptResponse(req)
	return errorCode, req
}
