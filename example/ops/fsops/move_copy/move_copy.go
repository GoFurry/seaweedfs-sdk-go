package main

import "github.com/GoFurry/seaweedfs-sdk-go/example"

func main() {
	example.Must(example.Fs.Copy(example.Ctx, "/user/100/postgres-17.6.tar", "/user/100/backup/"))
	example.Must(example.Fs.Move(example.Ctx, "/user/100/postgres-17.6.tar", "/user/100/archive/postgres-17.6.tar"))
}
