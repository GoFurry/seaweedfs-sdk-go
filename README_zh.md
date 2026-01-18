# SeaweedFS Go SDK

[![Last Version](https://img.shields.io/github/release/GoFurry/seaweedfs-sdk-go/all.svg?logo=github&color=brightgreen)](https://github.com/GoFurry/seaweedfs-sdk-go/releases)
[![License](https://img.shields.io/github/license/GoFurry/seaweedfs-sdk-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](go.mod)

[English README](README.md)

SeaweedFS Go SDK 是一个轻量级客户端库，用于通过 HTTP API 访问 [SeaweedFS](https://github.com/chrislusf/seaweedfs)。  
提供文件操作、目录操作、元数据访问和大文件上传的便捷方法。

---

## 安装

```bash
go get github.com/GoFurry/seaweedfs-sdk-go
```

---

## 目录结构

```
├─ internal
│  ├─ policy       # 安全策略与重试策略
│  └─ util         # 内部工具函数
└─ pkg
    └─ seaweedfs
        ├─ client.go      # SeaweedFSService 客户端和配置
        ├─ download.go    # 文件下载函数
        ├─ fsops.go       # 文件系统操作（创建、删除、移动、复制、列出）
        ├─ stat.go        # 文件/目录元数据操作
        ├─ types.go       # 公共类型和结构体
        ├─ upload.go      # 文件上传函数
        └─ util.go        # 公共工具函数
```

---

## 核心类型

### `SeaweedFSService`
连接 SeaweedFS filer 的客户端结构体。

```go
service := seaweedfs.NewSeaweedFSService("http://localhost:8888")
```

支持函数式选项进行配置：

- `WithSafetyPolicy(policy.SafetyPolicy)`
- `WithMaxDownloadChunks(int)`
- `WithMaxListPages(int)`
- `WithUploadMaxRetry(int)`
- `WithBackoff(base, max time.Duration)`

---

### `SeaweedStat`
表示文件或目录的元数据：

```go
type SeaweedStat struct {
    Path        string
    Name        string
    IsDir       bool
    Size        int64
    Mime        string
    Md5         string
    Mtime       time.Time
    Crtime      time.Time
    Mode        uint32
    Replication string
    Collection  string
    TtlSec      int32
    Tags        FileTags
}
```

---

### `SeaweedEntry`
目录下的条目：

```go
type SeaweedEntry struct {
    Name  string
    IsDir bool
    Size  int64
    Mime  string
    Mtime string
}
```

---

### `FileTags`
文件或目录的自定义标签：

```go
type FileTags map[string]string
```

---

## 常用方法

### 文件上传

```go
service.UploadWithOptions(ctx, seaweedfs.UploadMethodPut, "/path/to/file.txt", reader, opts, headers)
service.UploadLarge(ctx, seaweedfs.UploadMethodPut, "/bigfile.zip", reader, size, chunkSize, opts, headers, largeOptions)
service.UploadFileSmart(ctx, seaweedfs.UploadMethodPut, "/file.txt", fileHeader, 10*1024*1024, chunkSize, opts, headers)
```

### 文件下载

```go
rc, header, err := service.Download(ctx, "/path/to/file.txt")
rc, header, status, err := service.DownloadRange(ctx, "/path/to/file.txt", 0, 1024)
chunks := service.DownloadConcurrent(ctx, "/bigfile.zip", "/tmp/bigfile.zip", 4)
```

### 文件系统操作

```go
service.Mkdir(ctx, "/folder/")
service.Delete(ctx, "/file.txt", nil)
service.DeleteBatch(ctx, []string{"/a", "/b"}, nil, true, 4)
service.Move(ctx, "/a.txt", "/b.txt")
service.Copy(ctx, "/a.txt", "/copy.txt")
entries := service.List(ctx, "/folder/", "", "", nil)
```

### 元数据操作

```go
stat, err := service.Stat(ctx, "/file.txt", true)
batchStats, err := service.StatBatch(ctx, []string{"/a", "/b"}, 5, true, true)
exists, err := service.Exists(ctx, "/file.txt")
tags := service.GetTags(ctx, "/file.txt")
service.SetTags(ctx, "/file.txt", FileTags{"tag1":"value1"})
service.DeleteTags(ctx, "/file.txt", "tag1")
```

---

## 工具函数

```go
size, err := seaweedfs.LocalFileSize("/tmp/file.txt")
t, err := seaweedfs.ParseSeaweedTime("2026-01-18T00:00:00Z")
```
