package driver

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type LvmVolume struct {
	PVName      string //pv的名字
	LVName      string //逻辑卷的名字
	VolumeId    string //存储卷分配的id
	DevicePath  string //设备的地址
	VolumeGroup string //所属的卷组
	Maj         string // 主设备号
	Min         string //副设备号
	Size        int64  //大小，以GB为单位
}

type Operation interface {
	Create(prefix string) error
	Delete(prefix string) error
}

func (lvm *LvmVolume) Create(prefix string) error {
	createCmd := "lvcreate"
	if len(prefix) != 0 {
		createCmd = fmt.Sprintf("%s lvcreate", prefix)
	}
	volsz := fmt.Sprintf("%dG", lvm.Size)
	args := []string{"-L", volsz, lvm.VolumeGroup}
	cmd := GetCmd(createCmd, args)
	fmt.Printf("create lv, command %s\n", cmd)
	out, err := Run(cmd)
	if err != nil {
		glog.Errorf("create lv error, err=%+v, output=%s", err, string(out))
		return err
	}
	lvm.LVName = extractLVName(string(out))
	lvm.DevicePath = fmt.Sprintf("/dev/%s/%s", lvm.VolumeGroup, lvm.LVName)
	glog.Infof("success create lvm [%s] in vg [%s] with the path %s", lvm.LVName, lvm.VolumeGroup, lvm.DevicePath)
	return nil
}

func (lvm *LvmVolume) Delete(prefix string) error {
	deleteCmd := "lvremove"
	if len(prefix) != 0 {
		deleteCmd = fmt.Sprintf("%s lvremove", prefix)
	}
	args := []string{"-y", lvm.DevicePath}
	cmd := GetCmd(deleteCmd, args)
	fmt.Printf("remove lv, command %s", cmd)
	out, err := Run(cmd)
	if err != nil {
		glog.Errorf("%v failed to remove lvm, output: %s", err, string(out))
		return err
	}
	glog.Infof("success remove lvm [%s] in vg [%s] with the path %s", lvm.LVName, lvm.VolumeGroup, lvm.DevicePath)
	return nil
}

func extractLVName(str string) string {
	strs := strings.Split(str, `"`)
	return strs[1]
}
