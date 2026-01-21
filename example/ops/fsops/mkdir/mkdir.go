package main

import "github.com/GoFurry/seaweedfs-sdk-go/example"

func main() {
	example.Must(example.Fs.Mkdir(example.Ctx, "/user/100"))
}
