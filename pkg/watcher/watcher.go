package watcher

import (
	"context"
	"github.com/golang/glog"
	"github.com/tommenx/storage/pkg/rpc"
	"strings"
	"time"
)

type watcher struct {
	period time.Duration
	dataCh chan map[string][]string
	nodeId string
}

type Watcher interface {
	InitResource() error
	Run(stopCh <-chan struct{})
}

func NewWatcher(period time.Duration, nodeId string) Watcher {
	return &watcher{
		period: period,
		dataCh: make(chan map[string][]string),
		nodeId: nodeId,
	}
}

func (w *watcher) InitResource() error {
	return InitReport()
}

func (w *watcher) Run(stopCh <-chan struct{}) {
	go GetIostatInfo(w.dataCh)
	for {
		select {
		case data := <-w.dataCh:
			w.reportIOutil(data)
		case <-stopCh:
			glog.Info("stop report node storage info")
			return
		}
	}
}

func (w *watcher) reportIOutil(data map[string][]string) {
	ctx := context.Background()
	instance, err := rpc.GetAlivePod(ctx, "bounded")
	if err != nil {
		glog.Errorf("get alive pod error=%+v", err)
		return
	}
	report := make(map[string]string)
	for pod, volume := range instance {
		target := "vgdata-" + strings.ReplaceAll(volume, "-", "--")
		if util, ok := data[target]; ok {
			report[pod] = util[0] + "-" + util[1]
		}
	}
	if err := rpc.PutStorageUtil(ctx, report, w.nodeId); err != nil {
		glog.Errorf("PutStorageUtil error, err=%+v", err)
		return
	}

}
