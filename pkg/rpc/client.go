package rpc

import (
	"context"
	"errors"
	"github.com/golang/glog"
	"github.com/tommenx/cdproto/base"
	"github.com/tommenx/cdproto/cdpb"
	"google.golang.org/grpc"
)

var cli cdpb.CoordinatorClient

func Init(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		glog.Errorf("create grpc conn error, err=%+v", err)
		panic(err)
	}
	cli = cdpb.NewCoordinatorClient(conn)
}

func PutNodeStorage(ctx context.Context) error {
	req := &cdpb.PutNodeStorageRequest{
		Base: &base.Base{},
		Node: &cdpb.NodeStorage{},
	}
	req.Node.NodeName = "aaaa"
	rsp, err := cli.PutNodeStorage(ctx, req)
	if err != nil {
		glog.Errorf("call PutNodeStorage error, err=%+v", err)
		return err
	}
	if rsp.BaseResp.Code != 0 {
		glog.Errorf("remote server error,code=%v,msg=%v", rsp.BaseResp.Code, rsp.BaseResp.Message)
		return errors.New("remote server error")
	}
	return nil
}
