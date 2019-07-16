package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	path := "/root/.kube/config"
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("create kubernetes config error, err=%+v", err)
		panic(err)
	}
	clienset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("create kubernetes client error, err=%+v", err)
		panic(err)
	}
	informerFactory := informers.NewSharedInformerFactory(clienset, time.Second*30)
	//controller := controller.NewController(clienset, informerFactory)
	pvController := controller.NewPVController(informerFactory)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	ns, pv, err := pvController.GetPVCByPV("local-pv-96fb97f2")
	if err != nil {
		glog.Errorf("get pv error, err=%+v", err)
		panic(err)
	}
	glog.Infof("namespace is %s, name is %s", ns, pv)
}
