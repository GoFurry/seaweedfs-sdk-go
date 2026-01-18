// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It includes functions for uploading files, including large file uploads.
// 提供 SeaweedFS 的 Go 客户端, 包括上传文件和大文件分片上传功能.
package seaweedfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/GoFurry/seaweedfs-sdk-go/internal/policy"
	"github.com/GoFurry/seaweedfs-sdk-go/internal/util"
)

// UploadWithOptions uploads a file to SeaweedFS with optional parameters and headers.
// 上传文件, 支持 SeaweedFS 可选参数和 HTTP Header.
func (s *SeaweedFSService) UploadWithOptions(
	ctx context.Context,
	method UploadMethod, // HTTP method: PUT or POST / HTTP 方法: PUT 或 POST
	dst string, // Destination path / 目标路径
	r io.Reader, // Source reader / 数据源
	opts map[string]string, // Optional query parameters / 可选查询参数
	headers map[string]string, // Optional HTTP headers / 可选 HTTP 头
) error {

	dst = util.NormalizePath(dst)

	u := s.FilerEndpoint + dst
	if len(opts) > 0 {
		q := url.Values{}
		for k, v := range opts {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, string(method), u, r)
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s %s", resp.Status, b)
	}

	return nil
}

// UploadLarge uploads a large file in chunks, supporting retry and backoff policies.
// 分片上传大文件, 支持重试和回退策略.
func (s *SeaweedFSService) UploadLarge(
	ctx context.Context,
	method UploadMethod, // HTTP method / HTTP 方法
	dst string, // Destination path / 目标路径
	r io.Reader, // Source reader / 数据源
	size int64, // Total size of the file / 文件总大小
	chunkSize int64, // Size of each chunk / 每个分片大小
	opts map[string]string, // Optional query parameters / 可选查询参数
	headers map[string]string, // Optional HTTP headers / 可选 HTTP 头
	largeOpt *UploadLargeOptions, // Options for large upload / 大文件上传选项
) error {

	dst = util.NormalizePath(dst)

	if chunkSize <= 0 {
		chunkSize = 10 << 20 // 默认 10MB
	}

	if largeOpt == nil {
		largeOpt = &UploadLargeOptions{
			MaxRetry: 3,
		}
	}

	if largeOpt.MaxRetry > s.policy.UploadMaxRetry {
		largeOpt.MaxRetry = s.policy.UploadMaxRetry
	}

	var uploaded int64
	buf := make([]byte, chunkSize)

	for uploaded < size {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		readSize := chunkSize
		if size-uploaded < chunkSize {
			readSize = size - uploaded
		}

		n, err := io.ReadFull(r, buf[:readSize])
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		if n == 0 {
			break
		}

		var lastErr error

		for attempt := 0; attempt <= largeOpt.MaxRetry; attempt++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			chunkReader := bytes.NewReader(buf[:n])
			chunkOpts := make(map[string]string)
			for k, v := range opts {
				chunkOpts[k] = v
			}

			if largeOpt.UseOffset {
				chunkOpts["offset"] = strconv.FormatInt(uploaded, 10)
			} else if uploaded > 0 {
				chunkOpts["op"] = "append"
			}

			err = s.UploadWithOptions(ctx, method, dst, chunkReader, chunkOpts, headers)
			if err == nil {
				lastErr = nil
				break
			}

			if !policy.ShouldRetryUpload(err) {
				return fmt.Errorf("upload chunk failed (no retry) at offset=%d: %w", uploaded, err)
			}

			lastErr = err

			sleep := s.policy.BackoffBase * (1 << attempt)
			if sleep > s.policy.BackoffMax {
				sleep = s.policy.BackoffMax
			}
			sleep = time.Duration(float64(sleep) * (0.5 + rand.Float64()/2))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sleep):
			}
		}

		if lastErr != nil {
			return fmt.Errorf("upload chunk failed at offset=%d after %d retries: %w",
				uploaded, largeOpt.MaxRetry, lastErr)
		}

		uploaded += int64(n)
	}

	return nil
}

// UploadFileSmart intelligently uploads a file, choosing small or large upload automatically.
// 智能上传文件, 根据文件大小选择普通上传或分片上传.
func (s *SeaweedFSService) UploadFileSmart(
	ctx context.Context,
	method UploadMethod, // HTTP method / HTTP 方法
	dst string, // Destination path / 目标路径
	fh *multipart.FileHeader, // FileHeader from frontend / 前端文件头
	largeThreshold int64, // Threshold for large file / 大文件阈值
	chunkSize int64, // Chunk size / 分片大小
	opts map[string]string, // Optional query parameters / 可选查询参数
	headers map[string]string, // Optional HTTP headers / 可选 HTTP 头
) error {

	file, err := fh.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		file.Seek(0, io.SeekStart)
	}
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = contentType

	if fh.Size <= largeThreshold {
		return s.UploadWithOptions(ctx, method, dst, file, opts, headers)
	} else {
		return s.UploadLarge(ctx, method, dst, file, fh.Size, chunkSize, opts, headers,
			&UploadLargeOptions{
				MaxRetry:  3,
				UseOffset: true,
			})
	}
}
