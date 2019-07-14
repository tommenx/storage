package container

import (
	"context"
	"testing"
)

func TestGetCgroupPath(t *testing.T) {
	client := NewClient()
	dockerId := "dd13c9a7e5ce2238a982265c6f4f21cb9c4f76c3bd10fa50a98f51ac902a15ae"
	path, err := client.GetCgroupPath(context.Background(), dockerId)
	if err != nil {
		t.Errorf("error is %+v", err)
	}
	t.Logf("path is %s", path)
}
