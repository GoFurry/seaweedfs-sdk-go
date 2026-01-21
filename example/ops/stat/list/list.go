package main

import (
	"fmt"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	entries, err := example.Fs.List(example.Ctx, "/test/", "", "", nil)
	example.Must(err)

	for _, e := range entries {
		fmt.Printf(
			"%s  dir=%v  size=%d\n",
			e.Name,
			e.IsDir,
			e.Size,
		)
	}
}
