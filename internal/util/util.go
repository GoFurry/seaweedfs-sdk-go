// Package util provides utility functions for SeaweedFS SDK operations. 为 SeaweedFS SDK 提供通用工具函数.
package util

import (
	"path"
	"strings"
	"time"
)

// DerefString returns the value of a string pointer or an empty string if the pointer is nil.
// 返回字符串指针的值, 如果指针为 nil 则返回空字符串.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NormalizePath ensures a path starts with "/" and cleans it.
// 确保路径以 "/" 开头并进行路径清理.
func NormalizePath(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return path.Clean(p)
}

// ParseSeaweedTime parses a string in RFC3339 format into time.Time.
// If parsing fails, it returns the zero value of time.Time.
// 将 RFC3339 格式的字符串解析为 time.Time, 如果解析失败返回零值.
func ParseSeaweedTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}
