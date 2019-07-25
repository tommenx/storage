package controller

import (
	"github.com/golang/glog"
	listers "github.com/tommenx/storage/pkg/client/listers/storage.io/v1alpha1"
)

type storageLabelController struct {
	slLister listers.StorageLabelLister
}

type StorageLabel interface {
	GetStorageLabel(name string) (map[string]int64, error)
}

func NewStorageLabelController(slLister listers.StorageLabelLister) StorageLabel {
	return &storageLabelController{
		slLister: slLister,
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
