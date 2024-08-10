package zcrypto

import "testing"

func TestMD5Sum(t *testing.T) {
	s := "hello"
	t.Log(MD5Sum(s))
	t.Log(SHA1Sum(s))
	t.Log(SHA256Sum(s))
	t.Log(SHA512Sum(s))

	key := "world"
	t.Log(HmacMD5(s, key))
	t.Log(HmacSHA1(s, key))
	t.Log(HmacSHA256(s, key))
	t.Log(HmacSHA512(s, key))
}
