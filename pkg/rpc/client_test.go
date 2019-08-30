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
			},
		},
		{
			Name:  "hdd2",
			Level: "HDD",
			Resource: map[string]int64{
				"read_bps_device":   200,
				"read_iops_device":  300,
				"write_bps_device":  400,
				"write_iops_device": 500,
			},
		},
	}
	err := PutNodeStorage(context.Background(), "nodebbb", consts.KindCapability, storage)
	if err != nil {
		t.Error(err)
	}
}

func TestGetNodeStorage(t *testing.T) {
	Init(":50051")
	node := "all"
	kind := consts.KindRemaining
	infos, err := GetNodeStorage(context.Background(), node, kind)
	if err != nil {
		t.Errorf("error is %+v", err)
		return
	}
	for k, v := range infos {
		t.Logf("key is %v", k)
		for _, storage := range v.Storage {
			t.Logf("%+v", storage)
		}
	}
}
