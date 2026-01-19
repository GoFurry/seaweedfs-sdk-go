# SeaweedFS Go SDK

[![Last Version](https://img.shields.io/github/release/GoFurry/seaweedfs-sdk-go/all.svg?logo=github&color=brightgreen)](https://github.com/GoFurry/seaweedfs-sdk-go/releases)
[![License](https://img.shields.io/github/license/GoFurry/seaweedfs-sdk-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.24-blue)](go.mod)

[English README](README.md)

SeaweedFS Go SDK 是一个轻量级客户端库，用于通过 HTTP API 访问 [SeaweedFS](https://github.com/chrislusf/seaweedfs)。  
提供文件操作、目录操作、元数据访问和大文件上传的便捷方法。

---

## 🚀安装

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

## 🧭常用方法

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

## 🌟 使用示例（Gin + curl）

本节展示如何将 SeaweedFS Go SDK 集成到基于 Gin 的 HTTP 服务中。
每个示例都同时包含 **Gin 接口实现代码** 和 **对应的 `curl` 调用方式**，便于理解和快速验证。

---

### 1️⃣ 文件上传（自动选择普通 / 分片）

**Gin 接口代码**

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
    err = fs.UploadFileSmart(c, storage.UploadMethodPut, path, file, 20<<20, 10<<20, opts, headers)
    if err != nil {
        c.JSON(200, gin.H{"error": err})
        return
    }
    c.JSON(200, gin.H{"msg": "upload success")
})
```

**curl 示例**

```bash
curl -X POST "http://localhost:8080/upload?path=/test/hello.txt" \
  -F "file=@hello.txt"
```

**说明**

* 根据文件大小自动选择普通上传或分片上传
* 适用于绝大多数通用上传场景

---

### 2️⃣ 大文件分片上传

**Gin 接口代码**

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
    
    err = fs.UploadFileSmart(c, storage.UploadMethodPut, path, file, 20<<20, 10<<20, opts, headers)
    if err != nil {
        c.JSON(200, gin.H{"error": err})
        return
    }
    c.JSON(200, gin.H{"msg": "large upload success")
})
```

**curl 示例**

```bash
curl -X POST "http://localhost:8080/upload_large?path=/test/big.zip" \
  -F "file=@big.zip"
```

**说明**

* 支持分片上传、失败重试和回退机制
* 适用于大文件或网络不稳定场景

---

### 3️⃣ 文件下载（流式）

**Gin 接口代码**

```go
r.GET("/download", func(c *gin.Context) {
    path := c.Query("path")
    if path == "" {
        c.JSON(400, gin.H{"error": "path required"})
        return
    }
    
    rc, header, err := fs.Download(c, path)
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

**curl 示例**

```bash
curl "http://localhost:8080/download?path=/test/hello.txt" -o hello.txt
```

**说明**

* 全程流式下载，避免占用大量内存
* 自动透传 SeaweedFS 返回的 HTTP Header

---

### 4️⃣ 获取文件元信息（Stat）

**Gin 接口代码**

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

**curl 示例**

```bash
curl "http://localhost:8080/stat?path=/test/hello.txt"
```

**说明**

* 获取文件大小、类型、时间戳、副本策略等信息
* 常用于文件管理、校验和可视化展示

---

> 💡 **提示**
> 在生产环境中，建议将 SDK 的调用封装在自己的 Service 层中，
> 而不是直接在 HTTP Handler 中调用，以提升可维护性和扩展性。


## 📑文档参考
- [SeaweedFS Wiki: Filer-Server-API](https://github.com/seaweedfs/seaweedfs/wiki/Filer-Server-API)

## 🐺许可证
本项目基于 [MIT License](LICENSE) 开源, 允许商业使用、修改、分发, 无需保留原作者版权声明。