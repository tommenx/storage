package server

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
)

func (s *server) PutNodeStorage(ctx context.Context, req *cdpb.PutNodeStorageRequest) (*cdpb.PutNodeStorageResponse, error) {
	rsp := &cdpb.PutNodeStorageResponse{
		BaseResp: &base.BaseResp{},
	}
	nodeName := req.Node.NodeName
	val, err := proto.Marshal(req.Node)
	if err != nil {
		glog.Errorf("marshal node error, node_name=%s, err=%+v", nodeName, err)
		rsp.BaseResp.Code = 1
		rsp.BaseResp.Message = "marshal node error"
		return rsp, nil
	}
	err = s.db.PutNode(ctx, nodeName, val)
	if err != nil {
		glog.Errorf("etcd put node error, node_name=%s, err=%+v", nodeName, err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "etcd put node error"
		return rsp, nil
	}
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) GetNodeStorage(ctx context.Context, req *cdpb.GetNodeStorageRequest) (*cdpb.GetNodeStorageResponse, error) {
	rsp := &cdpb.GetNodeStorageResponse{
		BaseResp: &base.BaseResp{},
	}
	nodeMap := make(map[string]*cdpb.NodeStorage)
	vals, err := s.db.GetNodeList(ctx)
	if err != nil {
		glog.Errorf("etcd get node storage info error, err=%+v", err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "etcd get node list error"
		return rsp, nil
	}
	for node, val := range vals {
		storage := &cdpb.NodeStorage{}
		err = proto.Unmarshal(val, storage)
		if err != nil {
			glog.Errorf("unmarshal node storage error, err=%+v", err)
			rsp.BaseResp.Code = 2
			rsp.BaseResp.Message = "unmarshal node storage error"
			return rsp, nil
		}
		nodeMap[node] = storage
	}
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) PutPodResource(ctx context.Context, req *cdpb.PutPodResourceRequest) (*cdpb.PutPodResourceResponse, error) {
	return nil, nil
}

func (s *server) PutVolume(ctx context.Context, req *cdpb.PutVolumeRequest) (*cdpb.PutVolumeResponse, error) {
	return nil, nil
}

func (s *server) GetVolume(ctx context.Context, req *cdpb.GetVolumeRequest) (*cdpb.GetVolumeResponse, error) {
	return nil, nil
}
