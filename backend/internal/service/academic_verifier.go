package service

import (
	"context"
)

// AcademicVerifier 教务系统验证器接口
type AcademicVerifier interface {
	// VerifyCredentials 验证学号和密码
	VerifyCredentials(ctx context.Context, studentID, password, school string) (*StudentInfo, error)
}

// StudentInfo 学生信息
type StudentInfo struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	School    string `json:"school"`
	Major     string `json:"major"`
	Grade     string `json:"grade"`
}