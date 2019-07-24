package server

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/controller"
	"time"
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
	rsp.NodeMap = nodeMap
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) PutPodResource(ctx context.Context, req *cdpb.PutPodResourceRequest) (*cdpb.PutPodResourceResponse, error) {
	rsp := &cdpb.PutPodResourceResponse{
		BaseResp: &base.BaseResp{},
	}
	pod := req.Pod.Name
	namespace := req.Pod.Namespace
	val, err := proto.Marshal(req.Pod)
	if err != nil {
		glog.Errorf("marshal node error, pod_name=%s, err=%+v", pod, err)
		rsp.BaseResp.Code = 1
		rsp.BaseResp.Message = "marshal pod error"
		return rsp, nil
	}
	err = s.db.PutPod(ctx, namespace, pod, val)
	if err != nil {
		glog.Errorf("etcd put pod info error, err=%+v", err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "etcd put pod info error"
		return rsp, nil
	}
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) GetPodResource(ctx context.Context, req *cdpb.GetPodResourceRequest) (*cdpb.GetPodResourceResponse, error) {
	rsp := &cdpb.GetPodResourceResponse{
		BaseResp: &base.BaseResp{},
	}
	namespace := req.Namespace
	podName := req.Pod
	val, err := s.db.GetPod(ctx, namespace, podName)
	if err != nil {
		glog.Errorf("etcd get pod error, name=%s, err=%+v", podName, err)
		rsp.BaseResp.Code = 1
		rsp.BaseResp.Message = "etcd get pod error"
		return rsp, nil
	}
	pod := &cdpb.PodResource{}
	err = proto.UnmarshalMerge(val, pod)
	if err != nil {
		glog.Errorf("unmarshal pod error, err=%+v", err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "unmarshal pod error"
		return rsp, nil
	}
	rsp.Pod = pod
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

//TODO
//待确认能否在创建pod之前先创建pvc，并且能够绑定上
//否则就比较麻烦啦
func (s *server) PutVolume(ctx context.Context, req *cdpb.PutVolumeRequest) (*cdpb.PutVolumeResponse, error) {
	rsp := &cdpb.PutVolumeResponse{
		BaseResp: &base.BaseResp{},
	}
	pvName := req.Pv
	namespace, pvc, err := s.pv.GetPVCByPV(pvName)
	retry := 5
	for err == controller.ErrNotBound {
		if retry == 0 {
			break
		}
		namespace, pvc, err = s.pv.GetPVCByPV(pvName)
		time.Sleep(1 * time.Second)
		retry--
	}
	glog.Infof("retry count = %d", retry)
	glog.Errorf("now err is %+v", err)
	glog.Infof("pv info is %s %s", namespace, pvc)
	if err != nil {
		glog.Errorf("get bounded pvc error,pv name=%s, err=%+v", pvName, err)
		rsp.BaseResp.Code = 3
		rsp.BaseResp.Message = "get pvc by pv error"
		return rsp, nil
	}
	val, err := proto.Marshal(req.Volume)
	if err != nil {
		glog.Errorf("marshal volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = 1
		rsp.BaseResp.Message = "marshal volume error"
		return rsp, nil
	}
	err = s.db.PutPVC(ctx, namespace, pvc, val)
	if err != nil {
		glog.Errorf("put volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "put volume error"
		return rsp, nil
	}
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) GetVolume(ctx context.Context, req *cdpb.GetVolumeRequest) (*cdpb.GetVolumeResponse, error) {
	rsp := &cdpb.GetVolumeResponse{
		BaseResp: &base.BaseResp{},
	}
	namespace := req.Namespace
	pvc := req.Name
	val, err := s.db.GetPVC(ctx, namespace, pvc)
	if err != nil {
		glog.Errorf("get volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = 1
		rsp.BaseResp.Message = "get volume error"
		return rsp, nil
	}
	volume := &cdpb.Volume{}
	err = proto.Unmarshal(val, volume)
	if err != nil {
		glog.Errorf("unmarshal volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = 2
		rsp.BaseResp.Message = "unmarshal volume error"
		return rsp, nil
	}
	rsp.Volume = volume
	rsp.BaseResp.Code = 0
	rsp.BaseResp.Message = "success"
	return rsp, nil
}
