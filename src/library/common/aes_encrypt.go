package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"bytes"
	"encoding/base64"
	"fmt"
)

const (
	WEB_ENCRY_KEY  = "4aa170e712535eee3631e86db9506ad5"
	APP_ENCRY_KEY  = "419d1497814f1d2ef85c85cc3c9e0604"
	MASTER_KEY = "fjorjC7peqQtmrAeBBNa3DZ1XiufZ47f"
	AGENT_KEY = "3pZwiWYu8lscvjOuWYOfxKVmm9r0PNcZ"
)

// 偏移量的写法
func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return []byte{}
	}
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte{}
	}
	return origData[:(length - unpadding)]
}

func WebPwdDecrypt(crypted string) string {
	keyword := []byte(WEB_ENCRY_KEY)
	encrypted, _ := base64.StdEncoding.DecodeString (crypted)
	// encrypted必须是块大小的倍数
	if len(encrypted) % 16 != 0 {
		fmt.Println("密码长度不对")
		return ""
	}
	block, err := aes.NewCipher(keyword)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, keyword[:blockSize])
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = PKCS5UnPadding(origData)
	return string(origData)
}

func WebPwdEncrypt(pwd string) string {
	key := []byte(WEB_ENCRY_KEY)
	origData := []byte(pwd)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted)
}

func AesEncrypt(pwd string) string {
	key := []byte(APP_ENCRY_KEY)
	origData := []byte(pwd)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return hex.EncodeToString(crypted)
}

func AesDecrypt(crypted_pwd string) string {
	key := []byte(APP_ENCRY_KEY)
	crypted, _ := hex.DecodeString(crypted_pwd)
	// crypted必须是块大小的倍数
	if len(crypted) % 16 != 0 {
		fmt.Println("密码长度不对")
		return ""
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return string(origData)
}

func AgentEncrypt(pwd string) string {
	key := []byte(AGENT_KEY)
	origData := []byte(pwd)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return hex.EncodeToString(crypted)
}

func MasterDecrypt(crypted_pwd string) string {
	key := []byte(MASTER_KEY)
	crypted, _ := hex.DecodeString(crypted_pwd)
	// crypted必须是块大小的倍数
	if len(crypted) % 16 != 0 {
		fmt.Println("密码长度不对")
		return ""
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return string(origData)
}