package main

import (
	"flag"
	"github.com/tommenx/storage/pkg/api"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/server"
	"github.com/tommenx/storage/pkg/store"
)

var (
	path       string
	etcd       string
	me         string
	anotherURL string
	port       int
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&path, "config", "/root/.kube/config", "use to set config file path")
	flag.StringVar(&etcd, "etcd", "127.0.0.1:2389", "coordinator url")
	flag.StringVar(&me, "me", "fancy", "specify who am I ")
	flag.StringVar(&anotherURL, "another", "127.0.0.1:8888", "another coordinator url")
	flag.IntVar(&port, "port", 8888, "listen port ")

}

func main() {
	flag.Parse()
	db := store.NewEtcd([]string{etcd})
	cli, informerFactory := controller.NewCliAndInformer(path)
	pvController := controller.NewPVController(informerFactory)
	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
	pvcController := controller.NewRealPVCControl(pvcInformer.Lister())
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	s := server.NewServer(db, pvController, pvcController)
	apiServer := api.NewServer(cli, informerFactory, port, db, me, anotherURL)
	go apiServer.Run()
	s.Run()
}
