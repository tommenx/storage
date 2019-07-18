package controller

import (
	"github.com/golang/glog"
	imformers "github.com/tommenx/storage/pkg/client/informers/externalversions"
	listers "github.com/tommenx/storage/pkg/client/listers/storage.io/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type storageLabelController struct {
	slLister       listers.StorageLabelLister
	slListerSynced cache.InformerSynced
}

type storageLabelControllerInterafce interface {
	GetStorageLabel(name string) (map[string]int64, error)
	Run(stopCh <-chan struct{})
}

func NewStorageLabelController(slLister listers.StorageLabelLister) storageLabelControllerInterafce {
	return &storageLabelController{
		slLister: slLister,
	}
}

func (c *storageLabelController) Run(stopCh <-chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, c.slListerSynced) {
		glog.Error("sync storage label timeout")
		return
	}
}

func (c *storageLabelController) GetStorageLabel(name string) (map[string]int64, error) {
	storageLabel, err := c.slLister.StorageLabels("default").Get(name)
	if err != nil {
		glog.Errorf("get storage label error, err=%+v", err)
		return nil, err
	}
	request := make(map[string]int64)
	writeBpsDevice := storageLabel.Spec.WriteBpsDevice
	writeIopsDevice := storageLabel.Spec.WriteIopsDevice
	readBpsDevice := storageLabel.Spec.ReadBpsDevice
	readIopsDevice := storageLabel.Spec.ReadIopsDevice
	request["write_bps_device"] = writeBpsDevice
	request["write_iops_device"] = writeIopsDevice
	request["read_bps_device"] = readBpsDevice
	request["read_iops_device"] = readIopsDevice
	return request, nil
}

func NewFakeStorageLabelController(informerFactory imformers.SharedInformerFactory) storageLabelControllerInterafce {
	slInformer := informerFactory.Storage().V1alpha1().StorageLabels()
	return NewStorageLabelController(slInformer.Lister())
}
