package etcd

import (
	"context"
	"testing"
)

func TestPutAndGet(t *testing.T) {
	endpoints := []string{
		"127.0.0.1:2379",
	}
	h := NewEtcd(endpoints)
	ctx := context.TODO()
	infos := map[string]string{
		"a": "aaaa",
		"b": "bbbb",
		"c": "cccc",
	}
	for k, v := range infos {
		err := h.PutPod(ctx, k, k, []byte(v))
		err = h.PutNode(ctx, k, []byte(v))
		if err != nil {
			t.Error(err)
		}
	}
	//get
	for k, v := range infos {
		val, _ := h.GetPod(ctx, k, k)
		if v != string(val) {
			t.Errorf("put %s, get %s", v, string(val))
		}
	}
	kvs, _ := h.GetNodeList(ctx)
	for k, v := range kvs {
		if string(v) != infos[k] {
			t.Errorf("put %s, get %s", infos[k], string(v))
		}
	}

}
