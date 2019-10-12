package main

import (
	"flag"
	apiserver "github.com/tommenx/storage/pkg/api/server"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/server"
	"github.com/tommenx/storage/pkg/store"
)

var (
	path string
	etcd string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&path, "config", "/root/.kube/config", "use to set config file path")
	flag.StringVar(&etcd, "etcd", "127.0.0.1:2389", "coordinator url")
}

func main() {
	flag.Parse()
	db := store.NewEtcd([]string{etcd})
	cli, informerFactory := controller.NewCliAndInformer(path)
	pvController := controller.NewPVController(informerFactory)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	s := server.NewServer(db, pvController)
	go apiserver.StartServer(cli, informerFactory, 8888)
	s.Run()
}
