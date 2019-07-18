package isolate

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

//func TestSetBlkio(t *testing.T) {
//	request := map[string]int64{
//		"read_bps_device":   10,
//		"write_bps_device":  1,
//		"read_iops_device":  1000,
//		"write_iops_device": 500,
//	}
//	path := "/kubepods/burstable/pod74cedcef06a480b980163fb25e03abe6"
//	dockerId := "c282e951aa8a3d9433b3afb65eaedcedc53e5dafaa018483d3223abe85452bb9"
//	maj := "8"
//	min := "19"
//	SetBlkio(path, dockerId, request, maj, min)
//}

func TestWriteFile(t *testing.T) {
	path := "/sys/fs/cgroup/blkio/kubepods/burstable/pod74cedcef06a480b980163fb25e03abe6/c282e951aa8a3d9433b3afb65eaedcedc53e5dafaa018483d3223abe85452bb9"
	filePath := filepath.Join(path, "blkio.throttle.write_bps_device")
	content, err := ioutil.ReadFile(filePath)
	err = ioutil.WriteFile(filePath, []byte("11:22 123456"), 0777)
	if err != nil {
		t.Errorf("write file error, err=%+v", err)
	} else {
		t.Log(string(content))
	}

}
