package main

import (
	"fmt"
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/watcher"
	"time"
)

func main() {
	rpc.Init(":50051")
	config.Init("../../config.toml")
	watch := watcher.NewWatcher(time.Second * 5)
	stopCh := make(chan struct{})
	go watch.Run(stopCh)
	time.Sleep(100 * time.Second)
	stopCh <- struct{}{}
	fmt.Println("complete")
}
