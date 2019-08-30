package config

import "testing"

func TestGetCoordinator(t *testing.T) {
	path := "../../config.toml"
	Init(path)
	t.Logf("%+v", GetCoordinator())
}

func TestGetNode(t *testing.T) {
	path := "../../config.toml"
	Init(path)
	t.Logf("%+v", GetNode())
}
