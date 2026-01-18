// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It supports uploading, downloading, and other file operations.
// 提供 SeaweedFS 的 Go 客户端, 支持上传、下载及其他文件操作.
package seaweedfs

import (
	"net/http"
	"strings"
	"time"

	"github.com/GoFurry/seaweedfs-sdk-go/internal/policy"
)

// UploadMethod represents the HTTP method used for uploading. 表示上传使用的 HTTP 方法.
type UploadMethod string

const (
	UploadMethodPut  UploadMethod = "PUT"
	UploadMethodPost UploadMethod = "POST"
)

// UploadLargeOptions defines options for large file uploads.
// It includes maximum retry count per chunk and whether to use offset for resumable uploads.
// 定义大文件上传的选项, 包括每个分块最大重试次数和是否使用 offset 断点续传.
type UploadLargeOptions struct {
	MaxRetry  int  // Maximum retry attempts per chunk / 每个 chunk 最大重试次数
	UseOffset bool // Use offset for resumable upload / 是否使用 offset 断点续传
}

// ============ SeaweedFS Service ============

// SeaweedFSService represents a SeaweedFS client service.
// SeaweedFS 客户端服务.
type SeaweedFSService struct {
	FilerEndpoint string
	client        *http.Client
	policy        policy.SafetyPolicy
}

// DefaultSeaweedFSClient creates a default HTTP client for SeaweedFS with reasonable timeouts and connection limits.
// 创建一个默认的 HTTP 客户端, 包含合理的超时和连接限制.
func DefaultSeaweedFSClient() *http.Client {
	return &http.Client{
		Timeout: 300 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// NewSeaweedFSService creates a new SeaweedFSService instance with default HTTP client and safety policy.
// 创建 SeaweedFSService 实例, 使用默认 HTTP 客户端和安全策略.
func NewSeaweedFSService(endpoint string) *SeaweedFSService {
	return &SeaweedFSService{
		FilerEndpoint: strings.TrimRight(endpoint, "/"),
		client:        DefaultSeaweedFSClient(),
		policy:        policy.DefaultSafetyPolicy(),
	}
}

// NewSeaweedFSServiceWithClient creates a new SeaweedFSService instance with a custom HTTP client
// and optional configuration options.
// 创建 SeaweedFSService 实例, 使用自定义 HTTP 客户端和可选配置.
func NewSeaweedFSServiceWithClient(endpoint string, client *http.Client, opts ...Option) *SeaweedFSService {
	if client == nil {
		client = DefaultSeaweedFSClient()
	}
	s := &SeaweedFSService{
		FilerEndpoint: strings.TrimRight(endpoint, "/"),
		client:        client,
		policy:        policy.DefaultSafetyPolicy(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option defines a functional option for customizing SeaweedFSService behavior.
// 定义用于定制 SeaweedFSService 行为的函数选项.
type Option func(*SeaweedFSService)

// WithSafetyPolicy replaces the entire safety policy (for advanced users). 替换整个安全策略
func WithSafetyPolicy(p policy.SafetyPolicy) Option {
	return func(s *SeaweedFSService) {
		s.policy = p
	}
}

// WithMaxDownloadChunks sets the maximum number of download chunks. 设置最大下载分块数.
func WithMaxDownloadChunks(n int) Option {
	return func(s *SeaweedFSService) {
		if n > 0 {
			s.policy.MaxDownloadChunks = n
		}
	}
}

// WithMaxListPages sets the maximum number of pages returned in list operations. 设置列表操作的最大页数.
func WithMaxListPages(n int) Option {
	return func(s *SeaweedFSService) {
		if n > 0 {
			s.policy.MaxListPages = n
		}
	}
}

// WithUploadMaxRetry sets the maximum retry attempts for uploads. 设置上传最大重试次数.
func WithUploadMaxRetry(n int) Option {
	return func(s *SeaweedFSService) {
		if n >= 0 {
			s.policy.UploadMaxRetry = n
		}
	}
}

// WithBackoff sets the base and maximum backoff durations for retries. 设置重试的基准回退时间和最大回退时间.
func WithBackoff(base, max time.Duration) Option {
	return func(s *SeaweedFSService) {
		if base > 0 {
			s.policy.BackoffBase = base
		}
		if max > 0 {
			s.policy.BackoffMax = max
		}
	}
}
