package watcher

import (
	"github.com/golang/glog"
	"time"
)

type watcher struct {
	period time.Duration
	device string
}

type Watcher interface {
	Run(stopCh <-chan struct{})
}

func NewWatcher(period time.Duration, device string) Watcher {
	return &watcher{
		period: period,
		device: device,
	}
}

func (w *watcher) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(w.period)
	for {
		select {
		case <-ticker.C:
			GetRemainingResource(w.device)
		case <-stopCh:
			glog.Info("stop report node storage info")
			return
		}
	}

}
