// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It includes file system operations such as mkdir, delete, move, copy, and listing directories.
// 提供 SeaweedFS 的 Go 客户端, 包括文件系统操作, 如创建目录、删除、移动、复制及列出目录.
package seaweedfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/GoFurry/seaweedfs-sdk-go/internal/util"
)

// Mkdir creates a directory in SeaweedFS. 在 SeaweedFS 中创建目录.
func (s *SeaweedFSService) Mkdir(ctx context.Context, dir string) error {
	dir = util.NormalizePath(dir)
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.FilerEndpoint+dir, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mkdir failed: %s", resp.Status)
	}
	return nil
}

// Delete removes a file or directory.
// extra allows passing additional parameters compatible with SeaweedFS official API (e.g., recursive, skipChunkDeletion).
// 删除文件或目录, extra 用于传递可选参数, 兼容官方 API (如 recursive, skipChunkDeletion)
func (s *SeaweedFSService) Delete(ctx context.Context, p string, extra map[string]string) error {
	p = util.NormalizePath(p)

	q := make(url.Values)
	for k, v := range extra {
		q.Set(k, v)
	}

	u := s.FilerEndpoint + p
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete failed: %s %s", resp.Status, b)
}

// DeleteBatch deletes multiple files concurrently or sequentially depending on concurrency parameter.
// ignoreErrors indicates whether to continue on error or record them in the result map.
// 根据 concurrency 参数批量删除文件, ignoreErrors 表示是否忽略错误.
func (s *SeaweedFSService) DeleteBatch(
	ctx context.Context,
	paths []string,
	extra map[string]string,
	ignoreErrors bool,
	concurrency int,
) map[string]error {
	if concurrency <= 1 {
		results := make(map[string]error, len(paths))
		for _, p := range paths {
			err := s.Delete(ctx, p, extra)
			if ignoreErrors {
				results[p] = nil
			} else {
				results[p] = err
			}
		}
		return results
	}

	results := make(map[string]error, len(paths))
	mu := sync.Mutex{}
	sem := make(chan struct{}, concurrency)
	wg := sync.WaitGroup{}

	for _, p := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()
			err := s.Delete(ctx, path, extra)
			if ignoreErrors {
				err = nil
			}
			mu.Lock()
			results[path] = err
			mu.Unlock()
		}(p)
	}

	wg.Wait()
	return results
}

// Move renames or moves a file or directory to a new location. 重命名或移动文件/目录.
func (s *SeaweedFSService) Move(ctx context.Context, from, to string) error {
	from = util.NormalizePath(from)
	if strings.HasSuffix(to, "/") {
		base := path.Base(from)
		to = path.Join(to, base)
	}
	to = path.Clean(to)

	q := make(url.Values)
	q.Set("mv.from", from)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.FilerEndpoint+to+"?"+q.Encode(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("move failed: %s %s", resp.Status, b)
	}
	return nil
}

// Copy duplicates a file or directory to a new location. 复制文件或目录到新位置.
func (s *SeaweedFSService) Copy(ctx context.Context, from, to string) error {
	from = util.NormalizePath(from)
	if strings.HasSuffix(to, "/") {
		base := path.Base(from)
		to = path.Join(to, base)
	}
	to = path.Clean(to)

	q := make(url.Values)
	q.Set("cp.from", from)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.FilerEndpoint+to+"?"+q.Encode(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("copy failed: %s %s", resp.Status, b)
	}
	return nil
}

// ListPaged lists a single page of directory entries with optional filters.
// It returns entries, the last file name, and a HasMore flag indicating if more pages exist.
// 列出目录的一页文件, 可使用过滤条件, 返回文件列表、最后一个文件名及是否有更多页.
func (s *SeaweedFSService) ListPaged(
	ctx context.Context,
	dir string,
	lastFileName string,
	limit int,
	namePattern, namePatternExclude string,
	extra map[string]string,
) (ListPagedResult, error) {

	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	q := url.Values{}
	q.Set("format", "json")
	if lastFileName != "" {
		q.Set("lastFileName", lastFileName)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if namePattern != "" {
		q.Set("namePattern", namePattern)
	}
	if namePatternExclude != "" {
		q.Set("namePatternExclude", namePatternExclude)
	}
	for k, v := range extra {
		q.Set(k, v)
	}

	u := s.FilerEndpoint + dir + "?" + q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return ListPagedResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return ListPagedResult{}, fmt.Errorf("list failed: %s %s", resp.Status, b)
	}

	var raw struct {
		Path    string `json:"Path"`
		Entries []struct {
			FullPath string `json:"FullPath"`
			Mtime    string `json:"Mtime"`
			FileSize int64  `json:"FileSize"`
			Mime     string `json:"Mime"`
			Mode     uint32 `json:"Mode"`
		} `json:"Entries"`
		LastFileName string `json:"LastFileName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ListPagedResult{}, err
	}

	var result []SeaweedEntry
	for _, e := range raw.Entries {
		result = append(result, SeaweedEntry{
			Name:  path.Base(strings.TrimRight(e.FullPath, "/")),
			IsDir: e.Mode&uint32(os.ModeDir) != 0,
			Size:  e.FileSize,
			Mime:  e.Mime,
			Mtime: e.Mtime,
		})
	}

	return ListPagedResult{
		Entries: result,
		Last:    raw.LastFileName,
		HasMore: raw.LastFileName != "" && len(result) == limit,
	}, nil
}

// List lists all entries in a directory, automatically paging through results.
// It respects MaxListPages from the safety policy and supports cancellation via context.
// 列出目录下所有文件, 会自动分页, 遵守安全策略的 MaxListPages, 并支持 context 取消.
func (s *SeaweedFSService) List(
	ctx context.Context,
	dir string,
	namePattern, namePatternExclude string,
	extra map[string]string,
) ([]SeaweedEntry, error) {

	var all []SeaweedEntry
	last := ""
	limit := 100
	pageCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		page, err := s.ListPaged(ctx, dir, last, limit, namePattern, namePatternExclude, extra)
		if err != nil {
			return nil, err
		}

		all = append(all, page.Entries...)

		if !page.HasMore {
			break
		}

		if page.Last == last {
			return nil, fmt.Errorf("list aborted: lastFileName not advancing (possible infinite pagination)")
		}

		last = page.Last
		pageCount++

		if pageCount >= s.policy.MaxListPages {
			return nil, fmt.Errorf("list aborted: exceed max pages %d", s.policy.MaxListPages)
		}
	}

	return all, nil
}
