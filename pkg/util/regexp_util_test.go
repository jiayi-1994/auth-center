package util

import (
	"fmt"
	"testing"
)

func Test_validatePassword(t *testing.T) {
	passwords := []string{
		"Abc12345",     // 符合要求
		"abcd1234",     // 缺少大写字符
		"ABCD1234",     // 缺少小写字符
		"abcdEFGH",     // 缺少数字
		"123456789012", // 超过长度限制
	}

	for _, password := range passwords {
		valid := ValidatePassword(password)
		fmt.Printf("密码 '%s' 是否有效: %v\n", password, valid)
	}
}
