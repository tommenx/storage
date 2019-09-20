package driver

import "testing"

func TestGetDeviceNum(t *testing.T) {
	maj, min, ok := GetDeviceNum("vgdata", "pvc--c908ed3b--db73--11e9--be62--309c23e8d374", "")
	t.Logf("maj=%v, min=%v, ok=%v", maj, min, ok)
}

func TestParseExistLogicalVolume(t *testing.T) {
	volumeId := "pvc-1e598eb8-db80-11e9-be62-309c23e8d374"
	volumeGroup := "vgdata"
	vol, err := parseExistLogicalVolume(volumeId, volumeGroup, "")
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%+v", *vol)
	}
}
