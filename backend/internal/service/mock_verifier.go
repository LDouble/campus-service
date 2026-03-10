package service

import (
	"context"
	"errors"
	"strings"
)

// mockVerifier 模拟教务系统验证器
type mockVerifier struct{}

// NewMockVerifier 创建模拟验证器实例
func NewMockVerifier() AcademicVerifier {
	return &mockVerifier{}
}

// VerifyCredentials 模拟验证学号和密码
func (m *mockVerifier) VerifyCredentials(ctx context.Context, studentID, password, school string) (*StudentInfo, error) {
	// 模拟验证逻辑
	if studentID == "" || password == "" || school == "" {
		return nil, errors.New("学号、密码和学校不能为空")
	}

	// 模拟验证失败的情况
	if password == "wrong" {
		return nil, errors.New("学号或密码错误")
	}

	// 模拟不同学校的学生信息
	var name, major, grade string
	switch strings.ToLower(school) {
	case "清华大学":
		name = "张三"
		major = "计算机科学与技术"
		grade = "2021级"
	case "北京大学":
		name = "李四"
		major = "软件工程"
		grade = "2020级"
	default:
		name = "王五"
		major = "信息管理与信息系统"
		grade = "2022级"
	}

	return &StudentInfo{
		StudentID: studentID,
		Name:      name,
		School:    school,
		Major:     major,
		Grade:     grade,
	}, nil
}