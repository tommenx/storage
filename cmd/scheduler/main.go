package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/controller"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	path := "/root/.kube/config"
	informerFactory := controller.NewSharedInformerFactory(path)
	pvController := controller.NewPVController(informerFactory)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	ns, pv, err := pvController.GetPVCByPV("pvc-aaa22979-ae09-11e9-aba2-309c23e8d374")
	if err != nil {
		glog.Errorf("get pv error, err=%+v", err)
	}
	glog.Infof("namespace is %s, name is %s", ns, pv)
}
