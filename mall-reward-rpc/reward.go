package main

import (
	"context"
	"flag"
	"fmt"

	"mall-reward-rpc/internal/config"
	"mall-reward-rpc/internal/outbox"
	"mall-reward-rpc/internal/server"
	"mall-reward-rpc/internal/svc"
	"mall-reward-rpc/reward"

	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/reward.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, "/mall/config/reward-rpc", *configFile, &c)
	ctx := svc.NewServiceContext(c)

	// Outbox relay drains PENDING rows in the background. The default Publisher
	// is the log-only sink; flip to a real Kafka client by setting the relay's
	// publisher when one is wired in (see internal/outbox/relay.go).
	relay := outbox.NewRelay(ctx.DB, nil)
	relayCtx, relayCancel := context.WithCancel(context.Background())
	relay.Start(relayCtx)
	defer func() {
		relayCancel()
		relay.Stop()
	}()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		reward.RegisterRewardServer(grpcServer, server.NewRewardServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
