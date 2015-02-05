package server

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// 处理输入的用户名，使得其在标准输出中显示长度一致
// 中文占3字节，输出时则只占2个英文字长度
// 不足15字节则前缀空格补齐(leftOrRight为真)或后缀空格补齐(leftOrRight为假)
func constructUserName(name string, length int, leftOrRight bool) string {
	zhCharNum := (len(name) - utf8.RuneCountInString(name)) / 2 // 中文字符数
	charNum := len(name) - zhCharNum*3                          // 英文字符数（ASII）
	spaceNum := length - zhCharNum*2 - charNum                  // 应该添加的空格字符数

	if spaceNum < 0 {
		fmt.Println("处理用户名时发生错误")
	}

	str := ""
	if leftOrRight {
		str = strings.Repeat(" ", spaceNum) + name
	} else {
		str = name + strings.Repeat(" ", spaceNum)
	}

	return str
}

// 判断字符串数组切片中是否有指定的值
func hasValue(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
