package main

import (
	"github.com/GoFurry/seaweedfs-sdk-go/example"
	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

func main() {
	// Upload a local file to SeaweedFS
	// 从本地文件上传到 SeaweedFS
	err := example.Fs.UploadLocalFile(
		// Context for timeout/cancel control
		// 用于控制超时和取消
		example.Ctx,
		// HTTP method to use: PUT/POST
		// 上传方法: PUT 或 POST
		seaweedfs.UploadMethodPut,
		// Destination path in SeaweedFS 目标路径
		"/user/100/big_file.tar",
		// Local file path 本地文件路径
		"E:/test/image/big_file.tar",
		// Threshold for large file upload (files larger than this use chunked upload)
		// 大文件阈值, 超过此大小使用分片上传
		20<<20, // 20MB
		// Chunk size for large file upload
		// 分片上传时每片大小
		10<<20, // 10MB
		// Optional query parameters
		// 可选参数, 如 op=append 等
		nil,
		// Optional HTTP headers
		// 可选 HTTP 头, 如 Content-Type
		nil,
		// Optional progress callback
		// 可选上传进度回调, func(done int64, total int64)
		nil,
	)
	example.Must(err)
}
