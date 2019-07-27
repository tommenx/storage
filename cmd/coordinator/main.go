package main

import (
	"flag"
	apiserver "github.com/tommenx/storage/pkg/api/server"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/server"
	"github.com/tommenx/storage/pkg/store"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	db := store.NewEtcd([]string{"127.0.0.1:2389"})
	path := "/root/.kube/config"
	cli, informerFactory := controller.NewCliAndInformer(path)
	pvController := controller.NewPVController(informerFactory)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	s := server.NewServer(db, pvController)
	go apiserver.StartServer(cli, informerFactory, 8080)
	s.Run()
}
