package container

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
)

type Client struct {
	cli *client.Client
}

type ContainerInterafce interface {
	GetCgroupPath(ctx context.Context, dockerId string) (string, error)
}

func NewClient() ContainerInterafce {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		glog.Errorf("create docker client error, err=%+v", err)
		panic(err)
	}
	return &Client{
		cli: cli,
	}
}

func (c *Client) GetCgroupPath(ctx context.Context, dockerId string) (string, error) {
	data, err := c.cli.ContainerInspect(ctx, dockerId)
	if err != nil {
		glog.Errorf("get container error, err=%+v", err)
		return "", err
	}
	path := data.HostConfig.CgroupParent
	return path, nil
}
