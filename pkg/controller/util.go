package controller

import (
	"errors"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/client/clientset/versioned"
	imformers "github.com/tommenx/storage/pkg/client/informers/externalversions"
	"github.com/tommenx/storage/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	resyncDuration = time.Second * 30
)

func NewSharedInformerFactory(path string) kubeinformers.SharedInformerFactory {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("create kubernetes config error, err=%+v", err)
		panic(err)
	}
	clienset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("create kubernetes client error, err=%+v", err)
		panic(err)
	}
	informerFactory := kubeinformers.NewSharedInformerFactory(clienset, resyncDuration)
	return informerFactory
}

func NewCliAndInformer(path string) (kubernetes.Interface, kubeinformers.SharedInformerFactory) {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("create kubernetes config error, err=%+v", err)
		panic(err)
	}
	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Errorf("create kubernetes client error, err=%+v", err)
		panic(err)
	}
	informerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, resyncDuration)
	return kubeCli, informerFactory
}

func NewSLInformerFactory(path string) imformers.SharedInformerFactory {
	cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		glog.Errorf("create kubernetes config error, err=%+v", err)
		panic(err)
	}
	cli, err := versioned.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("failed to create Clientset: %v", err)
	}
	informerFactory := imformers.NewSharedInformerFactory(cli, resyncDuration)
	return informerFactory
}

func GetDockerIdByPod(pod *corev1.Pod) (string, error) {
	name, ok := pod.Annotations["storage.io/docker"]
	if !ok {
		return "", errors.New("do not specify docker name")
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == name {
			//去除开头的docker://
			if len(container.ContainerID) > 8 {
				return container.ContainerID[8:], nil
			} else {
				return "", errors.New("error container id")
			}
		}
	}
	return "", consts.ErrNotFound
}

func GetPVCName(pod *corev1.Pod) string {
	pvcName := ""
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcName = volume.PersistentVolumeClaim.ClaimName
		}
	}
	return pvcName
}
