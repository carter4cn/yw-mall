// Package configcenter provides etcd-backed config loading with local-file fallback
// and hot-reload support for selected config fields.
//
// Startup flow:
//   1. Read ETCD_HOSTS env var (comma-separated). If set, fetch config YAML from etcd.
//   2. If etcd is unavailable or key is absent, read the local YAML file instead.
//   3. Caller starts a Watcher to receive future changes (hot-reloadable fields only).
package configcenter

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	dialTimeout    = 3 * time.Second
	requestTimeout = 5 * time.Second
)

// EtcdHostsFromEnv returns hosts from the ETCD_HOSTS env var (comma-separated).
// Returns nil if the env var is unset or empty (caller should fall back to local file).
func EtcdHostsFromEnv() []string {
	v := os.Getenv("ETCD_HOSTS")
	if v == "" {
		return nil
	}
	hosts := strings.Split(v, ",")
	for i, h := range hosts {
		hosts[i] = strings.TrimSpace(h)
	}
	return hosts
}

// MustLoadWithFallback loads config from etcd if ETCD_HOSTS is set and the key
// exists; otherwise reads localPath. Fatals on unrecoverable errors.
func MustLoadWithFallback(etcdHosts []string, key, localPath string, dest any) {
	if err := LoadWithFallback(etcdHosts, key, localPath, dest); err != nil {
		logx.Must(fmt.Errorf("[configcenter] failed to load config %s: %w", key, err))
	}
}

// LoadWithFallback loads config from etcd; falls back to localPath on any error.
func LoadWithFallback(etcdHosts []string, key, localPath string, dest any) error {
	if len(etcdHosts) > 0 {
		data, err := fetchFromEtcd(etcdHosts, key)
		if err == nil {
			logx.Infof("[configcenter] loaded %s from etcd", key)
			return yaml.Unmarshal(data, dest)
		}
		logx.Infof("[configcenter] etcd unavailable (%v), falling back to %s", err, localPath)
	}
	return loadFromFile(localPath, dest)
}

func fetchFromEtcd(hosts []string, key string) ([]byte, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := cli.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("key %q not found in etcd", key)
	}
	return resp.Kvs[0].Value, nil
}

func loadFromFile(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return yaml.Unmarshal(data, dest)
}
