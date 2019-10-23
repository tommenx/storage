package api

import (
	"context"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/store"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
)

type executor struct {
	kubeClient      kubernetes.Interface
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
	podControl      controller.PodControlInterface
	db              store.EtcdInterface
}

type Executor interface {
	Run(stopCh <-chan struct{})
	SetOnePod(args *types.SetOnePodArgs) (*types.SetPodResult, error)
	SetBatchPod(args *types.SetBatchPodArgs) (*types.SetPodResult, error)
	GetInstanceUtil(args *types.GetInstanceArgs) (*types.GetInstanceResult, error)
}

func NewExecutor(kubeCli kubernetes.Interface, informerFactory informers.SharedInformerFactory, db store.EtcdInterface) Executor {
	podInformer := informerFactory.Core().V1().Pods()
	control := controller.NewRealPodControl(kubeCli, podInformer.Lister())
	return &executor{
		kubeClient:      kubeCli,
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		podControl:      control,
		db:              db,
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

func (e *executor) GetInstanceUtil(args *types.GetInstanceArgs) (*types.GetInstanceResult, error) {
	resp := &types.GetInstanceResult{}
	info, err := e.db.GetAlivePodInfo(context.Background(), "check")
	if err != nil {
		glog.Errorf("get instance error,err=%+v", err)
		resp.Code = 1
		resp.Message = "get instance util error"
		return resp, nil
	}
	instances := make([]types.Instance, 0)
	for name, val := range info {
		util := strings.Split(val, "-")
		instances = append(instances, types.Instance{Name: name, Read: util[0], Write: util[1]})
	}
	resp.Instances = instances
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}
