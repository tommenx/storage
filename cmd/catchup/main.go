package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/watcher"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	nodeName string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&nodeName, "node", "localhost.localdomain", "use to identify node")
}

func main() {
	flag.Parse()
	rpc.Init("10.48.144.34:50051")
	config.Init("../../config.toml")
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
	controller := controller.NewController(clienset, kubeInformerFactory, informerFactory, nodeName)
	watch := watcher.NewWatcher(time.Second * 60)
	err = watch.InitResource()
	if err != nil {
		glog.Errorf("init node resource error, err=%+v", err)
		panic(err)
	}
	stopCh := make(chan struct{})
	go kubeInformerFactory.Start(stopCh)
	go informerFactory.Start(stopCh)
	go watch.Run(stopCh)
	// 添加监控接口
	controller.Run(1, stopCh)
}
