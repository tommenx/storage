package server

import (
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/controller"
	"github.com/tommenx/storage/pkg/store"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	db         store.EtcdInterface
	pv         controller.PVController
	pvc        controller.PVCControlInterface
	data       map[string]string
	lastServer string
}

type Server interface {
	Run()
}

func NewServer(db store.EtcdInterface, pv controller.PVController, pvc controller.PVCControlInterface) Server {
	return &server{
		db:         db,
		pv:         pv,
		pvc:        pvc,
		data:       make(map[string]string),
		lastServer: "",
	}
}

func (s *server) Run() {
	lst, err := net.Listen("tcp", ":50051")
	if err != nil {
		glog.Errorf("listen error, err=%+v", err)
	}
	grpcServer := grpc.NewServer()
	cdpb.RegisterCoordinatorServer(grpcServer, s)
	grpcServer.Serve(lst)
}
