package main

import (
	"io"
	"os"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	// Download first 1MB (0~1*1024*1024 bytes) from remote file
	// 从远程文件下载前 1MB (字节范围 0~1024*1024)
	rc, _, _, err := example.Fs.DownloadRange(
		example.Ctx,      // Context for timeout/cancel 用于控制超时和取消
		"/test/test.jpg", // Remote file path on SeaweedFS 文件路径
		0,                // Start byte offset 起始字节偏移
		1024*1024,        // End byte offset (inclusive) 结束字节偏移
		nil,              // Progress callback function 下载进度回调函数
	)
	// Check for download error 检查下载错误
	example.Must(err)
	defer rc.Close()

	// Create local file to save downloaded part 创建本地文件用于保存下载的分片
	out, err := os.Create("part.bin")
	example.Must(err)
	defer out.Close()

	// Copy downloaded content to local file 将下载的数据写入本地文件
	_, err = io.Copy(out, rc)
	example.Must(err)
}
