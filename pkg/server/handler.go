package server

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/consts"
	"strings"
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
			glog.Errorf("cal resource error,node=%s", nodename)
			rsp.BaseResp.Code = consts.CodeInternalErr
			rsp.BaseResp.Message = "calculate resource error"
			return rsp, nil
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
	var errCheck error
	var errBound error
	if op == consts.OpAdd {
		err = s.putPodResource(ctx, namespace, pod, req.Pod)
		//	TODO 这里做简单处理，将pvc和pod的名字强绑定
		//	TODO pod的名字:tidb-cluster-tikv-0，pvc则为: tikv-tidb-cluster-tikv-0
		//	TODO 以后再慢慢改吧
		volumeName, _ := s.pvc.GetVolumeName(namespace, fmt.Sprintf("tikv-%s", pod))
		err = s.db.Put(ctx, fmt.Sprintf("%s%s", consts.KeyBounded, pod), []byte(volumeName))

	} else {
		err = s.db.DelPod(ctx, namespace, pod)
		errCheck = s.db.Del(ctx, fmt.Sprintf("%s%s", consts.KeyCheck, pod))
		errBound = s.db.Del(ctx, fmt.Sprintf("%s%s", consts.KeyBounded, pod))
	}
	if err != nil || errBound != nil || errCheck != nil {
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

func (s *server) GetAlivePod(ctx context.Context, req *cdpb.GetAlivePodRequest) (*cdpb.GetAlivePodResponse, error) {
	resp := &cdpb.GetAlivePodResponse{
		BaseResp: &base.BaseResp{},
	}
	var prefix string
	if req.Kind == "bounded" {
		prefix = consts.KeyBounded
	} else {
		prefix = consts.KeyCheck
	}
	data, err := s.db.Get(ctx, prefix, true)
	info := make(map[string]string)
	if err == consts.ErrNotExist {
		resp.Info = info
		resp.BaseResp.Code = 0
		resp.BaseResp.Message = "success"
		return resp, nil
	}
	if err != nil {
		glog.Errorf("get prefix error,err=%v", err)
		resp.BaseResp.Code = 1
		resp.BaseResp.Message = "etcd get data error"
		return resp, nil
	}

	for str, val := range data {
		key := extractKey(str)
		info[key] = string(val)
	}
	resp.Info = info
	resp.BaseResp.Code = 0
	resp.BaseResp.Message = "success"
	return resp, nil
}

func extractKey(key string) string {
	return key[strings.LastIndex(key, "/")+1:]
}

func (s *server) PutStorageUtil(ctx context.Context, req *cdpb.PutStorageUtilRequest) (*cdpb.PutStorageUtilResponse, error) {
	resp := &cdpb.PutStorageUtilResponse{
		BaseResp: &base.BaseResp{},
	}
	// /storage/check/pod-name 表示当前tikv使用存储的情况
	// /storage/log/pod-name/time, val 表示与当前用量 表示pod-name 磁盘使用的历史记录
	for podName, storageUtil := range req.Info {
		err1 := s.db.Put(ctx, fmt.Sprintf("%s%s", consts.KeyCheck, podName), []byte(storageUtil))
		if err1 != nil {
			glog.Errorf("put pod %s check error, err1=%v", podName, err1)
			resp.BaseResp.Code = 1
			resp.BaseResp.Message = "put pod check error"
			return resp, nil
		}
	}
	resp.BaseResp.Code = 0
	resp.BaseResp.Message = "success"
	return resp, nil
}

func (s *server) PutInstanceLimit(ctx context.Context, req *cdpb.PutInstanceLimitRequest) (*cdpb.PutInstanceLimitResponse, error) {
	resp := &cdpb.PutInstanceLimitResponse{
		BaseResp: &base.BaseResp{},
	}
	err := s.db.Put(ctx, fmt.Sprintf("%s%s", consts.KeyLimit, req.Name), []byte(req.Val))
	if err != nil {
		glog.Errorf("put instance %s error, err=%+v", req.Name, err)
		resp.BaseResp.Code = 1
		resp.BaseResp.Message = "put instance error"
		return resp, nil
	}
	glog.Infof("set instance limit success")
	resp.BaseResp.Code = 0
	resp.BaseResp.Message = "success"
	return resp, nil

}
