package rpc

import (
	"context"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/consts"
	"testing"
)

func TestPutNodeStorage(t *testing.T) {
	Init(":50051")
	storage := []*cdpb.Storage{
		{
			Name:  "ssd2",
			Level: "SSD",
			Resource: map[string]int64{
				"read_bps_device":   100,
				"read_iops_device":  200,
				"write_bps_device":  300,
				"write_iops_device": 400,
				"space":             100,
			},
		},
	}
	err := PutNodeStorage(context.Background(), "nodeaaa", consts.KindAllocation, storage)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNodeStorage(t *testing.T) {
	Init(":50051")
	//node := "localhost.localdomain"
	node := "all"
	kinds := []string{consts.KindAllocation}
	for _, kind := range kinds {
		t.Logf("kind is %+v", kind)
		infos, err := GetNodeStorage(context.Background(), node, kind)
		if err != nil {
			t.Errorf("error is %+v", err)
			return
		}
		for k, v := range infos {
			t.Logf("node is %v", k)
			for _, storage := range v.Storage {
				t.Logf("%+v", storage)
			}
		}
	}
}

func TestPutPodResource(t *testing.T) {
	//add
	Init(":50051")

	pod := &cdpb.PodResource{
		Name:      "test-pod",
		Namespace: "default",
		Node:      "localhost.localdomain",
		RequestResource: map[string]int64{
			"read_bps_device":  10,
			"write_bps_device": 20,
		},
		Level: consts.LevelSSD,
	}
	err := DirectPutPodResource(context.Background(), pod, consts.OpAdd)
	if err != nil {
		t.Errorf("put pod resource error,err=%+v", err)
	}
}

func TestDirectPutPodResource(t *testing.T) {
	Init(":50051")
	pod := &cdpb.PodResource{
		Name:      "test1",
		Namespace: "default",
	}
	err := DirectPutPodResource(context.Background(), pod, consts.OpDel)
	if err != nil {
		t.Errorf("put pod resource error, %v", err)
	}
}
func TestGetAlivePod(t *testing.T) {
	Init(":50051")
	data, err := GetAlivePod(context.Background(), "bounded")
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("%+v", data)
}

func TestPutStorageUtil(t *testing.T) {
	Init(":50051")
	err := PutStorageUtil(context.Background(), "podb", "456")
	if err != nil {
		t.Errorf("%+v", err)
	}
}
