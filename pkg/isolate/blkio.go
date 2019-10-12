package isolate

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"path/filepath"
)

var (
	prefixBlkioPath = "/sys/fs/cgroup/blkio"
	MB              = int64(1024 * 1024)
)

//bps单位是MB,iops的单位是次
//write_bps_device 10M
//read_bps_device 20M
func SetBlkio(cgroupParent string, dockerId string, requests map[string]int64, maj, min string) error {
	device := fmt.Sprintf("%s:%s", maj, min)
	requests = parseUnit(requests)
	path := filepath.Join(prefixBlkioPath, cgroupParent, dockerId)
	var err error
	for name, val := range requests {
		if name == "space" {
			continue
		}
		name := fmt.Sprintf("blkio.throttle.%s", name)
		devicePath := filepath.Join(path, name)
		limit := fmt.Sprintf("%s %d", device, val)
		fmt.Printf("path=%s,val = %s\n", devicePath, limit)
		err = ioutil.WriteFile(devicePath, []byte(limit), 0755)
		if err != nil {
			glog.Errorf("set blkio cgroup error, err=%+v", err)
			return err
		}
	}
	return err
}

func parseUnit(before map[string]int64) map[string]int64 {
	for name, size := range before {
		if name == "read_bps_device" || name == "write_bps_device" {
			size = size * MB
			before[name] = size
		}
	}
	return before

}
