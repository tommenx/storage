package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	mounter Mounter
}

func NewNodeServer(d *csicommon.CSIDriver, containerized bool) (*nodeServer, error) {
	mounter := NewMounter("")
	if containerized {
		mounter = NewMounter(prefix)
	}
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
		mounter:           mounter,
	}, nil
}
func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			nscap,
		},
	}, nil
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	source := req.StagingTargetPath
	targetPath := req.TargetPath
	glog.Infof("NodePublishVolume: start to mount to target path")
	if !strings.HasSuffix(targetPath, "/mount") {
		return nil, status.Errorf(codes.InvalidArgument, "malformed the value of target path: %s", targetPath)
	}
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Volume ID must be provided")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Staging Target Path must be provided")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Volume Capability must be provided")
	}
	if err := ns.mounter.EnsureFolder(targetPath); err != nil {
		glog.Errorf("")
		return nil, status.Error(codes.Internal, err.Error())
	}
	mounted, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if mounted {
		glog.Errorf("already mounted, target path=%+s", targetPath)
		return nil, status.Error(codes.Internal, "path already mounted")
	}
	fsType := req.VolumeCapability.GetMount().FsType
	opts := []string{}
	opts = append(opts, "bind")
	if req.Readonly {
		opts = append(opts, "ro")
	}
	if len(fsType) == 0 {
		fsType = "ext4"
	}
	if err := ns.mounter.Mount(source, targetPath, fsType, opts...); err != nil {
		glog.Errorf("mount from %s to %s error, err=%v", source, targetPath, err)
		return nil, status.Error(codes.Internal, "mount to target error")
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	targetPath := req.GetTargetPath()
	mounted, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		glog.Errorf("check target path mounted error, err=%+v", err)
		return nil, status.Error(codes.Internal, "check target path error")
	}
	if !mounted {
		glog.Infof("target path not mounted, path=%s", targetPath)
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	if err := ns.mounter.Umount(targetPath); err != nil {
		glog.Errorf("umount error, target path=%s,err=%+v", targetPath, err)
		return nil, status.Error(codes.Internal, "umount error")
	}
	glog.Infof("umount path %s success", targetPath)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	glog.Infof("NodeStageVolume: stage disk %s, taget path: %s", req.GetVolumeId(), req.StagingTargetPath)
	targetPath := req.StagingTargetPath
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume: no volumeId is provided")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume Capability must be provided")
	}
	if err := ns.mounter.EnsureFolder(targetPath); err != nil {
		return nil, status.Errorf(codes.Internal, "NodeStageVolume: can't mkdir targetPath: %s", targetPath)
	}
	mounted, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodeStageVolume: check targetPath mounted error, err= %+v", err)
	}
	if mounted {
		glog.Errorf("NodeStageVolume: %s is already mounted", targetPath)
		return nil, status.Errorf(codes.Internal, "NodeStageVolume: targetPath already mounted")
	}
	glog.Infof("NodeStageVolume:find vol, id=%s", req.VolumeId)
	vol, ok := volumes[req.VolumeId]
	if !ok {
		glog.Errorf("NodeStageVolume: can't find %s in the lvmVols", req.GetVolumeId())
		return nil, status.Error(codes.Internal, "NodeStageVolume: can't find the requiested lvmVol")
	}
	devicePath := vol.DevicePath
	fsType := req.VolumeCapability.GetMount().GetFsType()
	if len(fsType) == 0 {
		fsType = "ext4"
	}
	options := []string{}
	err = ns.mounter.Format(devicePath, fsType)
	if err != nil {
		glog.Errorf("NodeStageVolume: format targetPath error, err= %+v", err)
		return nil, status.Errorf(codes.Internal, "NodeStageVolume: format targetPath error, err= %+v", err)
	}
	err = ns.mounter.Mount(devicePath, targetPath, fsType, options...)
	if err != nil {
		glog.Errorf("NodeStageVolume: mount targetPath error, err= %+v", err)
		return nil, status.Errorf(codes.Internal, "NodeStageVolume: mount targetPath error, err= %+v", err)
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	targetPath := req.GetStagingTargetPath()
	glog.Infof("NodeUnstageVolume: Starting to unstage volume,target %s", targetPath)
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume: no VolumeID provided")
	}
	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume: no target path is provided")
	}
	mounted, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		glog.Errorf("check is mounted error, err=%+v, path=%s", err, targetPath)
		return nil, status.Error(codes.InvalidArgument, "NodeUnstageVolume: check is mounted error")
	}
	if !mounted {
		glog.Infof("target path not mounted, might have error, path=%s", targetPath)
		return &csi.NodeUnstageVolumeResponse{}, nil
	}
	err = ns.mounter.Umount(targetPath)
	if err != nil {
		glog.Errorf("umount target path error, path=%s, err=%+v", targetPath, err)
		return nil, status.Error(codes.Internal, "NodeUnstageVolume: umount error")
	}
	glog.Infof("NodeStageVolume: success unstage volume")
	return &csi.NodeUnstageVolumeResponse{}, nil
}
