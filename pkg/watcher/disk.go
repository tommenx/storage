package watcher

import (
	"context"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/consts"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/utils"
	"os/exec"
	"strings"
)

func GetRemainingResource(device string) (map[string]int64, error) {
	cmd := "iostat"
	args := []string{"-x", "-m", "-p", device}
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("get device status error, err=%+v", err)
		return nil, err
	}
	used := make(map[string]int64)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, device) {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				used["write_bps_device"] = utils.GetInt64(fields[6])
				used["read_bps_device"] = utils.GetInt64(fields[5])
			}
			break
		}
	}
	remaining := config.GetCapability()
	for k, v := range remaining {
		remaining[k] = v - used[k]
	}
	return remaining, nil
}

func ReportRemainingResource(node, device, name, level string) error {
	remaining, err := GetRemainingResource(device)
	if err != nil {
		glog.Errorf("get remaining resource error, err=%+v", err)
		return err
	}
	ctx := context.Background()
	info := &cdpb.Storage{
		Name:     name,
		Level:    level,
		Resource: remaining,
	}
	infos := []*cdpb.Storage{}
	infos = append(infos, info)
	err = rpc.PutNodeStorage(ctx, node, consts.KindCapability, infos)
	if err != nil {
		glog.Errorf("rpc put node storage error, err=%+v", err)
		return err
	}
	return nil
}
