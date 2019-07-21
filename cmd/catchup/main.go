package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/rpc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	rpc.Init("10.48.233.0:50051")
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
	kubeInformerFactory := controller.NewSharedInformerFactory(path)
	informerFactory := controller.NewSLInformerFactory(path)
	controller := controller.NewController(clienset, kubeInformerFactory, informerFactory)
	stopCh := make(chan struct{})
	go kubeInformerFactory.Start(stopCh)
	go informerFactory.Start(stopCh)
	controller.Run(1, stopCh)
}
