package main

import (
	"context"
	"flag"
	"fmt"

	"mall-logistics-rpc/internal/config"
	logisticsServer "mall-logistics-rpc/internal/server"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/internal/worker"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/logistics.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	ws := worker.NewOrderShippedWorker(ctx)
	ws.Start(context.Background())

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		logistics.RegisterLogisticsServer(grpcServer, logisticsServer.NewLogisticsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
