package controller

import (
	"errors"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/api/types"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type realPodControl struct {
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
}

type PodControlInterface interface {
	SetOnePod(ns string, requests []types.Request) error
	SetBatchPod(args map[string]string, read, write string) error
}

func NewRealPodControl(kubeCli kubernetes.Interface, podLister corelisters.PodLister) PodControlInterface {
	return &realPodControl{
		kubeCli:   kubeCli,
		podLister: podLister,
	}
}

func (c *realPodControl) SetOnePod(ns string, requests []types.Request) error {
	for _, request := range requests {
		pod, err := c.podLister.Pods(ns).Get(request.Pod)
		if err != nil {
			glog.Errorf("get pod %s/%s error, err=%+s", ns, pod, err)
			return errors.New("get pod info error")
		}
		if request.Read != "" {
			pod.Annotations["storage.io/read"] = request.Read
		}
		if request.Write != "" {
			pod.Annotations["storage.io/write"] = request.Write
		}
		_, err = c.kubeCli.CoreV1().Pods(ns).Update(pod)
		if err != nil {
			glog.Errorf("update pod %s/%s annotation error, err=%+v", ns, request.Pod, err)
			return errors.New("update pod info error")
		}
	}
	return nil
}

func (c *realPodControl) SetBatchPod(args map[string]string, read, write string) error {
	kv := labels.SelectorFromSet(args)
	pods, err := c.podLister.List(kv)
	if err != nil {
		glog.Errorf("list pod error, err=%+v", err)
		return err
	}
	for _, pod := range pods {
		if read != "" {
			pod.Annotations["storage.io/read"] = read
		}
		if write != "" {
			pod.Annotations["storage.io/write"] = write
		}
		_, err := c.kubeCli.CoreV1().Pods(pod.Namespace).Update(pod)
		if err != nil {
			glog.Errorf("set pod %s/%s label %s error, err=%+v", pod.Namespace, pod.Name, err)
			continue
		}
	}
	return nil
}
