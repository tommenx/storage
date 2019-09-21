package controller

import (
	"errors"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type realPodControl struct {
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
}

type PodControlInterface interface {
	SetOnePod(ns, name, label string) error
	SetBatchPod(args map[string]string, label string) error
}

func NewRealPodControl(kubeCli kubernetes.Interface, podLister corelisters.PodLister) PodControlInterface {
	return &realPodControl{
		kubeCli:   kubeCli,
		podLister: podLister,
	}
}

func (c *realPodControl) SetOnePod(ns, name, label string) error {
	pod, err := c.podLister.Pods(ns).Get(name)
	if err != nil {
		glog.Errorf("get pod %s/%s error, err=%+s", ns, pod, err)
		return errors.New("get pod info error")
	}
	pod.Annotations["storage.io/label"] = label
	_, err = c.kubeCli.CoreV1().Pods(ns).Update(pod)
	if err != nil {
		glog.Errorf("update pod %s/%s annotation error, err=%+v", ns, name, err)
		return errors.New("update pod info error")
	}
	return nil
}

func (c *realPodControl) SetBatchPod(args map[string]string, label string) error {
	kv := labels.SelectorFromSet(args)
	pods, err := c.podLister.List(kv)
	if err != nil {
		glog.Errorf("list pod error, err=%+v", err)
		return err
	}
	for _, pod := range pods {
		pod.Annotations["storage.io/label"] = label
		_, err := c.kubeCli.CoreV1().Pods(pod.Namespace).Update(pod)
		if err != nil {
			glog.Errorf("set pod %s/%s label %s error, err=%+v", pod.Namespace, pod.Name, err)
			continue
		}
	}
	return nil
}
