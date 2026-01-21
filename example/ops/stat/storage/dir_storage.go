package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	// Get directory usage statistics 获取指定目录的使用情况统计
	usage, err := example.Fs.GetDirUsage(
		// Context for timeout/cancel control 用于控制超时和取消
		example.Ctx,
		// Directory path to query 要查询的目录路径
		"/user",
	)
	example.Must(err)

	// Print usage information
	// 打印目录使用信息
	fmt.Printf(
		"used=%d MB, files=%d, dirs=%d\n",
		// Total size in MB 总使用空间 (MB)
		usage.TotalSize/1024/1024,
		// Number of files 文件数量
		usage.FileCount,
		// Number of directories 目录数量
		usage.DirCount,
	)
}
