package zcrypto

import (
	"fmt"
	"testing"
)

func TestAESEncryptCBC(t *testing.T) {
	var (
		message = "666ddrobotwoie878oasnx"
		key     = "0b1c8c60f6bb4fe1"
		iv      = "8028dbc1021b5712"
	)
	encrypted := AESEncryptCBC(message, key, iv)
	fmt.Println("加密数据:", encrypted)

	decrypted := AESDecryptCBC(encrypted, key, iv)
	fmt.Println("解密数据:", decrypted)
}
