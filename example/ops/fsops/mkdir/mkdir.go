package main

import "github.com/GoFurry/seaweedfs-sdk-go/example"

func main() {
	// Create a new directory in SeaweedFS
	// 在 SeaweedFS 中创建一个新目录
	example.Must(
		example.Fs.Mkdir(
			// Context for timeout/cancel control
			// 用于控制超时和取消
			example.Ctx,
			// Directory path to create
			// 要创建的目录路径
			"/user/100",
		),
	)
}
