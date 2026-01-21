package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {

	example.Must(
		example.Fs.Delete(
			// Context for timeout cancel 用于控制取消和超时
			example.Ctx,
			// File path to delete 删除的文件路径
			"/test/test.jpg",
			// Optional for SeaweedFS param (e.g., recursive, skipChunkDeletion)
			// 可选参数, 删除支持的相关参数
			nil,
		),
	)

	// List of file paths to delete 需要删除的文件路径列表
	paths := []string{
		"/test/a.jpg",
		"/test/b.jpeg",
		"/test/c.jpg",
	}

	result := example.Fs.DeleteBatch(
		example.Ctx,
		paths,
		nil,
		// ignore error msg 不记录错误信息
		false,
		// concurrency 并发数
		2,
	)

	// Iterate over results and print
	// 遍历结果并打印每个文件的删除情况
	for p, err := range result {
		fmt.Println(p, err)
	}
}
