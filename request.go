package synjonesSDK

import (
	"bytes"
	"context"
	"fmt"
	"github.com/BigJiaXuan/synjonesSDK/internal"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var RequestMethod = map[string]string{
	"method":       "",
	"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	"format":       "json",
	"app_key":      "",
	"access_token": "",
	"v":            "2.0",
	"sign_method":  "rsa",
	"request":      "",
}

type Conf struct {
	URL         string //tsm 地址
	AppKey      string
	Des3Key     string
	SvcPkcs8Key string
}

type Request interface {
	Send(ctx context.Context, token, request, method string) (code, resp string, err error)
}
type RequestImpl struct {
	conf    *Conf
	encrypt internal.EncryptUtil
}

func NewRequestImpl(conf *Conf) *RequestImpl {
	encrypt := internal.NewEncryptUtilImpl(conf.Des3Key)
	return &RequestImpl{conf: conf, encrypt: encrypt}
}

func (r *RequestImpl) Send(ctx context.Context, token, request, method string) (code, resp string, err error) {
	// 1.对request参数签名
	request = r.encrypt.SignRequest(request)
	// 2.对sign排序
	sign := r.sortSign(token, request, method)
	// 3.给sign签名
	sign = r.encrypt.SignatureSign(sign, r.conf.SvcPkcs8Key)
	// 准备http请求体
	req := RequestMethod
	req["method"] = method
	req["app_key"] = r.conf.AppKey
	req["access_token"] = token
	req["sign"] = sign
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
	fmt.Println("处理前的response", string(body))
	code, resp = r.decodeResponse(string(body))
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
func (r *RequestImpl) sortSign(token, request, method string) string {
	// 声明一个sign结构体
	sign := RequestMethod
	sign["access_token"] = token
	sign["app_key"] = r.conf.AppKey
	sign["method"] = method
	sign["request"] = request

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
func (r *RequestImpl) decodeResponse(response string) (errcode, request string) {
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
