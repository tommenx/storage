package store

import (
	"context"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/consts"
	v3 "go.etcd.io/etcd/clientv3"
	"time"
)

// /storage/nodes/node-name
// /storage/pods/ns/pod-name
// /storage/pvcs/ns/pvc-name
var (
	DefaultTimeout = 5 * time.Second
	prefix         = "/storage/"
	prefixNode     = prefix + "nodes/"
	prefixPod      = prefix + "pods/"
	prefixPVC      = prefix + "pvcs/"
)

type EtcdHandler struct {
	client *v3.Client
}

type EtcdInterface interface {
	PutNodeResource(ctx context.Context, node, kind, level, device string, val []byte) error
	GetNodeList(ctx context.Context) (map[string][]byte, error)
	PutPVC(ctx context.Context, ns, pvc string, val []byte) error
	GetPVC(ctx context.Context, ns, pvc string) ([]byte, error)
	PutPod(ctx context.Context, ns, name string, val []byte) error
	GetPod(ctx context.Context, ns, name string) ([]byte, error)
}

func NewEtcd(endpoints []string) EtcdInterface {
	cli, err := v3.New(v3.Config{
		Endpoints:   endpoints,
		DialTimeout: DefaultTimeout,
	})
	if err != nil {
		glog.Errorf("create ETCD client error, err=%+v", err)
		panic(err)
	}
	return &EtcdHandler{
		client: cli,
	}
}

func (h *EtcdHandler) Put(ctx context.Context, key string, val []byte) error {
	_, err := h.client.Put(ctx, key, string(val))
	if err != nil {
		glog.Errorf("ETCD put key error, key=%v, err=%+v", key, err)
		return err
	}
	return nil
}

func (h *EtcdHandler) Get(ctx context.Context, key string, prefix bool) (map[string][]byte, error) {
	ops := []v3.OpOption{}
	if prefix {
		ops = append(ops, v3.WithPrefix())
	}
	rsp, err := h.client.Get(ctx, key, ops...)
	if err != nil {
		glog.Errorf("ETCD get key error, key=%v, err=%+v", key, err)
		return nil, err
	}
	kvs := make(map[string][]byte)
	for _, kv := range rsp.Kvs {
		path := string(kv.Key)
		kvs[path] = kv.Value
	}
	if len(kvs) == 0 {
		return nil, consts.ErrNotExist
	}
	return kvs, nil
}

func (h *EtcdHandler) PutNodeResource(ctx context.Context, node, kind, level, device string, val []byte) error {
	key := getKey(prefixNode, node, kind, level, device)
	glog.Infof("node resource key is %s", key)
	return h.Put(ctx, key, val)

}

func (h *EtcdHandler) GetNodeList(ctx context.Context) (map[string][]byte, error) {
	kvs, err := h.Get(ctx, prefixNode, true)
	if err != nil {
		return nil, err
	}
	return kvs, nil
}

func (h *EtcdHandler) PutPVC(ctx context.Context, ns, pvc string, val []byte) error {
	key := getKey(prefixPVC, ns, pvc)
	return h.Put(ctx, key, val)
}

func (h *EtcdHandler) GetPVC(ctx context.Context, ns, pvc string) ([]byte, error) {
	key := getKey(prefixPVC, ns, pvc)
	kvs, err := h.Get(ctx, key, false)
	if err != nil {
		return nil, err
	}
	return kvs[key], nil
}

func (h *EtcdHandler) PutPod(ctx context.Context, ns, name string, val []byte) error {
	key := getKey(prefixPod, ns, name)
	return h.Put(ctx, key, val)
}

func (h *EtcdHandler) GetPod(ctx context.Context, ns, name string) ([]byte, error) {
	key := getKey(prefixPod, ns, name)
	kvs, err := h.Get(ctx, key, false)
	if err != nil {
		return nil, err
	}
	return kvs[key], nil
}

func getKey(prefix string, args ...string) string {
	for i, arg := range args {
		prefix += arg
		if i < len(args)-1 {
			prefix += "/"
		}
	}
	return prefix
}
