package scheduler

import (
	informers "github.com/tommenx/storage/pkg/client/informers/externalversions"
	"github.com/tommenx/storage/pkg/controller"
	"k8s.io/client-go/tools/cache"
	schedulerapiv1 "k8s.io/kubernetes/pkg/scheduler/api/v1"
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

func (s *scheduler) Filter(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.ExtenderFilterResult, error) {
	//nodes := *args.NodeNames
	//ctx := context.Background()
	//canSchedule := []string{}
	//canNotSchedule := []string{}
	//nodeStorages, err := rpc.GetNodeStorage(ctx)
	//if err != nil {
	//	glog.Errorf("get node storage error,err=%+v", err)
	//	return nil, err
	//}
	//storageLabel := args.Pod.Annotations["storage.io/label"]
	//want, err := s.storageLabel.GetStorageLabel(storageLabel)
	//if err != nil {
	//	glog.Errorf("get request storage resource error, label=%s, err=%v", storageLabel, err)
	//	return nil, err
	//}
	return nil, nil
}

func (s *scheduler) Priority(args *schedulerapiv1.ExtenderArgs) (*schedulerapiv1.HostPriorityList, error) {
	return nil, nil
}

func (s *scheduler) Bind(args *schedulerapiv1.ExtenderBindingArgs) (*schedulerapiv1.ExtenderBindingResult, error) {
	return nil, nil
}

func (s *scheduler) WaitSync(stop <-chan struct{}) {
	if !cache.WaitForCacheSync(stop, s.slListerSynced) {
		return
	}
}
