package example

import (
	"context"
	"log"
	"time"

	"github.com/GoFurry/seaweedfs-sdk-go/pkg/seaweedfs"
)

var (
	// SeaweedFS endpoint 服务地址
	Endpoint = "http://127.0.0.1:8888"

	// SeaweedFS client instance 客户端实例
	Fs = seaweedfs.NewSeaweedFSService(Endpoint)

	// Context with 5-minute timeout for API calls
	// 带 5 分钟超时的 Context, 用于控制 API 请求
	Ctx, _ = context.WithTimeout(context.Background(), 5*time.Minute)
)

// Must checks for error and terminates the program if not nil
// Must 用于检查错误, 如果错误不为空则终止程序
func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
