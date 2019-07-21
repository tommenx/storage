package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/rpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
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
	informerFactory := controller.NewSLInformerFactory(path)
	stopCh := make(chan struct{})
	go informerFactory.Start(stopCh)
	slInformer := informerFactory.Storage().V1alpha1().StorageLabels()
	slListerSyned := slInformer.Informer().HasSynced
	if !cache.WaitForCacheSync(stopCh, slListerSyned) {
		return
	}
	slController := controller.NewVolumeControl(slInformer.Lister())
	pod, err := clienset.CoreV1().Pods("default").Get("test-pod-5", metav1.GetOptions{})
	if err != nil {
		glog.Errorf("get pod info error, err=%+v", err)
		panic(err)
	}
	slController.Sync(pod)
}
