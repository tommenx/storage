package config

import "testing"

func TestGetCoordinator(t *testing.T) {
	path := "../../config.toml"
	err := Init(path)
	if err != nil {
		t.Errorf("get config error, err=%+v", err)
		return
	}
	t.Logf("%+v", GetCoordinator())
}

func TestGetNode(t *testing.T) {
	path := "../../config.toml"
	err := Init(path)
	if err != nil {
		t.Errorf("get config error, err=%+v", err)
		return
	}
	t.Logf("%+v", GetNode())
}
