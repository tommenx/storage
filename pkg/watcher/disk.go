package watcher

import (
	"context"
	"fmt"
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
	args := []string{"-x", "-m", "-p", device, "1", "2"}
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("get device status error, err=%+v", err)
		return nil, err
	}
	used := make(map[string]int64)
	lines := strings.Split(string(out), "\n")
	for index := len(lines) / 2; index < len(lines); index++ {
		line := lines[index]
		if strings.HasPrefix(line, device) {
			//fmt.Printf("line %+v\n", line)
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				used["write_bps_device"] = utils.GetInt64(fields[6])
				used["read_bps_device"] = utils.GetInt64(fields[5])
			}
			break
		}
	}
	fmt.Printf("used %+v \n", used)
	remaining := config.GetCapability()
	for k, v := range remaining {
		remaining[k] = v - used[k]
	}
	//fmt.Printf("reamianing %+v \n", remaining)
	return remaining, nil
}

func ReportRemainingResource() error {
	deviceName := config.GetNode().Storage.Device
	nodeName := config.GetNode().Name
	level := config.GetNode().Storage.Level
	volumeGroup := config.GetNode().Storage.Name
	remaining, err := GetRemainingResource(deviceName)
	if err != nil {
		glog.Errorf("get remaining resource error, err=%+v", err)
		return err
	}
	ctx := context.Background()
	info := &cdpb.Storage{
		Name:     volumeGroup,
		Level:    level,
		Resource: remaining,
	}
	infos := []*cdpb.Storage{}
	infos = append(infos, info)
	err = rpc.PutNodeStorage(ctx, nodeName, consts.KindRemaining, infos)
	if err != nil {
		glog.Errorf("rpc put node storage error, err=%+v", err)
		return err
	}
	return nil
}
