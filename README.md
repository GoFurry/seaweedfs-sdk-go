# SeaweedFS SDK for Go

[![Last Version](https://img.shields.io/github/release/GoFurry/seaweedfs-sdk-go/all.svg?logo=github&color=brightgreen)](https://github.com/GoFurry/seaweedfs-sdk-go/releases)
[![License](https://img.shields.io/github/license/GoFurry/seaweedfs-sdk-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](go.mod)

[中文文档](README_zh.md)

SeaweedFS SDK for Go is a lightweight client library for interacting with [SeaweedFS](https://github.com/chrislusf/seaweedfs) via its HTTP API.  
It provides convenient methods for file operations, directory operations, metadata access, and large file uploads.

---

## Installation

```bash
go get github.com/GoFurry/seaweedfs-sdk-go
```

---

## Package Structure

```
├─ internal
│  ├─ policy       # Safety and retry policies
│  └─ util         # Internal utilities
└─ pkg
    └─ seaweedfs
        ├─ client.go      # SeaweedFSService client and configuration
        ├─ download.go    # File download functions
        ├─ fsops.go       # File system operations (mkdir, delete, move, copy, list)
        ├─ stat.go        # File/directory metadata operations
        ├─ types.go       # Common types and structs
        ├─ upload.go      # File upload functions
        └─ util.go        # Helper utilities for public package
```

---

## Key Types

### `SeaweedFSService`
Main client structure for connecting to a SeaweedFS filer.

```go
service := seaweedfs.NewSeaweedFSService("http://localhost:8888")
```

Supports optional configuration via functional options:

- `WithSafetyPolicy(policy.SafetyPolicy)`
- `WithMaxDownloadChunks(int)`
- `WithMaxListPages(int)`
- `WithUploadMaxRetry(int)`
- `WithBackoff(base, max time.Duration)`

---

### `SeaweedStat`
Represents file or directory metadata:

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
Directory entry returned when listing folders:

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
Custom tags attached to files or directories:

```go
type FileTags map[string]string
```

---

## Common Methods

### File Upload

```go
service.UploadWithOptions(ctx, seaweedfs.UploadMethodPut, "/path/to/file.txt", reader, opts, headers)
service.UploadLarge(ctx, seaweedfs.UploadMethodPut, "/bigfile.zip", reader, size, chunkSize, opts, headers, largeOptions)
service.UploadFileSmart(ctx, seaweedfs.UploadMethodPut, "/file.txt", fileHeader, 10*1024*1024, chunkSize, opts, headers)
```

### File Download

```go
rc, header, err := service.Download(ctx, "/path/to/file.txt")
rc, header, status, err := service.DownloadRange(ctx, "/path/to/file.txt", 0, 1024)
chunks := service.DownloadConcurrent(ctx, "/bigfile.zip", "/tmp/bigfile.zip", 4)
```

### File System Operations

```go
service.Mkdir(ctx, "/folder/")
service.Delete(ctx, "/file.txt", nil)
service.DeleteBatch(ctx, []string{"/a", "/b"}, nil, true, 4)
service.Move(ctx, "/a.txt", "/b.txt")
service.Copy(ctx, "/a.txt", "/copy.txt")
entries := service.List(ctx, "/folder/", "", "", nil)
```

### Metadata Operations

```go
stat, err := service.Stat(ctx, "/file.txt", true)
batchStats, err := service.StatBatch(ctx, []string{"/a", "/b"}, 5, true, true)
exists, err := service.Exists(ctx, "/file.txt")
tags := service.GetTags(ctx, "/file.txt")
service.SetTags(ctx, "/file.txt", FileTags{"tag1":"value1"})
service.DeleteTags(ctx, "/file.txt", "tag1")
```

---

## Utilities

```go
size, err := seaweedfs.LocalFileSize("/tmp/file.txt")
t, err := seaweedfs.ParseSeaweedTime("2026-01-18T00:00:00Z")
```
