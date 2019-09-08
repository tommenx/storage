//1. 获取定义的存储的类型
//2. 根据pod的name和ns获取对应的cgroup_path，
//		a.如果存在，对其进行操作
//      b.如果不存在，需要通过dockerid获取cgroup_path
//3. 根据pod的pvc和ns去获取对应的信息，包括主副设备号
//4. 设置存储设备的隔离
package controller

import (
	"context"
	"errors"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/cdpb"
	listers "github.com/tommenx/storage/pkg/client/listers/storage.io/v1alpha1"
	"github.com/tommenx/storage/pkg/consts"
	"github.com/tommenx/storage/pkg/container"
	"github.com/tommenx/storage/pkg/isolate"
	"github.com/tommenx/storage/pkg/rpc"
	corev1 "k8s.io/api/core/v1"
)

type volumeControl struct {
	dockerController container.ContainerInterafce
	slController     StorageLabel
	nodeName         string
}

type VolumeControlInterface interface {
	Sync(pod *corev1.Pod) error
}

func NewVolumeControl(slLister listers.StorageLabelLister, nodeName string) VolumeControlInterface {
	slController := NewStorageLabelController(slLister)
	return &volumeControl{
		dockerController: container.NewClient(),
		slController:     slController,
		nodeName:         nodeName,
	}
}

func (c *volumeControl) Sync(pod *corev1.Pod) error {
	ns := pod.Namespace
	name := pod.Name
	ctx := context.Background()
	cgroupParentPath := ""
	dockerId := ""
	label, ok := pod.Annotations["storage.io/label"]
	if !ok {
		glog.Errorf("pod %s/%s do not identify storage label", ns, name)
		return errors.New("do not identify storage label")
	}
	requestResource, err := c.slController.GetStorageLabel(label)
	glog.Infof("storage label %s is %+v", label, requestResource)
	if err != nil {
		glog.Errorf("get storage label error, label=%s, err=%+v", label, err)
		return err
	}
	podResource, err := rpc.GetPodResource(ctx, ns, name)
	if err != nil && err != consts.ErrNotExist {
		glog.Errorf("get pod resource error, pod=%s/%s, err=%+v", ns, name, err)
		return err
	}
	if podResource == nil {
		podResource = &cdpb.PodResource{}
	}
	if err != nil {
		cgroupParentPath = podResource.CgroupPath
		dockerId = podResource.DockerId
	}
	if cgroupParentPath == "" || dockerId == "" {
		//查找对应的docker的id
		if len(dockerId) == 0 {
			dockerId, err = GetDockerIdByPod(pod)
			if err != nil {
				glog.Errorf("get dockerId by Pod error, err=%+v", err)
				return err
			}
		}
		cgroupParentPath, err = c.dockerController.GetCgroupPath(ctx, dockerId)
		if err != nil {
			glog.Errorf("get cGroupParentPath error, err=%+v", err)
			return err
		}
		podResource.Namespace = ns
		podResource.Name = name
		podResource.Node = c.nodeName
		podResource.CgroupPath = cgroupParentPath
		podResource.DockerId = dockerId
		podResource.RequestResource = requestResource
		podResource.Level = consts.LevelSSD
		err := rpc.DirectPutPodResource(ctx, podResource, consts.OpAdd)
		if err != nil {
			glog.Errorf("update pod resource error, err=%+v", err)
			return err
		}
	}
	pvc := GetPVCName(pod)
	if len(pvc) == 0 {
		glog.Errorf("can't get pvc name, pod=%s", name)
		return consts.ErrNotFound
	}
	volume, err := rpc.GetVolume(ctx, ns, pvc)
	if err != nil {
		glog.Errorf("get volume error, pvc=%s, err=%+v", pvc, err)
		return err
	}
	err = isolate.SetBlkio(cgroupParentPath, dockerId, requestResource, volume.Maj, volume.Min)
	if err != nil {
		glog.Errorf("set volume isolate error, err=%+v", err)
		return err
	}
	glog.Infof("set isolate success, pod=%s", name)
	return nil

}
