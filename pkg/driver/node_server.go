package driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
	mounter      Mounter
	pvController controller.PVController
}

func NewNodeServer(d *csicommon.CSIDriver, containerized bool, pvController controller.PVController) (*nodeServer, error) {
	mounter := NewMounter("")
	if containerized {
		mounter = NewMounter(prefix)
	}
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
		mounter:           mounter,
		pvController:      pvController,
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

func (ns *nodeServer) createLogicalVolume(volumeId string, volumeGroup string) (*LvmVolume, error) {
	pv, err := ns.pvController.GetPV(volumeId)
	if err != nil {
		glog.Errorf("get volume error, err=%+v", err)
		return nil, err
	}
	pvQuantity := pv.Spec.Capacity["storage"]
	pvSize := pvQuantity.Value()
	volume := &LvmVolume{
		PVName: volumeId,
		Size:   pvSize,
	}
	volume.VolumeGroup = volumeGroup
	err = volume.Create(prefix)
	if err != nil {
		glog.Errorf("create logical volume error, err=%+v", err)
		return nil, err
	}
	maj, min, ok := GetDeviceNum(volumeGroup, volumeId, prefix)
	glog.Infof("maj=%s, min=%s, ok=%v", maj, min, ok)
	if !ok {
		glog.Errorf("get device number error")
		return nil, errors.New("get device number error")
	}
	volume.Maj = maj
	volume.Min = min
	return volume, nil
}

func parseExistLogicalVolume(volumeId string, volumeGroup string, prefix string) (*LvmVolume, error) {
	var (
		deviceNumberStr string
		sizeStr         string
		mountPointStr   string
	)
	lsblkCmd := fmt.Sprintf("%s %s", prefix, "lsblk")
	afterVolumeId := strings.ReplaceAll(volumeId, "-", "--")
	label := fmt.Sprintf("%s-%s", volumeGroup, afterVolumeId)
	args := []string{`--output`, `NAME,MAJ:MIN,SIZE,MOUNTPOINT`}
	cmd := GetCmd(lsblkCmd, args)
	out, err := Run(cmd)
	if err != nil {
		return nil, errors.New("get device info error")
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if ok := strings.Contains(line, label); ok {
			fmt.Println(line)
			cols := strings.Fields(line)
			mountPointStr = cols[len(cols)-1]
			sizeStr = cols[len(cols)-2]
			deviceNumberStr = cols[len(cols)-3]
			break
		}
	}
	if len(deviceNumberStr)+len(sizeStr)+len(mountPointStr) == 0 {
		glog.Errorf("do not extract any info")
		return nil, errors.New("can't extract any info")
	}
	deviceNumbers := strings.Split(deviceNumberStr, ":")
	sizeStr = sizeStr[:len(sizeStr)-1]
	size := utils.Int64(sizeStr) * int64(GB)
	return &LvmVolume{
		PVName:      volumeId,
		DevicePath:  fmt.Sprintf("/dev/%s/%s", volumeGroup, volumeId),
		VolumeGroup: volumeGroup,
		Maj:         deviceNumbers[0],
		Min:         deviceNumbers[1],
		Size:        size,
	}, nil
}

func checkParameter(req *csi.NodePublishVolumeRequest, volumeGroup *string) error {
	if !strings.HasSuffix(req.TargetPath, "/mount") {
		return status.Errorf(codes.InvalidArgument, "malformed the value of target path: %s", req.TargetPath)
	}
	if req.VolumeId == "" {
		return status.Error(codes.InvalidArgument, "NodePublishVolume: Volume ID must be provided")
	}
	if req.StagingTargetPath == "" {
		return status.Error(codes.InvalidArgument, "NodePublishVolume: Staging Target Path must be provided")
	}
	if req.VolumeCapability == nil {
		return status.Error(codes.InvalidArgument, "NodePublishVolume: Volume Capability must be provided")
	}
	vg, ok := req.VolumeContext["kind"]
	if !ok {
		return status.Errorf(codes.InvalidArgument, "NodePublishVolume: Volume Kind must be provided")
	}
	*volumeGroup = vg
	return nil
}

func checkFsType(devicePath string) (string, error) {
	output, err := exec.Command("file", "-bsL", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "data" {
		return "", nil
	}
	output, err = exec.Command("blkid", "-c", "/dev/null", "-o", "export", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	parseErr := errors.New("cannot parse output of blkid.")
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "=")
		if len(fields) != 2 {
			return "", parseErr
		}
		if fields[0] == "TYPE" {
			return fields[1], nil
		}
	}
	return "", parseErr
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	glog.Infof("NodePublishVolume: start to mount to target path")
	targetPath := req.TargetPath
	var volumeGroup string
	if err := checkParameter(req, &volumeGroup); err != nil {
		glog.Errorf("check parameter failed")
		return nil, err
	}
	volumeId := req.GetVolumeId()
	devicePath := filepath.Join("/dev/", volumeGroup, volumeId)
	//确认设备是否已经创建
	vol := &LvmVolume{}
	if _, err := os.Stat(devicePath); os.IsNotExist(err) {
		vol, err = ns.createLogicalVolume(volumeId, volumeGroup)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else if err != nil {
		glog.Errorf("NodePublishVolume: state device path error")
		return nil, status.Error(codes.Internal, err.Error())
	}
	//确认目标地址是否已经创建或被挂载
	mounted, err := ns.mounter.IsMounted(targetPath)
	if err != nil {
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			mounted = false
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if mounted {
		glog.Errorf("already mounted, target path=%+s", targetPath)
		return nil, status.Error(codes.Internal, "path already mounted")
	}
	//确认文件系统
	existFSType, err := checkFsType(devicePath)
	if existFSType == "" {
		glog.Infof("device has no filesystem, format it ")
		if err := ns.mounter.Format(devicePath, "ext4"); err != nil {
			glog.Errorf("format filesystem error, err=%+v", err)
			return nil, status.Error(codes.Internal, "format filesystem error")
		}
	}
	//挂载设备至目录
	opts := []string{}
	if req.Readonly {
		opts = append(opts, "ro")
	}
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	opts = append(opts, mountFlags...)
	if err := ns.mounter.Mount(devicePath, targetPath, "ext4", opts...); err != nil {
		glog.Errorf("mount from %s to %s error, err=%v", devicePath, targetPath, err)
		return nil, status.Error(codes.Internal, "mount to target error")
	}
	vol, _ = parseExistLogicalVolume(volumeId, volumeGroup, "")
	glog.Infof("%+v", vol)
	rpcVolume := &cdpb.Volume{
		Name:          vol.PVName,
		VolumeGroup:   vol.VolumeGroup,
		Maj:           vol.Maj,
		Min:           vol.Min,
		LogicalVolume: volumeId,
		Size:          int32(vol.GetFormatSize()),
	}
	if err := rpc.PutVolume(ctx, volumeId, rpcVolume); err != nil {
		glog.Errorf("rpc put volume error, err=%+v", err)
		return nil, status.Error(codes.Internal, "rpc put volume error")
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
	glog.Infof("NodeStageVolume: do not need yet")
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	glog.Infof("NodeUnstageVolume: do not need yet ")
	return &csi.NodeUnstageVolumeResponse{}, nil
}
