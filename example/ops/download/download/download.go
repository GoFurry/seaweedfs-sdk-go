package main

import (
	"io"
	"os"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	// Download a file from SeaweedFS
	// 从 SeaweedFS 下载文件
	rc, _, err := example.Fs.Download(
		// Context for timeout/cancel 用于控制超时和取消
		example.Ctx,
		// File path to download 下载的文件路径
		"/test/test.jpg",
		// Progress callback 下载进度回调
		nil,
	)
	example.Must(err)
	defer rc.Close()

	// Create local file to save the downloaded content
	// 创建本地文件保存下载内容
	out, err := os.Create("downloaded.jpg")
	example.Must(err)
	defer out.Close()

	// Copy content from SeaweedFS response to local file
	// 将下载的文件内容写入本地文件
	_, err = io.Copy(out, rc)
	example.Must(err)
}
