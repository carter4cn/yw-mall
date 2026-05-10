package main

import (
	"flag"
	"fmt"

	"mall-risk-rpc/internal/config"
	"mall-risk-rpc/internal/server"
	"mall-risk-rpc/internal/svc"
	"mall-risk-rpc/risk"

	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/risk.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, configcenter.ServiceKey("yw-mall", "risk-rpc"), *configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		risk.RegisterRiskServer(grpcServer, server.NewRiskServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
