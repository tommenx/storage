package controller

import (
	"github.com/golang/glog"
	"k8s.io/client-go/informers"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type pvController struct {
	pvLister       corelisters.PersistentVolumeLister
	pvListerSynced cache.InformerSynced
}

type PVController interface {
	Run(stop <-chan struct{})
	GetPVCByPV(pvName string) (string, string, error)
}

func NewPVController(informerFactory informers.SharedInformerFactory) PVController {
	pvInformfer := informerFactory.Core().V1().PersistentVolumes()
	controller := &pvController{}
	controller.pvLister = pvInformfer.Lister()
	controller.pvListerSynced = pvInformfer.Informer().HasSynced
	return controller
}

func (c *pvController) Run(stopCh <-chan struct{}) {
	if !cache.WaitForCacheSync(stopCh, c.pvListerSynced) {
		glog.Error("sync pv timeout")
		return
	}
}

func (c *pvController) GetPVCByPV(pvName string) (string, string, error) {
	pv, err := c.pvLister.Get(pvName)
	if err != nil {
		glog.Errorf("get pv info error, pv name=%s, err=%+v", pvName, err)
		return "", "", nil
	}

	return pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, nil
}
