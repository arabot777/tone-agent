package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
)

var SecurityKey = "xQm9P4sW2bNcR8tYvG3zJhL6kDpF7wXe5"

// 加密
func AesEncryptCBC(origDataStr, keyStr string) (encryptedStr string) {

	origData := []byte(origDataStr) // 待加密的数据
	key := []byte(keyStr)           // 加密的密钥

	// 分组秘钥
	// NewCipher该函数限制了输入k的长度必须为16, 24或者32
	block, _ := aes.NewCipher(key)
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	origData = pkcs5Padding(origData, blockSize)                // 补全码
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize]) // 加密模式
	encrypted := make([]byte, len(origData))                    // 创建数组
	blockMode.CryptBlocks(encrypted, origData)                  // 加密
	return hex.EncodeToString(encrypted)
}

// 解密
func AesDecryptCBC(encryptedStr, keyStr string) (decryptedStr string) {
	if encryptedStr == "" {
		return ""
	}

	encrypted, _ := hex.DecodeString(encryptedStr) // 待解密的数据
	key := []byte(keyStr)                          // 解密的密钥

	block, _ := aes.NewCipher(key)                              // 分组秘钥
	blockSize := block.BlockSize()                              // 获取秘钥块的长度
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize]) // 加密模式
	decrypted := make([]byte, len(encrypted))                   // 创建数组
	blockMode.CryptBlocks(decrypted, encrypted)                 // 解密
	decrypted = pkcs5UnPadding(decrypted)                       // 去除补全码
	return string(decrypted)
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func AesEncryptCBCBase64(origDataStr, keyStr string) (encryptedStr string) {
	encryptedStr = AesEncryptCBC(origDataStr, keyStr)
	encrypted, _ := hex.DecodeString(encryptedStr)
	encryptedBase64 := base64.URLEncoding.EncodeToString(encrypted)
	return encryptedBase64
}

func AesDecryptCBCBase64(encryptedBase64Str, keyStr string) (decryptedStr string) {
	defer func() {
		if err := recover(); err != nil {
			decryptedStr = ""
		}
	}()
	encrypted, _ := base64.URLEncoding.DecodeString(encryptedBase64Str)
	encryptedHex := hex.EncodeToString(encrypted)
	return AesDecryptCBC(encryptedHex, keyStr)
}
