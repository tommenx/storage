package api

import (
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type exec struct {
	kubeClient      kubernetes.Interface
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
}

type Exec interface {
	SetPod(args *types.SetPodArgs) (*types.SetPodResult, error)
	Run(stopCh <-chan struct{})
}

func NewExec(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory) Exec {
	podInformer := informerFactory.Core().V1().Pods()
	return &exec{
		kubeClient:      kubeCli,
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
	}
}

func (e *exec) Run(stopCh <-chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, e.podListerSynced) {
		return
	}
}

func (e *exec) SetPod(args *types.SetPodArgs) (*types.SetPodResult, error) {
	resp := &types.SetPodResult{}
	ns := args.Namespace
	name := args.Pod
	label := args.StorageLabel
	pod, err := e.podLister.Pods(ns).Get(name)
	if err != nil {
		glog.Errorf("get pod %s/%s error, err=%+s", ns, pod, err)
		resp.Code = 1
		resp.Message = "get pod error"
		return resp, nil
	}
	pod.Annotations["storage.io/label"] = label
	_, err = e.kubeClient.CoreV1().Pods(ns).Update(pod)
	if err != nil {
		glog.Errorf("update pod %s/%s annotation error, err=%+v", ns, name, err)
		resp.Code = 2
		resp.Message = "update pod annotation error"
		return resp, nil
	}
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}
