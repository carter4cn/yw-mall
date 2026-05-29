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
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/mall-api.yaml", "the config file")

func main() {
	flag.Parse()

	etcdHosts := configcenter.EtcdHostsFromEnv()
	var c config.Config
	configcenter.MustLoadWithFallback(etcdHosts, configcenter.ServiceKey("yw-mall", "api-gateway"), *configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 全局错误处理 —— 把 gRPC error / errorx.CodeError / errors.New(...) 统一
	// 解包成 {code, message} JSON，FE request.ts 读 body.message 显示真实业务
	// 提示。未注册时 go-zero 默认把 err.Error() 当 plain text body 写回，FE
	// fallback 成 "请求失败"。
	httpx.SetErrorHandlerCtx(handler.BizErrorHandler)

	ctx := svc.NewServiceContext(c, etcdHosts)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
