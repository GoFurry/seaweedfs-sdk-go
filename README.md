# SeaweedFS SDK for Go

[![Last Version](https://img.shields.io/github/release/GoFurry/seaweedfs-sdk-go/all.svg?logo=github&color=brightgreen)](https://github.com/GoFurry/seaweedfs-sdk-go/releases)
[![License](https://img.shields.io/github/license/GoFurry/seaweedfs-sdk-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](go.mod)

[中文文档](README_zh.md)

SeaweedFS SDK for Go is a lightweight client library for interacting with [SeaweedFS](https://github.com/chrislusf/seaweedfs) via its HTTP API.  
It provides convenient methods for file operations, directory operations, metadata access, and large file uploads.

---

## 🚀Installation

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

## 🧭Common Methods

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

## 🌟 Usage Examples (Gin + curl)

This section demonstrates how to integrate the SeaweedFS Go SDK into a Gin-based HTTP service.
Each example includes both the Gin handler implementation and the corresponding `curl` command.

---

### 1️⃣ Upload File (Auto Small / Large)

**Gin handler**

```go
r.POST("/upload", func(c *gin.Context) {
    path := c.Query("path")
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    opts := map[string]string{
        //"ttl": "1d",
        //"op":  "append",
    }
    
    headers := map[string]string{
        //"Seaweed-Tag": "avatar",
    }
    
    // 统一小文件上传
    err = fs.UploadFileSmart(c, storage.UploadMethodPut, path, file, 20<<20, 10<<20, opts, headers, nil)
    if err != nil {
        c.JSON(200, gin.H{"error": err})
        return
    }
    c.JSON(200, gin.H{"msg": "upload success")
})
```

**curl**

```bash
curl -X POST "http://localhost:8080/upload?path=/test/hello.txt" \
  -F "file=@hello.txt"
```

**Description**

* Automatically selects PUT or chunked upload based on file size
* Suitable for most general upload scenarios

---

### 2️⃣ Large File Upload (Chunked)

**Gin handler**

```go
r.POST("/upload_large", func(c *gin.Context) {
    path := c.Query("path")
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    opts := map[string]string{
        
    }
    
    headers := map[string]string{}
	
    err = fs.UploadFileSmart(c, storage.UploadMethodPut, path, file, 20<<20, 10<<20, opts, headers, nil)
    if err != nil {
        c.JSON(200, gin.H{"error": err})
		return
    }
    c.JSON(200, gin.H{"msg": "large upload success")
})
```

**curl**

```bash
curl -X POST "http://localhost:8080/upload_large?path=/test/big.zip" \
  -F "file=@big.zip"
```

**Description**

* Chunked upload with retry support
* Suitable for large files or unstable networks

---

### 3️⃣ Download File (Streaming)

**Gin handler**

```go
r.GET("/download", func(c *gin.Context) {
    path := c.Query("path")
    if path == "" {
        c.JSON(400, gin.H{"error": "path required"})
        return
    }
    
    rc, header, err := fs.Download(c, path, nil)
    if err != nil {
        c.JSON(404, gin.H{"error": err.Error()})
        return
    }
    defer rc.Close()
    
    for k, v := range header {
        if len(v) > 0 {
            c.Header(k, v[0])
        }
    }
    
    c.Status(200)
    _, _ = io.Copy(c.Writer, rc)
})
```

**curl**

```bash
curl "http://localhost:8080/download?path=/test/hello.txt" -o hello.txt
```

**Description**

* Fully streaming download
* Headers are transparently forwarded to the client

---

### 4️⃣ Get File Metadata (Stat)

**Gin handler**

```go
r.GET("/stat", func(c *gin.Context) {
    path := c.Query("path")
    stat, err := fs.Stat(context.Background(), path, false)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, stat)
})
```

**curl**

```bash
curl "http://localhost:8080/stat?path=/test/hello.txt"
```

**Description**

* Retrieves file metadata including size, mime type, timestamps, and replication info
* Useful for file management and validation

---

> 💡 **Tip**
> In production environments, it is recommended to wrap the SDK calls inside your own service layer
> rather than exposing them directly in HTTP handlers.


## 📑Documentation References
- [SeaweedFS Wiki: Filer-Server-API](https://github.com/seaweedfs/seaweedfs/wiki/Filer-Server-API)

## 🐺License
This project is open-sourced under the [MIT License](LICENSE), which permits commercial use, modification, and distribution without requiring the original author's copyright notice to be retained.
