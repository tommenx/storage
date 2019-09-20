package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/tommenx/storage/pkg/controller"
)

const (
	PluginFolder = "/var/lib/kubelet/pligins/lvmplugin.csi.alibabacloud.com"
	DriverName   = "lvmplugin.csi.alibabacloud.com"
	CSIVersion   = "v1.0.0"
)

type lvmDriver struct {
	driver           *csicommon.CSIDriver
	endpoint         string
	idServer         csi.IdentityServer
	nodeServer       csi.NodeServer
	controllerServer csi.ControllerServer
	cap              []*csi.VolumeCapability_AccessMode
	cscap            []*csi.ControllerServiceCapability
}

type Driver interface {
	Run()
}

func NewLvmDriver(nodeID, endpoint string, path string) Driver {
	d := &lvmDriver{}
	d.endpoint = endpoint
	csiDriver := csicommon.NewCSIDriver(DriverName, CSIVersion, nodeID)
	d.driver = csiDriver
	d.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
	})
	d.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	d.idServer = csicommon.NewDefaultIdentityServer(d.driver)
	d.controllerServer = NewControllerServer(d.driver)
	_, informerFactory := controller.NewCliAndInformer(path)
	pvController := controller.NewPVController(informerFactory)
	stop := make(chan struct{})
	go informerFactory.Start(stop)
	pvController.Run(stop)
	nodeServer, err := NewNodeServer(d.driver, false, pvController)
	if err != nil {
		glog.Errorf("lvm can't start node server,err %v \n", err)
	}
	d.nodeServer = nodeServer
	return d
}

func (d *lvmDriver) Run() {
	glog.Infof("start to run csi-plugin, name:%s, version:%s", DriverName, CSIVersion)
	server := csicommon.NewNonBlockingGRPCServer()
	server.Start(d.endpoint, d.idServer, d.controllerServer, d.nodeServer)
	server.Wait()
}
