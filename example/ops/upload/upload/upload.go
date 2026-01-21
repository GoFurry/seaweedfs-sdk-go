package main

import (
	"github.com/GoFurry/seaweedfs-sdk-go/example"
	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

func main() {
	err := example.Fs.UploadLocalFile(
		example.Ctx,
		seaweedfs.UploadMethodPut,
		"/user/100/postgres-17.6.tar",
		"E:/test/image/postgres-17.6.tar",
		20<<20, // 20MB threshold
		10<<20, // 10MB chunk
		nil,
		nil,
		nil,
	)
	example.Must(err)
}
