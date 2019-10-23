package watcher

import (
	"github.com/golang/glog"
	"time"
)

type watcher struct {
	period time.Duration
}

type Watcher interface {
	InitResource() error
	Run(stopCh <-chan struct{})
}

func NewWatcher(period time.Duration) Watcher {
	return &watcher{
		period: period,
	}
}

func (w *watcher) InitResource() error {
	return InitReport()
}

func (w *watcher) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(w.period)
	for {
		select {
		case <-ticker.C:
			go CheckPodStorageUtil()
			//_ = ReportRemainingResource()
		case <-stopCh:
			glog.Info("stop report node storage info")
			return
		}
	}

}
