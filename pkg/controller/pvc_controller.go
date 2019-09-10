package controller

import (
	corev1 "k8s.io/api/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type PVCControlInterface interface {
	GetPVC(namespace, name string) (*corev1.PersistentVolumeClaim, error)
}

type realPVCControl struct {
	pvcLister corelisters.PersistentVolumeClaimLister
}

func NewRealPVCControl(pvcLister corelisters.PersistentVolumeClaimLister) PVCControlInterface {
	return &realPVCControl{
		pvcLister: pvcLister,
	}
}

func (c *realPVCControl) GetPVC(namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := c.pvcLister.PersistentVolumeClaims(namespace).Get(name)
	return pvc, err
}
