// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It includes file download operations, including ranged and concurrent downloads.
// 提供 SeaweedFS 的 Go 客户端, 包括文件下载操作, 支持范围下载和并发下载.
package seaweedfs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/GoFurry/seaweedfs-sdk-go/internal/util"
)

// Download downloads a file from SeaweedFS with the default options. 使用默认选项从 SeaweedFS 下载文件.
func (s *SeaweedFSService) Download(ctx context.Context, p string, progress ProgressFunc) (io.ReadCloser, http.Header, error) {
	// Delegate to DownloadWithOptions without query or headers
	rc, header, _, err := s.DownloadWithOptions(ctx, p, nil, nil, progress)
	return rc, header, err
}

// DownloadWithOptions downloads a file with custom query parameters and headers. 使用自定义查询参数和请求头下载文件.
func (s *SeaweedFSService) DownloadWithOptions(
	ctx context.Context,
	p string,
	query map[string]string,
	headers map[string]string,
	progress ProgressFunc,
) (io.ReadCloser, http.Header, int, error) {

	// Normalize path to ensure it starts with '/' and is clean
	p = util.NormalizePath(p)

	u := s.FilerEndpoint + p
	if len(query) > 0 {
		// Build query string
		q := url.Values{}
		for k, v := range query {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	// Create HTTP GET request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	// Apply custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}

	// Translate 404 to os.ErrNotExist for Go-style handling
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil, resp.StatusCode, os.ErrNotExist
	}

	// Any other >=400 status is treated as error
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, resp.StatusCode,
			fmt.Errorf("download failed: %s %s", resp.Status, string(b))
	}

	// Callback when finished
	if progress != nil {
		progress(-1, -1)
	}

	// Caller is responsible for closing the response body
	return resp.Body, resp.Header, resp.StatusCode, nil
}

// DownloadRange downloads a specific byte range [start, end] from a file. 下载文件的指定字节范围 [start, end].
func (s *SeaweedFSService) DownloadRange(
	ctx context.Context,
	p string,
	start, end int64,
	progress ProgressFunc,
) (io.ReadCloser, http.Header, int, error) {

	// Validate range start and end
	if start < 0 {
		return nil, nil, 0, fmt.Errorf("invalid range start")
	}
	if end >= 0 && end < start {
		return nil, nil, 0, fmt.Errorf("invalid range: end < start")
	}

	// Build HTTP Range header value
	rangeValue := ""
	if end >= 0 {
		rangeValue = fmt.Sprintf("bytes=%d-%d", start, end)
	} else {
		rangeValue = fmt.Sprintf("bytes=%d-", start)
	}
	headers := map[string]string{
		"Range": rangeValue,
	}

	// Perform ranged download
	rc, hdr, status, err := s.DownloadWithOptions(ctx, p, nil, headers, progress)
	if err != nil {
		return nil, nil, 0, err
	}

	// SeaweedFS may downgrade to 200 instead of 206
	if status != http.StatusPartialContent && status != http.StatusOK {
		rc.Close()
		return nil, nil, status, fmt.Errorf("unexpected status code: %d", status)
	}

	return rc, hdr, status, nil
}

// DownloadResume resumes downloading a file from a given offset. 从指定偏移量继续下载文件.
func (s *SeaweedFSService) DownloadResume(
	ctx context.Context,
	p string,
	offset int64,
	progress ProgressFunc,
) (io.ReadCloser, http.Header, int, error) {

	// Offset <= 0 means full download
	if offset <= 0 {
		return s.DownloadWithOptions(ctx, p, nil, nil, progress)
	}

	return s.DownloadRange(ctx, p, offset, -1, progress)
}

// DownloadChunkError represents an error for a specific chunk during concurrent download. 表示并发下载中某个分块的错误.
type DownloadChunkError struct {
	File string
	Err  error
}

func (e DownloadChunkError) Error() string {
	return fmt.Sprintf("chunk %s download failed: %v", e.File, e.Err)
}

// DownloadConcurrent downloads a file in concurrent chunks and returns a map of temporary file paths to errors.
// 并发下载文件, 返回临时文件路径到错误的映射.
func (s *SeaweedFSService) DownloadConcurrent(
	ctx context.Context,
	remotePath, dstPath string,
	chunkCount int,
	progress ProgressFunc,
) map[string]error {
	result := make(map[string]error)

	// Enforce maximum allowed concurrent chunks
	if chunkCount > s.policy.MaxDownloadChunks {
		chunkCount = s.policy.MaxDownloadChunks
	}

	// Fetch remote file metadata to get size
	stat, err := s.Stat(ctx, remotePath, false)
	if err != nil {
		result[dstPath] = fmt.Errorf("stat failed: %w", err)
		return result
	}
	size := stat.Size

	// Fallback to sequential download for small files
	if chunkCount <= 1 || size < int64(chunkCount*5<<20) {
		select {
		case <-ctx.Done():
			result[dstPath] = ctx.Err()
			return result
		default:
		}

		rc, _, _, err := s.DownloadResume(ctx, remotePath, 0, progress)
		if err != nil {
			result[dstPath] = err
			return result
		}
		defer rc.Close()

		f, err := os.Create(dstPath)
		if err != nil {
			result[dstPath] = err
			return result
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		result[dstPath] = err
		return result
	}

	// Calculate chunk size
	chunkSize := size / int64(chunkCount)
	tempFiles := make([]string, chunkCount)
	errs := make(chan DownloadChunkError, chunkCount)
	wg := sync.WaitGroup{}

	// Mutex for callback
	var totalDownloaded int64
	var mu sync.Mutex

	for i := 0; i < chunkCount; i++ {
		select {
		case <-ctx.Done():
			errs <- DownloadChunkError{
				File: fmt.Sprintf("%s.part%d", dstPath, i),
				Err:  ctx.Err(),
			}
			continue
		default:
		}

		wg.Add(1)
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == chunkCount-1 {
			end = size - 1
		}

		tmp := fmt.Sprintf("%s.part%d", dstPath, i)
		tempFiles[i] = tmp

		go func(start, end int64, tmp string) {
			defer wg.Done()

			// Abort early if context is cancelled
			select {
			case <-ctx.Done():
				errs <- DownloadChunkError{File: tmp, Err: ctx.Err()}
				return
			default:
			}

			// Download chunk range
			rc, _, _, err := s.DownloadRange(ctx, remotePath, start, end, nil)
			if err != nil {
				errs <- DownloadChunkError{File: tmp, Err: err}
				return
			}
			defer rc.Close()

			// Write chunk to temp file
			select {
			case <-ctx.Done():
				errs <- DownloadChunkError{File: tmp, Err: ctx.Err()}
				return
			default:
			}

			f, err := os.Create(tmp)
			if err != nil {
				errs <- DownloadChunkError{File: tmp, Err: err}
				return
			}
			defer f.Close()

			// Write and update progress
			n, err := io.Copy(f, rc)
			mu.Lock()
			totalDownloaded += n
			if progress != nil {
				progress(totalDownloaded, size)
			}
			mu.Unlock()

			errs <- DownloadChunkError{File: tmp, Err: err}
		}(start, end, tmp)
	}

	wg.Wait()
	close(errs)

	// Collect all chunk results
	for e := range errs {
		result[e.File] = e.Err
	}

	return result
}
