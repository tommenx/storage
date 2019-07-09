package etcd

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	v3 "go.etcd.io/etcd/clientv3"
	"strings"
	"time"
)

var (
	DefaultTimeout = 5 * time.Second
	prefix         = "/storage/"
	prefixNode     = prefix + "nodes/"
	prefixPod      = prefix + "pods/"
)

type EtcdHandler struct {
	client *v3.Client
}

type EtcdInterface interface {
}

func NewEtcd(endpoints []string) *EtcdHandler {
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
	return kvs, nil
}

func (h *EtcdHandler) PutNode(ctx context.Context, node string, val []byte) error {
	key := getNodeKey(node)
	return h.Put(ctx, key, val)

}

func (h *EtcdHandler) GetNodeList(ctx context.Context) (map[string][]byte, error) {
	kvs, err := h.Get(ctx, prefixNode, true)
	if err != nil {
		return nil, err
	}
	res := make(map[string][]byte)
	for path, val := range kvs {
		index := strings.LastIndex(path, "/")
		key := path[index+1:]
		res[key] = val
	}
	return res, nil
}

func (h *EtcdHandler) PutPod(ctx context.Context, ns, name string, val []byte) error {
	key := getPodKey(ns, name)
	return h.Put(ctx, key, val)
}

func (h *EtcdHandler) GetPod(ctx context.Context, ns, name string) ([]byte, error) {
	key := getPodKey(ns, name)
	kvs, err := h.Get(ctx, key, false)
	if err != nil {
		return nil, err
	}
	return kvs[key], nil
}

func getPodKey(ns, name string) string {
	key := fmt.Sprintf("%s/%s/%s", prefixPod, ns, name)
	return key
}

func getNodeKey(node string) string {
	return prefixNode + node
}
