package main

import (
	"flag"
	"github.com/tommenx/storage/pkg/api/server"
	"github.com/tommenx/storage/pkg/controller"
)

func init() {
	flag.Set("logtostderr", "true")

}

func main() {
	flag.Parse()
	path := "/root/.kube/config"
	cli, informer := controller.NewCliAndInformer(path)
	server.StartServer(cli, informer, 8080)
}
