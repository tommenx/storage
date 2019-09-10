package server

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/consts"
	"time"
)

//TODO
//put node update
func (s *server) putNodeStorage(ctx context.Context, nodeName, kind string, node *cdpb.Node) error {
	storages := node.Storage
	for _, storage := range storages {
		device := storage.Name
		level := storage.Level
		val, err := proto.Marshal(storage)
		if err != nil {
			glog.Errorf("marshal storage error,device=%s,err=%+v", device, err)
			return err
		}
		err = s.db.PutNodeResource(ctx, nodeName, kind, level, device, val)
		if err != nil {
			glog.Errorf("etcd put storage %s/%s error, err=%+v", node, device, err)
			return err
		}
	}
	return nil

}
func (s *server) PutNodeStorage(ctx context.Context, req *cdpb.PutNodeStorageRequest) (*cdpb.PutNodeStorageResponse, error) {
	rsp := &cdpb.PutNodeStorageResponse{
		BaseResp: &base.BaseResp{},
	}
	node := req.Name
	kind := req.Kind
	err := s.putNodeStorage(ctx, node, kind, req.Node)
	if err != nil {
		rsp.BaseResp.Code = consts.CodeEtcdErr
		rsp.BaseResp.Message = "internal error"
		return rsp, nil
	}
	rsp.BaseResp.Code = consts.CodeOK
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

//TODO
//最好后期改为map[node-name]map[level]*storage
func (s *server) getNodeStorage(ctx context.Context, node, kind string) (map[string]*cdpb.Node, error) {
	nodeMap := make(map[string]*cdpb.Node)
	vals, err := s.db.GetNodeResource(ctx, node, kind)
	if err != nil {
		glog.Errorf("etcd get node storage info error, err=%+v", err)
		return nil, err
	}
	for k, val := range vals {
		storages := []*cdpb.Storage{}
		storage := &cdpb.Storage{}
		err = proto.Unmarshal(val, storage)
		if err != nil {
			glog.Errorf("unmarshal node storage error, err=%+v", err)
			return nil, err
		}
		storages = append(storages, storage)
		node := &cdpb.Node{}
		node.Storage = storages
		nodeMap[k] = node
	}
	return nodeMap, nil

}

func (s *server) GetNodeStorage(ctx context.Context, req *cdpb.GetNodeStorageRequest) (*cdpb.GetNodeStorageResponse, error) {
	rsp := &cdpb.GetNodeStorageResponse{
		BaseResp: &base.BaseResp{},
	}
	nodeMap, err := s.getNodeStorage(ctx, req.Node, req.Kind)
	if err != nil {
		rsp.BaseResp.Code = consts.CodeEtcdErr
		rsp.BaseResp.Message = "internal error"
		return rsp, nil
	}
	rsp.Nodes = nodeMap
	rsp.BaseResp.Code = consts.CodeOK
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) putPodResource(ctx context.Context, ns, podName string, pod *cdpb.PodResource) error {
	val, err := proto.Marshal(pod)
	if err != nil {
		glog.Errorf("marshal node error, pod_name=%s, err=%+v", pod, err)
		return err
	}
	err = s.db.PutPod(ctx, ns, podName, val)
	if err != nil {
		glog.Errorf("etcd put pod info error, err=%+v", err)
		return err
	}
	return nil
}

func (s *server) PutPodResource(ctx context.Context, req *cdpb.PutPodResourceRequest) (*cdpb.PutPodResourceResponse, error) {
	rsp := &cdpb.PutPodResourceResponse{
		BaseResp: &base.BaseResp{},
	}
	pod := req.Pod.Name
	namespace := req.Pod.Namespace
	op := req.Operation
	failed := false
	nodeResult := &cdpb.Node{}
	nodename := ""
	if op == consts.OpDel {
		//获取pod的情况
		//为所在节点的Allocation添加资源
		//返回结果
		info, err := s.getPodResource(ctx, pod, namespace)
		if err != nil {
			glog.Errorf("get %s/%s error, err= %+v", namespace, pod, err)
			failed = true
		}
		nodename = info.Node
		nodes, err := s.getNodeStorage(ctx, nodename, consts.KindAllocation)
		if err != nil {
			glog.Errorf("get %s resource error, err=%+v", nodename, err)
			failed = true
		}
		nodeResult = nodes[nodename]
		for index, storage := range nodeResult.Storage {
			if storage.Level == info.Level {
				for k, request := range info.RequestResource {
					nodeResult.Storage[index].Resource[k] += request
				}
			}
		}
	} else if op == consts.OpAdd {
		nodename = req.Pod.Node
		glog.Infof("add pod, node is %+v", nodename)
		nodes, err := s.getNodeStorage(ctx, nodename, consts.KindAllocation)
		if err != nil {
			glog.Errorf("get %s resource error, err=%+v", nodename, err)
			failed = true
		}
		glog.Infof("req pod info is %+v", *req.Pod)
		if !failed {
			if node, ok := nodes[nodename]; ok {
				for index, storage := range node.Storage {
					if storage.Level == req.Pod.Level {
						for key, request := range req.Pod.RequestResource {
							node.Storage[index].Resource[key] -= request
						}
					}
				}
				nodeResult = node
			} else {
				glog.Infof("can't get node %s info", nodename)
			}
		}
	}
	if failed {
		glog.Errorf("cal resource error,node=%s", nodename)
		rsp.BaseResp.Code = consts.CodeInternalErr
		rsp.BaseResp.Message = "calculate resource error"
		return rsp, nil
	}
	err := s.putNodeStorage(ctx, nodename, consts.KindAllocation, nodeResult)
	if err != nil {
		glog.Errorf("put node storage error, err=%+v", err)
		rsp.BaseResp.Code = consts.CodeEtcdErr
		rsp.BaseResp.Message = "put node storage error"
		return rsp, nil
	}
	if op == consts.OpAdd {
		err = s.putPodResource(ctx, namespace, pod, req.Pod)
	} else {
		err = s.db.DelPod(ctx, namespace, pod)
	}
	if err != nil {
		rsp.BaseResp.Code = consts.CodeInternalErr
		rsp.BaseResp.Message = "internal error"
		return rsp, nil
	}
	rsp.BaseResp.Code = consts.CodeOK
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) GetPodResource(ctx context.Context, req *cdpb.GetPodResourceRequest) (*cdpb.GetPodResourceResponse, error) {
	rsp := &cdpb.GetPodResourceResponse{
		BaseResp: &base.BaseResp{},
	}
	namespace := req.Namespace
	podName := req.Pod
	pod, err := s.getPodResource(ctx, podName, namespace)
	if err == consts.ErrNotExist {
		rsp.BaseResp.Code = consts.CodeNotExisted
		rsp.BaseResp.Message = "pod do not exist in store"
		return rsp, nil
	}
	if err != nil {
		rsp.BaseResp.Code = consts.CodeMarshalErr
		rsp.BaseResp.Message = "unmarshal pod error"
		return rsp, nil
	}
	rsp.Pod = pod
	rsp.BaseResp.Code = consts.CodeOK
	rsp.BaseResp.Message = "success"
	return rsp, nil
}

func (s *server) getPodResource(ctx context.Context, podName, namespace string) (*cdpb.PodResource, error) {
	val, err := s.db.GetPod(ctx, namespace, podName)
	if err != nil {
		return nil, err
	}
	pod := &cdpb.PodResource{}
	err = proto.UnmarshalMerge(val, pod)
	if err != nil {
		glog.Errorf("unmarshal pod error, err=%+v", err)
		return nil, err
	}
	return pod, nil
}

func (s *server) PutVolume(ctx context.Context, req *cdpb.PutVolumeRequest) (*cdpb.PutVolumeResponse, error) {
	rsp := &cdpb.PutVolumeResponse{
		BaseResp: &base.BaseResp{},
	}
	pvName := req.Pv
	namespace, pvc, err := s.pv.GetPVCByPV(pvName)
	retry := 5
	for err == consts.ErrNotBound {
		if retry == 0 {
			break
		}
		namespace, pvc, err = s.pv.GetPVCByPV(pvName)
		time.Sleep(1 * time.Second)
		retry--
	}
	if err != nil {
		glog.Errorf("get bounded pvc error,pv name=%s, err=%+v", pvName, err)
		rsp.BaseResp.Code = consts.CodeNotFound
		rsp.BaseResp.Message = "get pvc by pv error"
		return rsp, nil
	}
	val, err := proto.Marshal(req.Volume)
	if err != nil {
		glog.Errorf("marshal volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = consts.CodeMarshalErr
		rsp.BaseResp.Message = "marshal volume error"
		return rsp, nil
	}
	err = s.db.PutPVC(ctx, namespace, pvc, val)
	if err != nil {
		glog.Errorf("put volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = consts.CodeEtcdErr
		rsp.BaseResp.Message = "put volume error"
		return rsp, nil
	}
	rsp.BaseResp.Code = consts.CodeOK
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
		rsp.BaseResp.Code = consts.CodeEtcdErr
		rsp.BaseResp.Message = "get volume error"
		return rsp, nil
	}
	volume := &cdpb.Volume{}
	err = proto.Unmarshal(val, volume)
	if err != nil {
		glog.Errorf("unmarshal volume error, pvc=%v, err=%+v", pvc, err)
		rsp.BaseResp.Code = consts.CodeMarshalErr
		rsp.BaseResp.Message = "unmarshal volume error"
		return rsp, nil
	}
	rsp.Volume = volume
	rsp.BaseResp.Code = consts.CodeOK
	rsp.BaseResp.Message = "success"
	return rsp, nil
}
