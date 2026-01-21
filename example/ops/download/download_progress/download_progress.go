package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	// Remote file path on SeaweedFS 文件路径
	downloadPath := "/user/100/big_file.tar"
	// Local destination path to save downloaded file 本地保存路径
	dstPath := "E:/test/big_file.tar"

	// Define download progress callback
	// 定义下载进度回调函数
	downloadProgress := func(done int64, total int64) {
		// Calculate percentage completed 计算下载百分比
		percent := float64(done) / float64(total) * 100
		// Print progress in place 在同一行打印进度
		fmt.Printf("\r下载进度: %.2f%% (%d/%d bytes)", percent, done, total)
	}

	fmt.Println("开始下载...")
	// Download file concurrently with 4 chunks and progress callback
	// 并发分块下载文件, 并传入下载进度回调
	err := example.Fs.DownloadConcurrent(
		example.Ctx,      // Context for timeout/cancel 用于控制超时和取消
		downloadPath,     // Remote file path on SeaweedFS
		dstPath,          // Local destination path
		4,                // Number of concurrent chunks 并发分块数
		downloadProgress, // Progress callback function 下载进度回调函数
	)
	// Check for download errors
	// 检查下载过程中是否有错误
	if err != nil {
		flag := false
		for _, e := range err {
			if e != nil {
				flag = true
				fmt.Println("\ndownload fail:", err)
			}
		}
		if flag {
			return
		}
	}
	fmt.Println("\ndownload complete!")
}
