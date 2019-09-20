package driver

import (
	"fmt"
	"github.com/golang/glog"
	"math"
	"strings"
)

type LvmVolume struct {
	PVName      string //pv的名字
	DevicePath  string //设备的地址
	VolumeGroup string //所属的卷组
	Maj         string // 主设备号
	Min         string //副设备号
	Size        int64  //大小，以B为单位
}

var GB float64 = 1024 * 1024 * 1024

type Operation interface {
	Create(prefix string) error
	Delete(prefix string) error
}

func (lvm *LvmVolume) GetFormatSize() int64 {
	sz := int64(math.Ceil(float64(lvm.Size) / GB))
	return sz
}

func (lvm *LvmVolume) Create(prefix string) error {
	createCmd := "lvcreate"
	if len(prefix) != 0 {
		createCmd = fmt.Sprintf("%s lvcreate", prefix)
	}
	sz := int64(math.Ceil(float64(lvm.Size) / GB))
	volsz := fmt.Sprintf("%dG", sz)
	args := []string{"-L", volsz, "-n", lvm.PVName, lvm.VolumeGroup}
	cmd := GetCmd(createCmd, args)
	fmt.Printf("create lv, command %s\n", cmd)
	out, err := Run(cmd)
	if err != nil {
		glog.Errorf("create lv error, err=%+v, output=%s", err, string(out))
		return err
	}
	lvm.DevicePath = fmt.Sprintf("/dev/%s/%s", lvm.VolumeGroup, lvm.PVName)
	glog.Infof("success create lvm [%s] in vg [%s] with the path %s", lvm.PVName, lvm.VolumeGroup, lvm.DevicePath)
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
	glog.Infof("success remove lvm [%s] in vg [%s] with the path %s", lvm.PVName, lvm.VolumeGroup, lvm.DevicePath)
	return nil
}

func extractLVName(str string) string {
	strs := strings.Split(str, `"`)
	return strs[1]
}
