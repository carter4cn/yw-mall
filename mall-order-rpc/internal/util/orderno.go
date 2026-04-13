package util

import (
	"fmt"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	lastSeq int64
)

func GenerateOrderNo() string {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	seq := now.UnixNano() / 1e6
	if seq <= lastSeq {
		seq = lastSeq + 1
	}
	lastSeq = seq

	return fmt.Sprintf("%s%06d", now.Format("20060102150405"), seq%1000000)
}
