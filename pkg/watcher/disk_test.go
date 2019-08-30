package watcher

import (
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/rpc"
	"testing"
)

func TestGetRemainingResource(t *testing.T) {
	_, err := GetRemainingResource("sda")
	if err != nil {
		t.Errorf("error is %+v", err)
	}
}

func TestReportRemainingResource(t *testing.T) {
	rpc.Init(":50051")
	config.Init("../../config.toml")
	err := ReportRemainingResource()
	if err != nil {
		t.Errorf("remaining resource error, err=%+v", err)
		return
	}
	t.Logf("success")
}
