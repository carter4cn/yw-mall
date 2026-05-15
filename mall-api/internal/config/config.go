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
	ShopRpc        zrpc.RpcClientConf
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

	// S4.2 failed-login lock + S4.5 anything-else-Redis. Optional; when Host
	// is empty mall-api falls back to no-lock behaviour and the c-side login
	// won't trigger lockout.
	Redis struct {
		Host string
		Pass string
		DB   int
	}

	// S4.5 erase endpoint needs direct mall_user DB access for the
	// soft-delete + anonymisation transaction. When empty, the erase
	// endpoint returns "unavailable".
	UserDataSource string
}
