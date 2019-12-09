package api

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"github.com/tommenx/storage/pkg/consts"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/store"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"strconv"
	"strings"
	"time"
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
	GetInstanceUtil() (*types.GetInstanceResult, error)
	Query(what, which string) (*types.QueryResult, error)
	GetResourceTimeCompletion(which string) (*types.ResourceCompletionResult, error)
	PutSetting(key, val string) (*types.SetPodResult, error)
	GetInstanceUseFree() (*types.GetInstanceUseFreeResult, error)
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

func (e *executor) getInstanceUtil() ([]types.Instance, error) {
	info, err := e.db.GetAlivePodInfo(context.Background(), "check")
	if err != nil {
		glog.Errorf("get instance error,err=%+v", err)
		return nil, err
	}
	instances := make([]types.Instance, 0)
	for name, val := range info {
		util := strings.Split(val, "-")
		instances = append(instances, types.Instance{Name: name, Read: util[0], Write: util[1]})
	}
	return instances, nil
}

func (e *executor) GetInstanceUtil() (*types.GetInstanceResult, error) {
	resp := &types.GetInstanceResult{}
	instances, err := e.getInstanceUtil()
	if err != nil {
		resp.Code = 1
		resp.Message = "get instance util error"
		return resp, nil
	}
	resp.Instances = instances
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}

func (e *executor) GetInstanceUseFree() (*types.GetInstanceUseFreeResult, error) {
	resp := &types.GetInstanceUseFreeResult{}
	utils, err := e.getInstanceUtil()
	if err != nil {
		glog.Errorf("get instance error,err=%+v", err)
		resp.Code = 1
		resp.Message = "get instance util error"
		return resp, nil
	}
	limits, err := e.db.GetAlivePodInfo(context.Background(), "limit")
	if err != nil {
		glog.Errorf("get instance limit error,err=%+v", err)
		resp.Code = 1
		resp.Message = "get instance limit error"
		return resp, nil
	}
	instances := make([]*types.InstanceUseFree, 0)
	for _, util := range utils {
		useStr := util.Write
		name := util.Name
		use, _ := strconv.ParseFloat(useStr, 64)
		limit, _ := strconv.Atoi(limits[name])
		free := limit - int(use)
		instances = append(instances, &types.InstanceUseFree{
			Name: name,
			Use:  int(use),
			Free: free,
		})
	}
	resp.Instances = instances
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}

func (e *executor) Query(what, which string) (*types.QueryResult, error) {
	var err error
	var pair map[string][]byte
	resp := &types.QueryResult{}
	var key string
	if what == "resource_time" {
		key = fmt.Sprintf("/storage/show/%s/resourceTime", which)
	} else if what == "resource_allocation" {
		key = fmt.Sprintf("/storage/show/%s/resourceAllocation", which)
	} else if what == "qps" {
		key = fmt.Sprintf("/storage/ycsb/log/%s", which)
	} else if what == "request_qps" {
		key = fmt.Sprintf("/storage/ycsb/requestQPS/%s", which)
	}
	pair, err = e.db.Get(context.Background(), key, false)
	//出错重试5次，直到取到值
	count := 5
	for err == consts.ErrNotExist {
		pair, err = e.db.Get(context.Background(), key, false)
		count--
		time.Sleep(100 * time.Millisecond)
		if count == 0 {
			resp.Val = "0"
			resp.Code = 0
			resp.Message = "success"
			return resp, nil
		}
	}
	if err != nil {
		resp.Code = 1
		resp.Message = fmt.Sprintf("get %s error", key)
		return resp, nil
	}
	data := pair[key]
	resp.Val = string(data)
	resp.Code = 0
	resp.Message = "success"
	return resp, nil
}

func (e *executor) GetResourceTimeCompletion(which string) (*types.ResourceCompletionResult, error) {
	keyResourceTime := fmt.Sprintf("/storage/show/%s/resourceTime", which)
	keyRequestOperation := fmt.Sprintf("/storage/ycsb/reuqestOperation/%s", which)
	keyOperation := fmt.Sprintf("/storage/ycsb/operation/%s", which)
	resourceTime := e.QueryInt(keyResourceTime)
	op := e.QueryInt(keyOperation)
	requestOp := e.QueryInt(keyRequestOperation)
	per := op * 100 / requestOp
	return &types.ResourceCompletionResult{
		ResourceTime: resourceTime,
		Completion:   per,
		Code:         0,
		Message:      "success",
	}, nil
}

func (e *executor) QueryInt(key string) int {
	pair, err := e.db.Get(context.Background(), key, false)
	if err != nil {
		glog.Errorf("get %s error,err=%+v", key, err)
		return 1
	}
	data := pair[key]
	val, _ := strconv.Atoi(string(data))
	return val
}

func (e *executor) PutSetting(key, val string) (*types.SetPodResult, error) {
	res := &types.SetPodResult{}
	k := fmt.Sprintf("/storage/setting/%s", key)
	err := e.db.Put(context.Background(), k, []byte(val))
	if err != nil {
		glog.Errorf("put setting %s error,err=%+v", key, err)
		res.Code = 1
		res.Message = "put setting error"
		return res, nil
	}
	res.Code = 0
	res.Message = "success"
	return res, nil
}
