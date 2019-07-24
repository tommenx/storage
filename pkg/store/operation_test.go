package store

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/tommenx/cdproto/cdpb"
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

func TestGet(t *testing.T) {
	store := NewEtcd([]string{"127.0.0.1:2379"})
	ctx := context.Background()
	val, err := store.GetPVC(ctx, "aaa", "ccc")
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("%+v", val)
}

func TestPutPodResource(t *testing.T) {
	etcd := NewEtcd([]string{"127.0.0.1:2379"})
	ctx := context.Background()
	ns := "default"
	name := "test-pod-5"
	pod := &cdpb.PodResource{
		Name:      name,
		Namespace: ns,
		Node:      "localhost.localdomain",
		DockerId:  "71f9210c75c97dceb1050c243ed580d7a784cec89ff67169b52a14fa309bed75",
	}
	val, err := proto.Marshal(pod)
	if err != nil {
		t.Errorf("marshal pod resource error, err=%+v", err)
		return
	}
	err = etcd.PutPod(ctx, ns, name, val)
	if err != nil {
		t.Errorf("put pod resource error, err=%+v", err)
		return
	}
	t.Logf("success")
}

func TestGetPodResource(t *testing.T) {
	etcd := NewEtcd([]string{"127.0.0.1:2379"})
	ctx := context.Background()
	ns := "default"
	name := "test-pod-5"
	val, err := etcd.GetPod(ctx, ns, name)
	if err != nil {
		t.Errorf("get pod resource error, err=%+v", err)
		return
	}
	pod := &cdpb.PodResource{}
	err = proto.Unmarshal(val, pod)
	if err != nil {
		t.Errorf("unmarshal pod resource error, err=%+v", err)
		return
	}
	t.Logf("%+v", pod)
}

func TestPutVolume(t *testing.T) {
	etcd := NewEtcd([]string{"127.0.0.1:2379"})
	ctx := context.Background()
	ns := "default"
	pvc := "lvm-pvc"
	volume := &cdpb.Volume{
		Maj: "253",
		Min: "6",
	}
	val, err := proto.Marshal(volume)
	if err != nil {
		t.Errorf("marshal volume resource error, err=%+v", err)
		return
	}
	err = etcd.PutPVC(ctx, ns, pvc, val)
	if err != nil {
		t.Errorf("put volume error, err=%+v", err)
	}
}

func TestGetVolume(t *testing.T) {
	etcd := NewEtcd([]string{"127.0.0.1:2389"})
	ctx := context.Background()
	ns := "default"
	name := "lvm-pvc-4"
	val, err := etcd.GetPVC(ctx, ns, name)
	if err != nil {
		t.Errorf("get pod resource error, err=%+v", err)
		return
	}
	vol := &cdpb.Volume{}
	err = proto.Unmarshal(val, vol)
	if err != nil {
		t.Errorf("unmarshal pod resource error, err=%+v", err)
		return
	}
	t.Logf("%+v", vol)
}
