package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	usage, err := example.Fs.GetDirUsage(example.Ctx, "/user")
	example.Must(err)

	fmt.Printf(
		"used=%d MB, files=%d, dirs=%d\n",
		usage.TotalSize/1024/1024,
		usage.FileCount,
		usage.DirCount,
	)
}
