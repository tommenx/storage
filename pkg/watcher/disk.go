package watcher

import (
	"bufio"
	"context"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/config"
	"github.com/tommenx/storage/pkg/consts"
	"github.com/tommenx/storage/pkg/rpc"
	"github.com/tommenx/storage/pkg/utils"
	"io"
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
	remaining := config.GetCapability()
	for k, v := range remaining {
		remaining[k] = v - used[k]
	}
	return remaining, nil
}

func CheckPodStorageUtil() {
	ctx := context.Background()
	instance, err := rpc.GetAlivePod(ctx, "bounded")
	if err != nil {
		glog.Errorf("get alive pod error=%+v", err)
		return
	}
	cmd := "iostat"
	args := []string{"-x", "-m", "-N", "2", "2"}
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		glog.Errorf("get device status error, err=%+v", err)
		return
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[len(lines)/2+1:]
	utilInfo := formatIostatResult(lines)
	report := make(map[string]string)
	for pod, volume := range instance {
		target := "vgdata-" + strings.ReplaceAll(volume, "-", "--")
		if util, ok := utilInfo[target]; ok {
			report[pod] = util[0] + "-" + util[1]
		}
	}
	if err := rpc.PutStorageUtil(ctx, report, "aaa"); err != nil {
		glog.Errorf("PutStorageUtil error, err=%+v", err)
		return
	}
}

//读-写用 - 分割
func formatIostatResult(strs []string) map[string][]string {
	data := make(map[string][]string)
	for _, line := range strs {
		fields := strings.Fields(line)
		if len(fields) > 7 {
			data[fields[0]] = append(data[fields[0]], fields[5])
			data[fields[0]] = append(data[fields[0]], fields[6])
		}
	}
	return data
}
func GetIostatInfo(msgCh chan map[string][]string) error {
	command := "iostat"
	args := []string{"-x", "-m", "-N", "10"}
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	f1 := false
	lines := make([]string, 0)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}

		if strings.HasPrefix(line, "Device") {
			f1 = true
		}
		if f1 {
			lines = append(lines, line)
		}
		if f1 && line == "\n" {
			f1 = false
			lines = lines[:len(lines)-1]
			data := formatIostatResult(lines)
			msgCh <- data
			lines = make([]string, 0)
		}
	}
	return nil
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

func InitReport() error {
	nodeName := config.GetNode().Name
	level := config.GetNode().Storage.Level
	volumeGroup := config.GetNode().Storage.Name
	capability := config.GetCapability()
	ctx := context.Background()
	info := &cdpb.Storage{
		Name:     volumeGroup,
		Level:    level,
		Resource: capability,
	}
	infos := []*cdpb.Storage{}
	infos = append(infos, info)
	err := rpc.PutNodeStorage(ctx, nodeName, consts.KindRemaining, infos)
	err = rpc.PutNodeStorage(ctx, nodeName, consts.KindCapability, infos)
	err = rpc.PutNodeStorage(ctx, nodeName, consts.KindAllocation, infos)
	if err != nil {
		glog.Errorf("rpc put node storage error, err=%+v", err)
		return err
	}
	return nil
}
