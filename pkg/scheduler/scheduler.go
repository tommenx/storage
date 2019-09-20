package scheduler

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	informers "github.com/tommenx/storage/pkg/client/informers/externalversions"
	"github.com/tommenx/storage/pkg/consts"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
	"math"
)

type Scheduler interface {
	Filter(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.ExtenderFilterResult, error)
	Priority(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.HostPriorityList, error)
	Bind(args *schedulerapiv1.ExtenderBindingArgs) (*schedulerapiv1.ExtenderBindingResult, error)
}

type scheduler struct {
	slListerSynced cache.InformerSynced
	storageLabel   controller.StorageLabel
}

func NewScheduler(informerFactory informers.SharedInformerFactory) Scheduler {
	s := &scheduler{}
	slInformer := informerFactory.Storage().V1alpha1().StorageLabels()
	s.slListerSynced = slInformer.Informer().HasSynced
	s.storageLabel = controller.NewStorageLabelController(slInformer.Lister())
	return s
}

func (s *scheduler) getRequirements(pod *corev1.Pod) (map[string]int64, error) {
	label := pod.Annotations["storage.io/label"]
	size := pod.Annotations["storage.io/space"]
	sz := utils.GetInt64(size)
	want, err := s.storageLabel.GetStorageLabel(label)
	want["space"] = sz
	if err != nil {
		glog.Errorf("get storage label error, err=%+v", err)
		return nil, fmt.Errorf("get request storage resource error")
	}
}

func (s *scheduler) Filter(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.ExtenderFilterResult, error) {
	nodeNames := *args.NodeNames
	nodeInfos, err := rpc.GetNodeStorage(context.Background(), "all", consts.KindAllocation)
	canSchedule := make([]string, 0, len(nodeNames))
	canNotSchedule := make(map[string]string)
	rsp := &schedulerapiv1.ExtenderFilterResult{}
	if err != nil {
		glog.Errorf("filter: get node storage error, err=%+v", err)
		rsp.NodeNames = args.NodeNames
		rsp.FailedNodes = canNotSchedule
		rsp.Error = "get node storage info failed"
		return rsp, nil
	}
	want, err := s.getRequirements(args.Pod)
	if err != nil {
		glog.Errorf("filter: get storage label error, err=%+v", err)
		rsp.NodeNames = args.NodeNames
		rsp.FailedNodes = canNotSchedule
		rsp.Error = err.Error()
		return rsp, nil
	}
	for nodename, info := range nodeInfos {
		ok := true
		for _, storage := range info.Storage {
			for k, v := range want {
				if storage.Resource[k] > v {
					continue
				} else {
					ok = false
					glog.Infof("node %s %s not enough, want %d, have %d", nodename, k, v, storage.Resource[k])
					canNotSchedule[nodename] = fmt.Sprintf("%s not enough,want %d have %d", k, v, storage.Resource[k])
				}
			}
		}
		if ok {
			canSchedule = append(canSchedule, nodename)
		}
	}
	rsp.NodeNames = &canSchedule
	rsp.FailedNodes = canNotSchedule
	rsp.Error = "schedule success"
	return rsp, nil
}

func (s *scheduler) Priority(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.HostPriorityList, error) {
	hostPriorityList := schedulerapiv1.HostPriorityList{}
	request, err := s.getRequirements(args.Pod)
	if err != nil {
		glog.Errorf("priority: get request resource error")
		return &hostPriorityList, nil
	}
	nodeInfo, err := rpc.GetNodeStorage(context.Background(), "all", consts.KindAllocation)
	if err != nil {
		glog.Error("priority: get node storage info error, err=%+v", err)
		return &hostPriorityList, nil
	}
	for name, info := range nodeInfo {
		score := float64(0)
		for _, storage := range info.Storage {
			for k, v := range request {
				score += float64(v) * 1.0 / float64(storage.Resource[k])
			}
		}
		scoreVal := int(math.Floor(score))
		hostPriorityList = append(hostPriorityList, schedulerapiv1.HostPriority{
			Host:  name,
			Score: scoreVal,
		})
	}
	return &hostPriorityList, nil
}

func (s *scheduler) Bind(args *schedulerapiv1.ExtenderBindingArgs) (*schedulerapiv1.ExtenderBindingResult, error) {
	return nil, nil
}

func (s *scheduler) WaitSync(stop <-chan struct{}) {
	if !cache.WaitForCacheSync(stop, s.slListerSynced) {
		return
	}
}
