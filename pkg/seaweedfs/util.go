// Package seaweedfs provides utility functions for the SeaweedFS Go client.
// 提供 SeaweedFS Go 客户端的辅助工具函数。
package seaweedfs

import (
	"fmt"
	"os"
	"time"
)

// LocalFileSize returns the size of a local file in bytes.
// Returns an error if the file does not exist or is not a regular file.
// 返回本地文件的大小 (字节), 如果文件不存在或不是普通文件, 则返回错误.
func LocalFileSize(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !fi.Mode().IsRegular() {
		return 0, fmt.Errorf("not a regular file")
	}
	return fi.Size(), nil
}

// ParseSeaweedTime parses a RFC3339-formatted string into time.Time.
// Returns an error if the string cannot be parsed.
// 将 RFC3339 格式的时间字符串解析为 time.Time. 如果解析失败, 则返回错误.
func ParseSeaweedTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}
