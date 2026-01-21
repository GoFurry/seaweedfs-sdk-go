package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

func main() {
	// Define upload progress callback
	// 定义上传进度回调
	uploadProgress := func(done int64, total int64) {
		percent := float64(done) / float64(total) * 100
		fmt.Printf("\r上传进度: %.2f%% (%d/%d bytes)", percent, done, total)
	}

	// Local file path to upload
	// 要上传的本地文件路径
	localFile := "E:/test/image/big_file.tar"

	// Upload local file with progress callback
	// 上传本地文件, 支持上传进度回调
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
		localFile,
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
		uploadProgress,
	)
	example.Must(err)

	// Upload completed 上传完成
	fmt.Println("\nupload progress completed!")
}
