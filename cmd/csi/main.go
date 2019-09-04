package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/driver"
	"github.com/tommenx/storage/pkg/rpc"
	"os"
)

var (
	endpoint   string
	nodeId     string
	configPath string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&nodeId, "nodeid", "host1", "node id")
	flag.StringVar(&configPath, "config", "../config.toml", "config file path")
	flag.StringVar(&endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")
}
func main() {
	flag.Parse()
	rpc.Init("10.48.144.34:50051")
	drivername := "lvmplugin.csi.alibabacloud.com"
	glog.V(4).Infoln("CSI Driver: ", drivername, nodeId, endpoint)
	driver := driver.NewLvmDriver(nodeId, endpoint)
	driver.Run()
	os.Exit(0)
}
