package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/driver"
	"github.com/tommenx/storage/pkg/rpc"
	"os"
)

var (
	endpoint    string
	nodeId      string
	configPath  string
	coordinator string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&nodeId, "nodeid", "host1", "node id")
	flag.StringVar(&configPath, "config", "../config.toml", "config file path")
	flag.StringVar(&endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	flag.StringVar(&coordinator, "coordinator", "10.48.247.109:50051", "coordinator path")
}
func main() {
	flag.Parse()
	rpc.Init(coordinator)
	drivername := "lvmplugin.csi.alibabacloud.com"
	glog.V(4).Infoln("CSI Driver: ", drivername, nodeId, endpoint)
	path := "/root/.kube/config"
	driver := driver.NewLvmDriver(nodeId, endpoint, path)
	driver.Run()
	os.Exit(0)
}
