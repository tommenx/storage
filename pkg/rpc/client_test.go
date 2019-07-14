package rpc

import (
	"context"
	"fmt"
	"github.com/tommenx/cdproto/cdpb"
	"testing"
)

func TestPutNodeStorage(t *testing.T) {
	Init(":50051")
	storage := []*cdpb.Storage{
		{
			Name:        "ssd",
			StorageType: cdpb.StorageType_SSD,
			Resource: map[string]int64{
				"read_bps_device":   100,
				"read_iops_device":  200,
				"write_bps_device":  300,
				"write_iops_device": 400,
			},
		},
		{
			Name:        "hdd",
			StorageType: cdpb.StorageType_HDD,
			Resource: map[string]int64{
				"read_bps_device":   200,
				"read_iops_device":  300,
				"write_bps_device":  400,
				"write_iops_device": 500,
			},
		},
	}
	err := PutNodeStorage(context.Background(), "bbb", storage)
	if err != nil {
		t.Error(err)
	}

}

func TestGetNodeStorage(t *testing.T) {
	Init(":50051")
	nodes, err := GetNodeStorage(context.Background())
	if err != nil {
		t.Errorf("get node error,err=%+v", err)
	}
	fmt.Printf("node count = %d", len(nodes))
	for n, v := range nodes {
		fmt.Printf("node name1 = %s\n", n)
		fmt.Printf("node name2 = %s\n", v.NodeName)
		for _, s := range v.Storage {
			fmt.Printf("%v\n", s.Name)
			fmt.Printf("%v\n", s.StorageType)
			fmt.Printf("%v\n", s.Resource)
		}
	}
}
