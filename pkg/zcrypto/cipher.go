package zcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// -------------------------*------------------------- symmetrical encryption -------------------------#-------------------------

// -------------------------*------------------------- CBC mode  PKCS7 padding -------------------------#-------------------------

func AESEncryptCBC(message, key string, IV ...string) string {
	var iv string
	if len(IV) > 0 {
		iv = IV[0]
	}
	ciphertext, err := aesEncryptCBC([]byte(message), []byte(key), []byte(iv))
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func AESDecryptCBC(encodedData, key string, IV ...string) string {
	var iv string
	if len(IV) > 0 {
		iv = IV[0]
	}
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return ""
	}
	plaintext, err := aesDecryptCBC(decodedData, []byte(key), []byte(iv))
	if err != nil {
		return ""
	}
	return string(plaintext)
}

func aesEncryptCBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(iv) == 0 {
		iv = key[:blockSize]
	} else if len(iv) != blockSize {
		return nil, fmt.Errorf("IV length must be equal to block size")
	}

	data = PKCS7Padding(data, blockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)
	return ciphertext, nil
}

func aesDecryptCBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	if len(iv) == 0 {
		iv = key[:blockSize]
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	return PKCS7UnPadding(plaintext)
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func PKCS7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("invalid padding")
	}
	return data[:(length - unpadding)], nil
}

// -------------------------*------------------------- public key encryption -------------------------#-------------------------
