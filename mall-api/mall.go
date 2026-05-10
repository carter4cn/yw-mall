// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	"mall-api/internal/config"
	"mall-api/internal/handler"
	"mall-api/internal/svc"
	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/mall-api.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, configcenter.ServiceKey("yw-mall", "api-gateway"), *configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c, etcdHosts)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
