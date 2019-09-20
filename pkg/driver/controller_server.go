package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

var (
	prefix = "nsenter --mount=/proc/1/ns/mnt"
)

func NewControllerServer(d *csicommon.CSIDriver) csi.ControllerServer {
	c := &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
	return c
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("CreateVolume: driver not support Create volume: %v", err)
		return nil, err
	}
	if len(req.Name) == 0 {
		glog.Errorf("CreateVolume:Volume name cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "Volume Name cannot be empty")
	}
	if req.VolumeCapabilities == nil {
		glog.Errorf("CreateVolume: Volume Capabilities cannot be empty")
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities cannot be empty")
	}
	if _, ok := req.GetParameters()["kind"]; !ok {
		glog.Errorf("CreateVolume: error VolumeGroup from input")
		return nil, status.Error(codes.InvalidArgument, "CreateVolume: error VolumeGroup from input")
	}
	volumeId := req.GetName()
	rsp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeId,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
		},
	}
	glog.Infof("success create volume %s, size %d", volumeId, req.GetCapacityRange().GetRequiredBytes())
	return rsp, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	glog.Infof("DeleteVolumes: success delete volume %s", req.GetVolumeId())
	return &csi.DeleteVolumeResponse{}, nil
}
func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	glog.Infof("ControllerPublishVolume: do not support yet")
	return &csi.ControllerPublishVolumeResponse{}, nil
}
func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	glog.Infof("ControllerPublishVolume: do not support yet")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
