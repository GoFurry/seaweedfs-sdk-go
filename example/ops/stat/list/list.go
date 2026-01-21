package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	// List files and directories under a path
	// 列出指定路径下的文件和目录
	entries, err := example.Fs.List(
		// Context for timeout/cancel control 用于控制超时和取消
		example.Ctx,
		// Directory path to list 要列出的目录路径
		"/test/",
		// Name filter (include) 文件名过滤(包含)
		"",
		// Name filter (exclude) 文件名过滤(不包含)
		"",
		// Optional query parameters
		// 可选查询参数
		nil,
	)
	example.Must(err)

	// Iterate over results and print each entry
	// 遍历结果并打印每个文件或目录信息
	for _, e := range entries {
		fmt.Printf(
			"%s  dir=%v  size=%d\n",
			e.Name,  // Entry name 条目名称
			e.IsDir, // Whether entry is a directory 是否为目录
			e.Size,  // Size of the file 文件大小
		)
	}
}
