package main

import "github.com/GoFurry/seaweedfs-sdk-go/example"

func main() {
	// Copy a file from source to destination directory
	// 将文件从源路径复制到目标目录
	example.Must(
		example.Fs.Copy(
			// Context for timeout/cancel control 用于控制超时和取消
			example.Ctx,
			// Source file path 源文件路径
			"/user/100/big_file.tar",
			// Destination directory path 目标目录路径
			"/user/100/backup/",
		),
	)

	// Move a file from source to destination path
	// 将文件从源路径移动到目标路径
	example.Must(
		example.Fs.Move(
			// Context for timeout/cancel control 用于控制超时和取消
			example.Ctx,
			// Source file path 源文件路径
			"/user/100/big_file.tar",
			// Destination file path 目标文件路径
			"/user/100/archive/big_file.tar",
		),
	)
}
