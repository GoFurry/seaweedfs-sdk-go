package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	example.Must(example.Fs.Delete(example.Ctx, "/test/test.jpg", nil))

	paths := []string{
		"/test/a.jpg",
		"/test/mmexport1757515637886.jpeg",
		"/test/c.jpg",
	}

	result := example.Fs.DeleteBatch(example.Ctx, paths, nil, false, 2)
	for p, err := range result {
		fmt.Println(p, err)
	}
}
