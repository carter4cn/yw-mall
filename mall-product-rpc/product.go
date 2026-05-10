package main

import (
	"flag"
	"fmt"

	"mall-product-rpc/internal/config"
	"mall-product-rpc/internal/server"
	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/product.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, "/mall/config/product-rpc", *configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		product.RegisterProductServer(grpcServer, server.NewProductServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
