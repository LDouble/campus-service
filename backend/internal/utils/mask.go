package utils

import (
	"strings"
)

// MaskPhone 脱敏手机号
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskStudentID 脱敏学号
func MaskStudentID(studentID string) string {
	if len(studentID) <= 4 {
		return studentID
	}
	return studentID[:2] + strings.Repeat("*", len(studentID)-4) + studentID[len(studentID)-2:]
}

// MaskName 脱敏姓名
func MaskName(name string) string {
	if len(name) <= 1 {
		return name
	}
	if len(name) == 2 {
		return string([]rune(name)[0]) + "*"
	}
	runes := []rune(name)
	return string(runes[0]) + strings.Repeat("*", len(runes)-2) + string(runes[len(runes)-1])
}