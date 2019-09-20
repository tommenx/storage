package config

import (
	"github.com/BurntSushi/toml"
	"github.com/golang/glog"
)

//用于catch-up连接coordinator和上报资源
type Config struct {
	Coordinator Coordinator `toml:"coordinator"`
	Node        Node        `toml:"node"`
}

type Coordinator struct {
	Ip   string `toml:"ip"`
	Port string `toml:"port"`
}

type Storage struct {
	Name   string `toml:"name"`
	Device string `toml:"device"`
	Level  string `toml:"level"`
	Space  int64  `toml:"space"`
	Write  int64  `toml:"write"`
	Read   int64  `toml:"read"`
}

type Node struct {
	Name    string  `toml:"name"`
	Storage Storage `toml:"storage"`
}

var c Config

func Init(path string) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		glog.Errorf("read config file error, err=%+v", err)
		panic(err)
	}
}

func GetCoordinator() Coordinator {
	return c.Coordinator
}

func GetNode() Node {
	return c.Node
}

func GetCapability() map[string]int64 {
	data := make(map[string]int64)
	data["write_bps_device"] = c.Node.Storage.Write
	data["read_bps_device"] = c.Node.Storage.Read
	data["space"] = c.Node.Storage.Space
	return data
}
