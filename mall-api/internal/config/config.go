// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc     zrpc.RpcClientConf
	ProductRpc  zrpc.RpcClientConf
	OrderRpc    zrpc.RpcClientConf
	CartRpc     zrpc.RpcClientConf
	PaymentRpc  zrpc.RpcClientConf
	ActivityRpc zrpc.RpcClientConf
	RuleRpc     zrpc.RpcClientConf
	WorkflowRpc zrpc.RpcClientConf
	RewardRpc   zrpc.RpcClientConf
	RiskRpc     zrpc.RpcClientConf
	ReviewRpc      zrpc.RpcClientConf
	LogisticsRpc   zrpc.RpcClientConf
	Kuaidi100      struct {
		WebhookCustomer string
		WebhookKey      string
	}

	ReviewMedia struct {
		MaxImages      int
		MaxImageSizeMB int
		MaxVideoSizeMB int
		Bucket         string
	}
	AdminToken string
	MinIO      struct {
		Endpoint  string
		AccessKey string
		SecretKey string
		UseSSL    bool
	}
}
