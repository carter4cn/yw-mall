// Phase 1 优惠活动 RPC 服务入口
// 端口 9018, etcd key yw-mall/promotion-rpc, 配置文件 etc/promotion.yaml
package main

import (
	"flag"
	"fmt"

	"mall-promotion-rpc/internal/config"
	"mall-promotion-rpc/internal/server"
	"mall-promotion-rpc/internal/svc"
	"mall-promotion-rpc/promotion"

	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/promotion.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, configcenter.ServiceKey("yw-mall", "promotion-rpc"), *configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		promotion.RegisterPromotionServer(grpcServer, server.NewPromotionServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting promotion rpc server at %s...\n", c.ListenOn)
	s.Start()
}
