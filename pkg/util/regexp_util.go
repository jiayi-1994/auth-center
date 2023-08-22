package util

import (
	"fmt"
	"regexp"
)

func ValidatePassword(password string) bool {
	// 密码长度在8到20之间且需包含至少一个大写字符，一个小写字符和一个数字
	if len(password) < 8 || len(password) > 20 {
		fmt.Println("Password length must be between 8 and 20.")
		return false
	}

	// 密码必须包含至少一个大写字母
	uppercase := regexp.MustCompile("[A-Z]")
	if !uppercase.MatchString(password) {
		fmt.Println("Password must contain at least one uppercase letter.")
		return false
	}

	// 密码必须包含至少一个小写字母
	lowercase := regexp.MustCompile("[a-z]")
	if !lowercase.MatchString(password) {
		fmt.Println("Password must contain at least one lowercase letter.")
		return false
	}

	// 密码必须包含至少一个数字
	number := regexp.MustCompile("[0-9]")
	if !number.MatchString(password) {
		fmt.Println("Password must contain at least one number.")
		return false
	}
	return true
}
