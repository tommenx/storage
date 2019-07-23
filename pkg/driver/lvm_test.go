package driver

import "testing"

func TestCreate(t *testing.T) {
	lvm := &LvmVolume{}
	lvm.Size = 1
	lvm.VolumeGroup = "vgdata"
	err := lvm.Create("nsenter --mount=/proc/1/ns/mnt")
	if err != nil {
		t.Errorf("err is %+v", err)
	}
	t.Logf("lvm is %+v", lvm)
}

func TestDelete(t *testing.T) {
	lvm := &LvmVolume{}
	lvm.DevicePath = "/dev/vgdata/lvol8"
	err := lvm.Delete("nsenter --mount=/proc/1/ns/mnt")
	if err != nil {
		t.Errorf("err is %+v", err)
	}
}
