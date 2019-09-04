package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/pborman/uuid"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

var (
	volumes = make(map[string]*LvmVolume)
	prefix  = "nsenter --mount=/proc/1/ns/mnt"
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
	vol := &LvmVolume{}
	vol.PVName = req.Name
	if req.GetCapacityRange() != nil {
		vol.Size = req.GetCapacityRange().GetRequiredBytes()
	} else {
		vol.Size = 1
	}
	storeVol := GetLVMVolumeByPVName(vol.PVName)
	if storeVol != nil {
		if storeVol.Size != vol.Size {
			return nil, status.Errorf(codes.Internal, "disk %s size is different with requested for disk", req.GetName())
		} else {
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      storeVol.VolumeId,
					CapacityBytes: storeVol.Size,
					VolumeContext: req.GetParameters(),
				},
			}, nil
		}
	}
	vol.VolumeId = uuid.NewUUID().String()
	vol.VolumeGroup = req.GetParameters()["kind"]
	err := vol.Create(prefix)
	if err != nil {
		glog.Errorf("create logical volume error, err=%+v", err)
		return nil, status.Errorf(codes.Internal, "create logical volume error, err=%+v", err)
	}
	maj, min, ok := GetDeviceNum(vol, prefix)
	if !ok {
		glog.Errorf("get device number error")
		return nil, status.Errorf(codes.Internal, "get device number error")
	}
	vol.Maj = maj
	vol.Min = min
	volumes[vol.VolumeId] = vol
	glog.Infof("create volume success,volId=%s", vol.VolumeId)
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      vol.VolumeId,
			CapacityBytes: vol.Size,
			VolumeContext: req.GetParameters(),
		},
	}, nil

}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	glog.Infof("DeleteVolumes: Starting delete volume %s", req.GetVolumeId())
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.Errorf("DeleteVolume: Invaild delete volume args %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "DeleteVolume: invalid delete volume args %v", err)
	}
	vol, ok := volumes[req.VolumeId]
	if !ok {
		glog.Errorf("can not find volume by id, id=%s", req.VolumeId)
		//return nil, status.Errorf(codes.InvalidArgument, "DeleteVolume: can't find volume")
		return &csi.DeleteVolumeResponse{}, nil
	}
	err := vol.Delete(prefix)
	if err != nil {
		glog.Errorf("delete lv error, err=%+v", err)
		return nil, status.Errorf(codes.Internal, "delete lv error, err=%+v", err)
	}
	return &csi.DeleteVolumeResponse{}, nil
}
func (cs *controllerServer) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	vol, ok := volumes[req.VolumeId]
	if !ok {
		glog.Errorf("can't find volume by volumeId, id=%s", req.VolumeId)
		return &csi.ControllerPublishVolumeResponse{}, nil
	}
	rpcVolume := &cdpb.Volume{
		Name:          vol.PVName,
		VolumeGroup:   vol.VolumeGroup,
		Uuid:          vol.VolumeId,
		Maj:           vol.Maj,
		Min:           vol.Min,
		LogicalVolume: vol.LVName,
	}
	if err := rpc.PutVolume(ctx, rpcVolume.Name, rpcVolume); err != nil {
		glog.Errorf("rpc put volume error, err=%+v", err)
		return nil, status.Error(codes.Internal, "rpc put volume error")
	}
	return &csi.ControllerPublishVolumeResponse{}, nil
}
func (cs *controllerServer) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	glog.Infof("do not support ControllerUnpublishVolume now")
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}
