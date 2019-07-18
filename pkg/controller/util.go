package controller

import (
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/client/clientset/versioned"
	imformers "github.com/tommenx/storage/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	resyncDuration = time.Second * 30
)

func NewSharedInformerFactory(path string) kubeinformers.SharedInformerFactory {
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
	informerFactory := kubeinformers.NewSharedInformerFactory(clienset, resyncDuration)
	return informerFactory
}

func NewSLInformerFactory(path string) imformers.SharedInformerFactory {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("create kubernetes config error, err=%+v", err)
		panic(err)
	}
	cli, err := versioned.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("failed to create Clientset: %v", err)
	}
	informerFactory := imformers.NewSharedInformerFactory(cli, resyncDuration)
	return informerFactory
}
