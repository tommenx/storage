package watcher

import (
	"github.com/golang/glog"
	"time"
)

type watcher struct {
	period time.Duration
}

type Watcher interface {
	Run(stopCh <-chan struct{})
}

func NewWatcher(period time.Duration) Watcher {
	return &watcher{
		period: period,
	}
}

func (w *watcher) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(w.period)
	for {
		select {
		case <-ticker.C:
			_ = ReportRemainingResource()
		case <-stopCh:
			glog.Info("stop report node storage info")
			return
		}
	}

}
