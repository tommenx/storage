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
	"net/http"
	"time"
)

var (
	nodeName    string
	configPath  string
	coordinator string
)

var (
	LOGFILE_PREFIX = "/var/log/alicloud/"
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&nodeName, "node", "localhost.localdomain", "use to identify node")
	flag.StringVar(&configPath, "config", "./config.toml", "use to set config file path")
	flag.StringVar(&coordinator, "coordinator", "10.48.247.109:50051", "coordinator url")
}

func main() {
	flag.Parse()
	rpc.Init(coordinator)
	config.Init(configPath)
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
	watch := watcher.NewWatcher(time.Second*5, nodeName)
	if err := watch.InitResource(); err != nil {
		glog.Errorf("init node resource error, err=%+v", err)
		panic(err)
	}
	stopCh := make(chan struct{})
	go kubeInformerFactory.Start(stopCh)
	go informerFactory.Start(stopCh)
	go watch.Run(stopCh)
	// 添加监控接口
	go controller.Run(1, stopCh)
	glog.Fatal(http.ListenAndServe(":50053", nil))
}
