package internal

import (
	"crypto"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/golang-module/dongle"
)

//提供tsm发送请求前参数处理的签名办法

type EncryptUtil interface {
	// SignRequest 对称签名request参数
	SignRequest(text string) string
	// SignatureSign 签名生成sign
	SignatureSign(text, key string) string
	// DecryptResponse 解密request
	DecryptResponse(request string) string
	// EncodeByBase64 Base64编码
	EncodeByBase64(text string) string
	// DecodeByBase64 Base64解码
	DecodeByBase64(text string) string
}
type EncryptUtilImpl struct {
	cipher *dongle.Cipher
}

func NewEncryptUtilImpl(key string) *EncryptUtilImpl {
	cipher := dongle.NewCipher()
	cipher.SetKey(dongle.Decode.FromString(key).ByBase64().ToString())
	cipher.SetMode(dongle.CBC)
	cipher.SetIV(make([]byte, des.BlockSize))
	cipher.SetPadding(dongle.PKCS5)
	return &EncryptUtilImpl{cipher: cipher}
}

// SignRequest
//
//	@Description: 使用DESede/CBC/PKCS5Padding,iv向量位8字节的16进制0 进行对称加密，再进行base64编码
//	@receiver e
//	@param request 待签名参数
//	@param key 签名密钥
//	@return string
//	@return error
func (e *EncryptUtilImpl) SignRequest(text string) string {
	return dongle.Encrypt.FromString(text).By3Des(e.cipher).ToBase64String()
}

// SignatureSign
//
//	@Description: 对sign签名
//	@receiver e
//	@param text
//	@param key
//	@return string
func (e *EncryptUtilImpl) SignatureSign(text, key string) string {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return ""
	}
	private, err := x509.ParsePKCS8PrivateKey(block.Bytes) //之前看java demo中使用的是pkcs8
	if err != nil {
		return ""
	}
	h := crypto.Hash.New(crypto.SHA1) //进行SHA1的散列
	h.Write([]byte(text))
	hashed := h.Sum(nil)
	// 进行rsa加密签名
	signedData, err := rsa.SignPKCS1v15(rand.Reader, private.(*rsa.PrivateKey), crypto.SHA1, hashed)
	data := base64.StdEncoding.EncodeToString(signedData)
	return data
}

// DecryptResponse
//
//	@Description: 解码request参数
//	@receiver e
//	@param request
//	@return string
func (e *EncryptUtilImpl) DecryptResponse(request string) string {
	return dongle.Decrypt.FromBase64String(request).By3Des(e.cipher).ToString()
}

// EncodeByBase64
//
//	@Description: base64解码
//	@receiver e
//	@param text
//	@return string
func (e *EncryptUtilImpl) EncodeByBase64(text string) string {
	return dongle.Encode.FromString(text).ByBase64().ToString()
}

// DecodeByBase64
//
//	@Description: base64编码
//	@receiver e
//	@param text
//	@return string
func (e *EncryptUtilImpl) DecodeByBase64(text string) string {
	return dongle.Decode.FromString(text).ByBase64().ToString()
}
