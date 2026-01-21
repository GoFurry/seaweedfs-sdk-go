# SeaweedFS Go SDK – Examples

This directory contains **standalone examples** demonstrating how to use the SeaweedFS Go SDK.
Each example is a **minimal, self-contained Go program** that focuses on a single capability.

本目录包含一组 **独立示例程序**，用于演示如何使用 SeaweedFS Go SDK。
每个示例都是 **最小可运行的 Go 程序**，专注于某一个具体能力。

---

## Directory Structure / 目录结构说明

```
example/
│  init.go
│  README.md
│
└─ops
    ├─delete
    │      delete.go
    │
    ├─download
    │  ├─download
    │  │      download.go
    │  │
    │  └─download_range
    │          download_range.go
    │
    ├─fsops
    │  ├─mkdir
    │  │      mkdir.go
    │  │
    │  └─move_copy
    │          move_copy.go
    │
    ├─stat
    │  ├─list
    │  │      list.go
    │  │
    │  └─storage
    │          dir_storage.go
    │
    └─upload
            upload.go
```

---

## File Descriptions / 文件说明

### `init.go`

**Global initialization example**

- Initializes `SeaweedFSService`
- Creates shared `context.Context`
- Used as a reference for SDK setup

**全局初始化示例**

- 初始化 `SeaweedFSService`
- 创建共享的 `context.Context`
- 作为 SDK 初始化的参考模板

---

## ops – Operation Examples / 操作示例

Each subdirectory under `ops` represents **one category of filesystem operations**.

`ops` 目录下的每个子目录，表示一类 **文件系统操作能力**。

---

### `ops/upload/upload.go`

**File upload example** ⭐

- Demonstrates smart upload
- Automatically switches between normal and large-file upload
- Suitable for local files or backend services

**文件上传示例** ⭐

- 演示智能上传逻辑
- 自动选择普通上传或分片上传
- 适合本地文件或后台任务

---

### `ops/download/download/download.go`

**Full file download example** ⭐

- Downloads a remote file to local disk
- Uses streaming to avoid large memory usage

**完整文件下载示例** ⭐

- 将远端文件下载到本地
- 使用流式读取，避免占用大量内存

---

### `ops/download/download_range/download_range.go`

**Range download example** ⭐

- Downloads part of a file using byte range
- Useful for resumable downloads

**范围下载示例** ⭐

- 使用字节区间下载文件片段
- 适用于断点续传场景

---

### `ops/fsops/mkdir/mkdir.go`

**Directory creation example** ⭐

- Creates a directory in SeaweedFS
- Demonstrates filer directory operations

**目录创建示例** ⭐

- 在 SeaweedFS 中创建目录
- 演示 Filer 的目录操作能力

---

### `ops/fsops/move_copy/move_copy.go`

**Move & copy example** ⭐

- Demonstrates file move
- Demonstrates file copy

**移动与复制示例** ⭐

- 演示文件移动
- 演示文件复制

---

### `ops/delete/delete.go`

**Delete example** ⭐

- Deletes a single file
- Can be extended for batch deletion

**删除示例** ⭐

- 删除单个文件
- 可扩展为批量删除

---

### `ops/stat/list/list.go`

**Directory listing example** ⭐

- Lists files in a directory
- Supports pagination internally

**目录列表示例** ⭐

- 列出目录下的文件
- 内部支持分页逻辑

---

### `ops/stat/storage/dir_storage.go`

**Directory storage usage statistics** ⭐

- Recursively calculates total storage size of a directory
- Suitable for user quota / billing systems

**目录存储量统计示例** ⭐

- 递归统计目录占用的总存储空间
- 适合用户配额、计费、监控等场景

---

## Design Philosophy / 设计理念

- Each example is **framework-independent**
- No Gin / Fiber / HTTP framework dependencies
- Suitable for direct copy into CLI tools or services

**每个示例都遵循以下原则：**

- 与 Web 框架无关
- 不依赖 Gin / Fiber 等 HTTP 框架
- 可直接复制到 CLI 或后台服务中使用

---
