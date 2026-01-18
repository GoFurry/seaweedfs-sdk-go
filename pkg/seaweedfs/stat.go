// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It includes file and directory metadata operations such as stat, exists, and tags management.
// 提供 SeaweedFS 的 Go 客户端, 包括文件/目录元数据操作, 如 stat、exists 和标签管理.
package seaweedfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/GoFurry/seaweedfs-sdk-go/internal/util"
	"golang.org/x/sync/errgroup"
)

// Stat retrieves metadata of a file or directory.
// includeTags indicates whether to also fetch custom tags for the entry.
// 获取文件或目录的元数据, includeTags 表示是否同时获取自定义标签.
func (s *SeaweedFSService) Stat(ctx context.Context, p string, includeTags bool) (*SeaweedStat, error) {
	if !path.IsAbs(p) {
		p = "/" + p
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.FilerEndpoint+p+"?metadata=true", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stat failed: %s %s", resp.Status, string(body))
	}

	var raw struct {
		FullPath    string  `json:"FullPath"`
		Mtime       string  `json:"Mtime"`
		Crtime      string  `json:"Crtime"`
		Mode        uint32  `json:"Mode"`
		Mime        string  `json:"Mime"`
		Replication string  `json:"Replication"`
		Collection  string  `json:"Collection"`
		TtlSec      int32   `json:"TtlSec"`
		Md5         *string `json:"Md5"`
		FileSize    int64   `json:"FileSize"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	stat := &SeaweedStat{
		Path:        raw.FullPath,
		Name:        path.Base(raw.FullPath),
		IsDir:       raw.Mode&uint32(os.ModeDir) != 0,
		Size:        raw.FileSize,
		Mime:        raw.Mime,
		Md5:         util.DerefString(raw.Md5),
		Mtime:       util.ParseSeaweedTime(raw.Mtime),
		Crtime:      util.ParseSeaweedTime(raw.Crtime),
		Mode:        raw.Mode,
		Replication: raw.Replication,
		Collection:  raw.Collection,
		TtlSec:      raw.TtlSec,
	}

	if includeTags {
		tags, err := s.GetTags(ctx, raw.FullPath)
		if err != nil {
			tags = nil
		}
		stat.Tags = tags
	}

	return stat, nil
}

// StatBatch retrieves metadata for multiple files or directories concurrently.
// concurrency specifies the number of parallel requests.
// ignoreErrors indicates whether to skip errors and continue processing.
// 批量获取文件或目录元数据, concurrency 指定并发数, ignoreErrors 表示是否忽略错误.
func (s *SeaweedFSService) StatBatch(ctx context.Context, paths []string, concurrency int, ignoreErrors bool, includeTags bool) (map[string]*SeaweedStat, error) {
	if concurrency <= 0 {
		concurrency = 10
	}

	result := make(map[string]*SeaweedStat, len(paths))
	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, concurrency)

	for _, p := range paths {
		p := p
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				return nil
			default:
			}

			stat, err := s.Stat(ctx, p, includeTags)
			if err != nil && !ignoreErrors {
				return err
			}

			mu.Lock()
			if err != nil && ignoreErrors {
				result[p] = nil
			} else {
				result[p] = stat
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// Exists checks whether a file or directory exists. 检查文件或目录是否存在.
func (s *SeaweedFSService) Exists(ctx context.Context, p string) (bool, error) {
	_, err := s.Stat(ctx, p, false)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsBatch checks existence of multiple files or directories concurrently.
// concurrency specifies the number of parallel requests.
// ignoreErrors indicates whether to treat errors as non-existent.
// 批量检查文件或目录是否存在, concurrency 指定并发数, ignoreErrors 表示是否忽略错误.
func (s *SeaweedFSService) ExistsBatch(ctx context.Context, paths []string, concurrency int, ignoreErrors bool) (map[string]bool, error) {
	if concurrency <= 0 {
		concurrency = 10
	}

	result := make(map[string]bool, len(paths))
	mu := sync.Mutex{}
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, concurrency)

	for _, p := range paths {
		p := p
		select {
		case <-ctx.Done():
			mu.Lock()
			result[p] = false
			mu.Unlock()
			continue
		default:
		}

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			select {
			case <-ctx.Done():
				mu.Lock()
				result[p] = false
				mu.Unlock()
				return nil
			default:
			}

			exists, err := s.Exists(ctx, p)
			if err != nil {
				if ignoreErrors {
					mu.Lock()
					result[p] = false
					mu.Unlock()
					return nil
				}
				return err
			}

			mu.Lock()
			result[p] = exists
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// SetTags sets custom tags on a file or directory. 为文件或目录设置自定义标签.
func (s *SeaweedFSService) SetTags(ctx context.Context, path string, tags FileTags) error {
	path = util.NormalizePath(path)
	if len(tags) == 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.FilerEndpoint+path+"?tagging", nil)
	if err != nil {
		return err
	}

	for k, v := range tags {
		req.Header.Set("Seaweed-"+k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("set tags failed: %s %s", resp.Status, body)
}

// GetTags retrieves custom tags of a file or directory. 获取文件或目录的自定义标签.
func (s *SeaweedFSService) GetTags(ctx context.Context, path string) (FileTags, error) {
	path = util.NormalizePath(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, s.FilerEndpoint+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get tags failed: %s %s", resp.Status, body)
	}

	tags := make(FileTags)
	for k, vals := range resp.Header {
		if strings.HasPrefix(k, "Seaweed-") && len(vals) > 0 {
			tags[strings.TrimPrefix(k, "Seaweed-")] = vals[0]
		}
	}
	return tags, nil
}

// DeleteTags deletes custom tags of a file or directory.
// If keys is empty, all tags with "Seaweed-" prefix are removed.
// 删除文件或目录的自定义标签, 如果 keys 为空则删除所有 Seaweed- 前缀标签.
func (s *SeaweedFSService) DeleteTags(ctx context.Context, path string, keys ...string) error {
	path = util.NormalizePath(path)
	u := s.FilerEndpoint + path + "?tagging"
	if len(keys) > 0 {
		u += "=" + strings.Join(keys, ",")
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
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete tags failed: %s %s", resp.Status, body)
}
