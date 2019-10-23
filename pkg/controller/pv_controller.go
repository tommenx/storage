package controller

import (
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/consts"
	corev1 "k8s.io/api/core/v1"
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
	GetPV(volumeId string) (*corev1.PersistentVolume, error)
	GetPodByPV(pvName string) (string, string, error)
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
		return "", "", err
	}
	if pv == nil {
		glog.Errorf("pv not found, pv name=%s", pvName)
		return "", "", consts.ErrNotFound
	}
	if pv.Spec.ClaimRef == nil {
		glog.Errorf("can not found pv claim, pv name=%s", pvName)
		return "", "", consts.ErrNotBound
	}
	return pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, nil
}

func (c *pvController) GetPV(volumeId string) (*corev1.PersistentVolume, error) {
	return c.pvLister.Get(volumeId)
}

// namespace,name,error
func (c *pvController) GetPodByPV(pvName string) (string, string, error) {
	pvInfo, err := c.pvLister.Get(pvName)
	if err != nil {
		glog.Errorf("get pv info error, err=%+v", err)
		return "", "", err
	}
	glog.Infof("annotation: %+v", pvInfo.Annotations)
	podName, ok := pvInfo.Annotations["tidb.pingcap.com/pod-name"]
	if !ok {
		return "", "", consts.ErrNotBound
	}
	return "default", podName, nil
}
