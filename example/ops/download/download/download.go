package main

import (
	"io"
	"os"

	"github.com/GoFurry/seaweedfs-sdk-go/example"
)

func main() {
	rc, _, err := example.Fs.Download(example.Ctx, "/test/test.jpg")
	example.Must(err)
	defer rc.Close()

	out, err := os.Create("downloaded.jpg")
	example.Must(err)
	defer out.Close()

	_, err = io.Copy(out, rc)
	example.Must(err)
}
