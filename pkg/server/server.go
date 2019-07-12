package server

import (
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/cdpb"
	"github.com/tommenx/storage/pkg/etcd"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	db etcd.EtcdInterface
}

type Server interface {
	Run()
}

func NewServer(db etcd.EtcdInterface) Server {
	return &server{
		db: db,
	}
}

func (s *server) Run() {
	lst, err := net.Listen("tcp", ":50051")
	if err != nil {
		glog.Errorf("listen error, err=%+v", err)
	}
	grpcServer := grpc.NewServer()
	cdpb.RegisterCoordinatorServer(grpcServer, &server{})
	grpcServer.Serve(lst)
}
