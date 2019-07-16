package controller

import (
	"github.com/golang/glog"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

func NewSharedInformerFactory(path string) informers.SharedInformerFactory {
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
	return informerFactory
}
