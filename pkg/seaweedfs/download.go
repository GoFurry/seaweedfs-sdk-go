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
func (s *SeaweedFSService) Download(ctx context.Context, p string) (io.ReadCloser, http.Header, error) {
	rc, header, _, err := s.DownloadWithOptions(ctx, p, nil, nil)
	return rc, header, err
}

// DownloadWithOptions downloads a file with custom query parameters and headers. 使用自定义查询参数和请求头下载文件.
func (s *SeaweedFSService) DownloadWithOptions(
	ctx context.Context,
	p string,
	query map[string]string,
	headers map[string]string,
) (io.ReadCloser, http.Header, int, error) {

	p = util.NormalizePath(p)

	u := s.FilerEndpoint + p
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, nil, resp.StatusCode, os.ErrNotExist
	}

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, resp.StatusCode,
			fmt.Errorf("download failed: %s %s", resp.Status, string(b))
	}

	return resp.Body, resp.Header, resp.StatusCode, nil
}

// DownloadRange downloads a specific byte range [start, end] from a file. 下载文件的指定字节范围 [start, end].
func (s *SeaweedFSService) DownloadRange(
	ctx context.Context,
	p string,
	start, end int64,
) (io.ReadCloser, http.Header, int, error) {

	if start < 0 {
		return nil, nil, 0, fmt.Errorf("invalid range start")
	}
	if end >= 0 && end < start {
		return nil, nil, 0, fmt.Errorf("invalid range: end < start")
	}

	rangeValue := ""
	if end >= 0 {
		rangeValue = fmt.Sprintf("bytes=%d-%d", start, end)
	} else {
		rangeValue = fmt.Sprintf("bytes=%d-", start)
	}

	headers := map[string]string{
		"Range": rangeValue,
	}

	rc, hdr, status, err := s.DownloadWithOptions(ctx, p, nil, headers)
	if err != nil {
		return nil, nil, 0, err
	}

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
) (io.ReadCloser, http.Header, int, error) {

	if offset <= 0 {
		return s.DownloadWithOptions(ctx, p, nil, nil)
	}

	return s.DownloadRange(ctx, p, offset, -1)
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
) map[string]error {
	result := make(map[string]error)

	if chunkCount > s.policy.MaxDownloadChunks {
		chunkCount = s.policy.MaxDownloadChunks
	}

	stat, err := s.Stat(ctx, remotePath, false)
	if err != nil {
		result[dstPath] = fmt.Errorf("stat failed: %w", err)
		return result
	}
	size := stat.Size

	if chunkCount <= 1 || size < int64(chunkCount*5<<20) {
		select {
		case <-ctx.Done():
			result[dstPath] = ctx.Err()
			return result
		default:
		}

		rc, _, _, err := s.DownloadResume(ctx, remotePath, 0)
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

	chunkSize := size / int64(chunkCount)
	tempFiles := make([]string, chunkCount)
	errs := make(chan DownloadChunkError, chunkCount)
	wg := sync.WaitGroup{}

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

			select {
			case <-ctx.Done():
				errs <- DownloadChunkError{File: tmp, Err: ctx.Err()}
				return
			default:
			}

			rc, _, _, err := s.DownloadRange(ctx, remotePath, start, end)
			if err != nil {
				errs <- DownloadChunkError{File: tmp, Err: err}
				return
			}
			defer rc.Close()

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

			_, err = io.Copy(f, rc)
			errs <- DownloadChunkError{File: tmp, Err: err}
		}(start, end, tmp)
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		result[e.File] = e.Err
	}

	return result
}

// MergeFiles merges multiple files in order into a target file.
// If cleanup is true, source files will be deleted after merging.
// 将多个文件按顺序合并到目标文件, cleanup 为 true 时删除源分片.
func MergeFiles(outputPath string, parts []string, cleanup bool) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, p := range parts {
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return err
		}
	}

	if cleanup {
		for _, p := range parts {
			_ = os.Remove(p)
		}
	}

	return nil
}
