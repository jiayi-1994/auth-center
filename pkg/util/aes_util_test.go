package util

import "testing"

func TestEncryptByAes(t *testing.T) {
	aes, err := EncryptByAes([]byte("test11111"))
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(aes)

	byAes, err := DecryptByAes(aes)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(string(byAes))
}
