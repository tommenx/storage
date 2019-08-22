package watcher

import "testing"

func TestGetRemainingResource(t *testing.T) {
	_, err := GetRemainingResource("sda")
	if err != nil {
		t.Errorf("error is %+v", err)
	}
}
