package api

import (
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type executor struct {
	kubeClient      kubernetes.Interface
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
	podControl      controller.PodControlInterface
}

type Executor interface {
	Run(stopCh <-chan struct{})
	SetOnePod(args *types.SetOnePodArgs) (*types.SetPodResult, error)
	SetBatchPod(args *types.SetBatchPodArgs) (*types.SetPodResult, error)
}

func NewExecutor(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory) Executor {
	podInformer := informerFactory.Core().V1().Pods()
	control := controller.NewRealPodControl(kubeCli, podInformer.Lister())
	return &executor{
		kubeClient:      kubeCli,
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		podControl:      control,
	}
}

func (e *executor) Run(stopCh <-chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, e.podListerSynced) {
		return
	}
}

func (e *executor) SetBatchPod(args *types.SetBatchPodArgs) (*types.SetPodResult, error) {
	resp := &types.SetPodResult{}
	tag := args.Tag
	val := args.Val
	selector := make(map[string]string)
	selector[tag] = val
	err := e.podControl.SetBatchPod(selector, args.Read, args.Write)
	if err != nil {
		glog.Errorf("set pods %s:%s label error, err=%+v", tag, val, err)
		resp.Code = 2
		resp.Message = "update pods annotation error"
		return resp, nil
	}
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}

func (e *executor) SetOnePod(args *types.SetOnePodArgs) (*types.SetPodResult, error) {
	resp := &types.SetPodResult{}
	ns := args.Namespace
	err := e.podControl.SetOnePod(ns, args.Requests)
	if err != nil {
		glog.Errorf("update pod annotation error, err=%s", err.Error())
		resp.Code = 2
		resp.Message = "update pod annotation error"
		return resp, nil
	}
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}
