package main

import (
	"io"
	"os"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	rc, _, _, err := example.Fs.DownloadRange(example.Ctx, "/test/test.jpg", 0, 1024*1024)
	example.Must(err)
	defer rc.Close()

	out, err := os.Create("part.bin")
	example.Must(err)
	defer out.Close()

	_, err = io.Copy(out, rc)
	example.Must(err)
}
