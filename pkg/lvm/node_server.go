package lvm

import (
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/kubernetes/pkg/util/mount"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	mounter mount.Interface
}

func NewNodeServer(d *csicommon.CSIDriver, containerized bool) (*nodeServer, error) {
	mounter := mount.New("")
	if containerized {
	}
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
		mounter:           mounter,
	}, nil
}
