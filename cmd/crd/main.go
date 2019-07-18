package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/controller"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
)

func init() {
	flag.Set("logtostderr", "true")
}
func main() {
	flag.Parse()
	path := "/root/.kube/config"
	informerFactory := controller.NewSLInformerFactory(path)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	slController := controller.NewFakeStorageLabelController(informerFactory)
	slController.Run(stop)
	request, err := slController.GetStorageLabel("fast")
	if err != nil {
		glog.Errorf("get storage label error")
	}
	fmt.Printf("%+v", request)
}
