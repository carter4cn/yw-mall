package main

import (
	"flag"
	"fmt"

	"mall-workflow-rpc/internal/config"
	"mall-workflow-rpc/internal/server"
	"mall-workflow-rpc/internal/svc"
	"mall-workflow-rpc/workflow"

	"mall-common/configcenter"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/workflow.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, "/mall/config/workflow-rpc", *configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		workflow.RegisterWorkflowServer(grpcServer, server.NewWorkflowServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
