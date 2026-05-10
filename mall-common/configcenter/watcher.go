package configcenter

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/zeromicro/go-zero/core/logx"
)

// Watcher watches an etcd key and calls onChange whenever the value is updated.
type Watcher struct {
	hosts []string
}

// NewWatcher creates a Watcher using the provided etcd hosts.
func NewWatcher(hosts []string) *Watcher {
	return &Watcher{hosts: hosts}
}

// Watch runs forever in a goroutine; reconnects automatically on failure.
// Call it with: go watcher.Watch(key, fn)
func (w *Watcher) Watch(key string, onChange func(newValue []byte)) {
	for {
		if err := w.runWatch(key, onChange); err != nil {
			logx.Errorf("[configcenter] watcher error for %s: %v — retrying in 5s", key, err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (w *Watcher) runWatch(key string, onChange func([]byte)) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   w.hosts,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	logx.Infof("[configcenter] watching %s", key)
	watchCh := cli.Watch(context.Background(), key)

	for resp := range watchCh {
		if resp.Err() != nil {
			return resp.Err()
		}
		for _, event := range resp.Events {
			if event.Type == clientv3.EventTypePut {
				logx.Infof("[configcenter] config changed: %s (rev=%d)", key, event.Kv.ModRevision)
				onChange(event.Kv.Value)
			}
		}
	}
	return nil
}
