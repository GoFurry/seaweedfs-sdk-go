package example

import (
	"context"
	"log"
	"time"

	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

var (
	Endpoint = "http://192.168.153.121:44488"
	Fs       = seaweedfs.NewSeaweedFSService(Endpoint)
	Ctx, _   = context.WithTimeout(context.Background(), 5*time.Minute)
)

func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
