// Package policy defines safety policies for SeaweedFS operations.
// It includes retry rules, backoff strategies, and limits on chunks and pages.
// 为 SeaweedFS 操作定义安全策略, 包括重试规则、回退策略和分块/分页限制.
package policy

import (
	"context"
	"errors"
	"net"
	"time"
)

// SafetyPolicy defines safety rules for SeaweedFS operations.
// It includes maximum upload retries, backoff durations, maximum download chunks, and maximum list pages.
// 定义 SeaweedFS 操作的安全策略, 包括最大上传重试次数、回退时间、最大下载分块数和最大列表页数.
type SafetyPolicy struct {
	UploadMaxRetry    int           // Maximum upload retry attempts / 上传最大重试次数
	BackoffBase       time.Duration // Base backoff duration / 回退基准时间
	BackoffMax        time.Duration // Maximum backoff duration / 最大回退时间
	MaxDownloadChunks int           // Maximum number of download chunks / 最大下载分块数
	MaxListPages      int           // Maximum number of pages in list operations / 最大列表页数
}

// DefaultSafetyPolicy returns the default safety policy. 返回默认安全策略.
func DefaultSafetyPolicy() SafetyPolicy {
	return SafetyPolicy{
		UploadMaxRetry:    3,
		BackoffBase:       200 * time.Millisecond,
		BackoffMax:        5 * time.Second,
		MaxDownloadChunks: 64,
		MaxListPages:      1000,
	}
}

// ============ Retry Decision ============

// ShouldRetryUpload determines whether an upload error is retryable.
// It returns false for context errors, certain HTTP status codes, and true for network errors.
// 判断上传错误是否可重试. 对于 context 错误、特定 HTTP 状态码返回 false, 对于网络错误返回 true.
func ShouldRetryUpload(err error) bool {
	if err == nil {
		return false
	}

	// Context errors: never retry / context 错误: 绝不重试
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// HTTP status code (if available) / HTTP 状态码 (如果存在)
	type httpStatusError interface {
		StatusCode() int
	}

	var se httpStatusError
	if errors.As(err, &se) {
		code := se.StatusCode()
		if code >= 400 && code < 500 {
			return code == 408 || code == 429
		}
	}

	// Network errors: retryable / 网络错误: 可重试
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return true
}
