// Package seaweedfs provides a Go client for interacting with SeaweedFS.
// It defines common types used across SeaweedFS operations.
// 提供 SeaweedFS 的 Go 客户端, 并定义了 SDK 各操作使用的通用类型.
package seaweedfs

import "time"

// FileTags represents custom tags for a file or directory.
// 表示文件或目录的自定义标签, key-value 形式.
type FileTags map[string]string

// SeaweedStat represents the metadata of a file or directory in SeaweedFS.
// 表示 SeaweedFS 中文件或目录的元数据.
type SeaweedStat struct {
	Path        string    `json:"path"`                  // Full path of the entry / 文件或目录的完整路径
	Name        string    `json:"name"`                  // Base name / 基础名称
	IsDir       bool      `json:"isDir"`                 // Whether it's a directory / 是否为目录
	Size        int64     `json:"size"`                  // File size in bytes / 文件大小 (字节)
	Mime        string    `json:"mime"`                  // MIME type / 文件类型
	Md5         string    `json:"md5,omitempty"`         // Optional MD5 checksum / 可选 MD5 校验值
	Mtime       time.Time `json:"mtime"`                 // Last modification time / 最后修改时间
	Crtime      time.Time `json:"crtime"`                // Creation time / 创建时间
	Mode        uint32    `json:"mode"`                  // File mode / 文件模式
	Replication string    `json:"replication,omitempty"` // Optional replication info / 可选副本信息
	Collection  string    `json:"collection,omitempty"`  // Optional collection / 可选集合
	TtlSec      int32     `json:"ttlSec,omitempty"`      // Optional TTL in seconds / 可选生存时间 (秒)
	Tags        FileTags  `json:"tags"`                  // Custom tags / 自定义标签
}

// SeaweedEntry represents a file or directory entry when listing directories.
// 表示目录列表中的文件或目录条目.
type SeaweedEntry struct {
	Name  string `json:"name"`  // Base name / 基础名称
	IsDir bool   `json:"isDir"` // Whether it's a directory / 是否为目录
	Size  int64  `json:"size"`  // File size in bytes / 文件大小 (字节)
	Mime  string `json:"mime"`  // MIME type / 文件类型
	Mtime string `json:"mtime"` // Modification time as string / 修改时间 (字符串)
}

// ListPagedResult represents a single page result of a directory listing.
// 表示目录分页列表结果.
type ListPagedResult struct {
	Entries []SeaweedEntry // Entries in this page / 当前页的条目
	Last    string         // Name of the last entry / 本页最后一个条目的名称
	HasMore bool           // Whether there are more pages / 是否还有更多分页
}
