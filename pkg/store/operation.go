package store

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/consts"
	v3 "go.etcd.io/etcd/clientv3"
	"strings"
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
	Put(ctx context.Context, key string, val []byte) error
	Get(ctx context.Context, key string, prefix bool) (map[string][]byte, error)
	Del(ctx context.Context, key string) error
	PutNodeResource(ctx context.Context, node, kind, level, device string, val []byte) error
	GetNodeResource(ctx context.Context, node, kind string) (map[string][]byte, error)
	PutPVC(ctx context.Context, ns, pvc string, val []byte) error
	GetPVC(ctx context.Context, ns, pvc string) ([]byte, error)
	PutPod(ctx context.Context, ns, name string, val []byte) error
	GetPod(ctx context.Context, ns, name string) ([]byte, error)
	DelPod(ctx context.Context, ns, name string) error
	GetAlivePodInfo(ctx context.Context, kind string) (map[string]string, error)
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

func (h *EtcdHandler) Del(ctx context.Context, key string) error {
	_, err := h.client.Delete(ctx, key)
	if err != nil {
		glog.Errorf("delete key %s error,err=%+v", key, err)
		return err
	}
	return nil
}

func (h *EtcdHandler) PutNodeResource(ctx context.Context, node, kind, level, device string, val []byte) error {
	key := getKey(prefixNode, node, kind, level, device)
	fmt.Println("node resource key is ", key)
	return h.Put(ctx, key, val)

}

func (h *EtcdHandler) GetNodeResource(ctx context.Context, node, kind string) (map[string][]byte, error) {
	key := ""
	if node == "all" {
		key = prefixNode
	} else {
		key = getKey(prefixNode, node)
	}
	fmt.Printf("key is %+v", key)
	fmt.Println("node", node)
	fmt.Println("kind", kind)
	kvs, err := h.Get(ctx, key, true)
	if err != nil {
		return nil, err
	}
	infos := make(map[string][]byte)
	for k, v := range kvs {
		fields := strings.FieldsFunc(k, func(r rune) bool {
			if r == '/' {
				return true
			}
			return false
		})
		if fields[3] == kind {
			if fields[2] == node || node == "all" {
				infos[fields[2]] = v
			}
		}
	}
	return infos, nil
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

func (h *EtcdHandler) DelPod(ctx context.Context, ns, name string) error {
	key := getKey(prefixPod, ns, name)
	err := h.Del(ctx, key)
	if err != nil {
		glog.Errorf("etcd del %s error, err=%+v", key, err)
		return err
	}
	return nil
}

func (h *EtcdHandler) GetAlivePodInfo(ctx context.Context, kind string) (map[string]string, error) {
	var prefix string
	if kind == "bounded" {
		prefix = consts.KeyBounded
	} else {
		prefix = consts.KeyCheck
	}
	data, err := h.Get(ctx, prefix, true)
	if err != nil {
		glog.Errorf("get prefix %s error,err=%v", err)
		return nil, err
	}
	info := make(map[string]string)
	for str, val := range data {
		key := extractKey(str)
		info[key] = string(val)
	}
	return info, nil
}

func extractKey(key string) string {
	return key[strings.LastIndex(key, "/")+1:]
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
