// Package seaweedfs provides utility functions for the SeaweedFS Go client.
// 提供 SeaweedFS Go 客户端的辅助工具函数
package seaweedfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ==========================
// 文件操作 / File Utilities
// ==========================

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

// FileExists checks whether a local file exists.
// 检查本地文件是否存在.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir checks whether a path is a directory.
// 判断路径是否为目录.
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// ListFilesInDir lists all files in a directory. If recursive is true, it traverses subdirectories.
// 列出目录下所有文件, recursive 为 true 时递归子目录.
func ListFilesInDir(path string, recursive bool) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, p)
		} else if !recursive && p != path {
			return filepath.SkipDir
		}
		return nil
	})
	return files, err
}

// CleanupFiles deletes a list of files, ignoring errors if a file does not exist.
// 批量删除文件, 如果文件不存在则忽略.
func CleanupFiles(paths []string) {
	for _, p := range paths {
		_ = os.Remove(p)
	}
}

// SplitFileOffsets splits a file of given size into chunks of chunkSize.
// Returns a slice of [start, end] pairs for each chunk.
// 将文件按指定 chunkSize 分片, 返回每个分片的 [start, end] 偏移量.
func SplitFileOffsets(size int64, chunkSize int64) [][2]int64 {
	if chunkSize <= 0 {
		chunkSize = 10 << 20 // 默认10MB
	}
	var offsets [][2]int64
	for start := int64(0); start < size; start += chunkSize {
		end := start + chunkSize - 1
		if end >= size {
			end = size - 1
		}
		offsets = append(offsets, [2]int64{start, end})
	}
	return offsets
}

// ReadableSize converts a size in bytes to a human-readable string like KB/MB/GB.
// 将字节大小转换为人类可读格式.
func ReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	pre := []string{"KB", "MB", "GB", "TB", "PB"}[exp]
	return fmt.Sprintf("%.2f%s", float64(size)/float64(div), pre)
}

// ==========================
// 时间处理 / Time Utilities
// ==========================

// ParseSeaweedTime parses a RFC3339-formatted string into time.Time.
// Returns an error if the string cannot be parsed.
// 将 RFC3339 格式的时间字符串解析为 time.Time. 如果解析失败, 返回错误.
func ParseSeaweedTime(t string) (time.Time, error) {
	return time.Parse(time.RFC3339, t)
}

// FormatSeaweedTime formats a time.Time into RFC3339 string for SeaweedFS.
// 将 time.Time 格式化为 SeaweedFS 使用的 RFC3339 字符串.
func FormatSeaweedTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// TimeSinceOrZero returns the duration since t. Returns 0 if t is zero.
// 计算从 t 到现在的时间间隔, t 为零值返回 0.
func TimeSinceOrZero(t time.Time) time.Duration {
	if t.IsZero() {
		return 0
	}
	return time.Since(t)
}

// ==========================
// 路径处理 / Path Utilities
// ==========================

// NormalizePath ensures the path starts with '/' and has no duplicate slashes.
// 确保路径以 '/' 开头并去除重复斜杠.
func NormalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	p = "/" + strings.TrimLeft(p, "/")
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	return p
}

// JoinPath safely joins multiple path segments, avoiding duplicate slashes.
// 安全拼接多个路径段, 避免重复斜杠.
func JoinPath(parts ...string) string {
	return NormalizePath(strings.Join(parts, "/"))
}

// TempFileName generates a temporary filename with optional prefix and suffix.
// 生成临时文件名, 可选前缀和后缀.
func TempFileName(prefix, suffix string) string {
	t := time.Now().UnixNano()
	return fmt.Sprintf("%s%d%s", prefix, t, suffix)
}
