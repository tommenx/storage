package rpc

import (
	"context"
	"errors"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/consts"
	"google.golang.org/grpc"
)

var cli cdpb.CoordinatorClient

func Init(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		glog.Errorf("create grpc conn error, err=%+v", err)
		panic(err)
	}
	cli = cdpb.NewCoordinatorClient(conn)
}

func PutNodeStorage(ctx context.Context, node, kind string, storage []*cdpb.Storage) error {
	req := &cdpb.PutNodeStorageRequest{
		Base: &base.Base{},
		Node: &cdpb.Node{},
	}
	req.Name = node
	req.Kind = kind
	req.Node.Storage = storage
	rsp, err := cli.PutNodeStorage(ctx, req)
	if err != nil {
		glog.Errorf("call PutNodeStorage error, err=%+v", err)
		return err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error,code=%v,msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return errors.New("remote server error")
	}
	return nil
}

func GetNodeStorage(ctx context.Context, node, kind string) (map[string]*cdpb.Node, error) {
	req := &cdpb.GetNodeStorageRequest{
		Base: &base.Base{},
	}
	req.Kind = kind
	req.Node = node
	rsp, err := cli.GetNodeStorage(ctx, req)
	if err != nil {
		glog.Errorf("call get node storage error, err=%+v", err)
		return nil, err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, coede=%v, msg=%+v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return nil, errors.New("remote server get node storage error")
	}
	return rsp.Nodes, nil
}

func PutPodResource(ctx context.Context, basic map[string]string, request map[string]int64) error {
	req := &cdpb.PutPodResourceRequest{}
	pod := &cdpb.PodResource{}
	if name, ok := basic["name"]; ok {
		pod.Name = name
	}
	if namespace, ok := basic["namespace"]; ok {
		pod.Namespace = namespace
	}
	if dockerId, ok := basic["docker_id"]; ok {
		pod.DockerId = dockerId
	}
	if cgroupPath, ok := basic["cgroup_path"]; ok {
		pod.CgroupPath = cgroupPath
	}
	pod.RequestResource = request
	req.Pod = pod
	rsp, err := cli.PutPodResource(ctx, req)
	if err != nil {
		glog.Errorf("call put pod resource error, err=%+v", err)
		return err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return errors.New("remote server put pod resource error")
	}
	return nil
}

func DirectPutPodResource(ctx context.Context, pod *cdpb.PodResource, op int32) error {
	req := &cdpb.PutPodResourceRequest{}
	req.Pod = pod
	req.Operation = op
	rsp, err := cli.PutPodResource(ctx, req)
	if err != nil {
		glog.Errorf("call put pod resource error, err=%+v", err)
		return err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return errors.New("remote server put pod resource error")
	}
	return nil
}

//TODO
// return value error
func GetPodResource(ctx context.Context, ns, name string) (*cdpb.PodResource, error) {
	req := &cdpb.GetPodResourceRequest{
		Namespace: ns,
		Pod:       name,
	}
	rsp, err := cli.GetPodResource(ctx, req)
	if err != nil {
		glog.Errorf("call get pod resource error, err=%+v", err)
		return nil, err
	}
	if rsp.BaseResp.Code == consts.CodeNotExisted {
		glog.Infof("%s/%s do not exist in store", ns, name)
		return nil, consts.ErrNotExist
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return nil, errors.New("remote server get pod resource error")
	}
	return rsp.Pod, nil
}

func PutVolume(ctx context.Context, pv string, volume *cdpb.Volume) error {
	req := &cdpb.PutVolumeRequest{
		Base:   &base.Base{},
		Volume: volume,
		Pv:     pv,
	}
	rsp, err := cli.PutVolume(ctx, req)
	if err != nil {
		glog.Errorf("call put volume error, err=%+v", err)
		return err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return errors.New("remote server put volume error")
	}
	return nil

}

func GetVolume(ctx context.Context, ns, pvc string) (*cdpb.Volume, error) {
	req := &cdpb.GetVolumeRequest{
		Base:      &base.Base{},
		Namespace: ns,
		Name:      pvc,
	}
	rsp, err := cli.GetVolume(ctx, req)
	if err != nil {
		glog.Errorf("call get volume error, err=%+v", err)
		return nil, err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return nil, errors.New("remote server get volume error")
	}
	return rsp.Volume, nil
}

func GetAlivePod(ctx context.Context, kind string) (map[string]string, error) {
	req := &cdpb.GetAlivePodRequest{
		Base: &base.Base{},
		Kind: kind,
	}
	resp, err := cli.GetAlivePod(ctx, req)
	if err != nil {
		glog.Errorf("call get alive pod error, err=%+v", err)
		return nil, err
	}
	if resp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", resp.BaseResp.Code, resp.BaseResp.Message)
		return nil, errors.New("remote server get volume error")
	}
	return resp.Info, nil
}

func PutStorageUtil(ctx context.Context, info map[string]string) error {
	req := &cdpb.PutStorageUtilRequest{
		Base: &base.Base{},
		Info: info,
	}
	resp, err := cli.PutStorageUtil(ctx, req)
	if err != nil {
		glog.Errorf("call PutStorageUtil error, err=%+v", err)
		return err
	}
	if resp.BaseResp.Code != 0 {
		glog.Errorf("remote server error, code=%d, msg=%v", resp.BaseResp.Code, resp.BaseResp.Message)
		return err
	}
	return nil
}
