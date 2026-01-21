package main

import (
	"github.com/GoFurry/seaweedfs-sdk-go/example"
	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

func main() {
	// Output path for the merged file 合并后的目标文件路径
	outputPath := "E:/test/big_file.tar"

	// Merge multiple downloaded parts into a single file
	// 将多个下载的分片合并成一个完整文件
	err := seaweedfs.MergeFiles(
		outputPath, // Destination file path 目标文件路径
		// List of part file paths in order 按顺序的分片文件路径列表
		[]string{
			"E:/test/big_file.tar.part0",
			"E:/test/big_file.tar.part1",
			"E:/test/big_file.tar.part2",
			"E:/test/big_file.tar.part3",
		},
		// Cleanup: delete part files after merge
		// 合并后是否删除源分片文件
		true,
	)
	// Check for errors during merge
	// 检查合并过程中的错误
	example.Must(err)
}
