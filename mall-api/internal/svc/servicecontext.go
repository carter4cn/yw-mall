// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"mall-activity-rpc/activityclient"
	"mall-api/internal/config"
	"mall-cart-rpc/cartclient"
	"mall-order-rpc/orderclient"
	"mall-payment-rpc/paymentclient"
	"mall-product-rpc/productclient"
	"mall-reward-rpc/rewardclient"
	"mall-risk-rpc/riskclient"
	"mall-rule-rpc/ruleclient"
	"mall-user-rpc/userclient"
	"mall-workflow-rpc/workflowclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	UserRpc     userclient.User
	ProductRpc  productclient.Product
	OrderRpc    orderclient.Order
	CartRpc     cartclient.Cart
	PaymentRpc  paymentclient.Payment
	ActivityRpc activityclient.Activity
	RuleRpc     ruleclient.Rule
	WorkflowRpc workflowclient.Workflow
	RewardRpc   rewardclient.Reward
	RiskRpc     riskclient.Risk
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		UserRpc:     userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc:  productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		OrderRpc:    orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		CartRpc:     cartclient.NewCart(zrpc.MustNewClient(c.CartRpc)),
		PaymentRpc:  paymentclient.NewPayment(zrpc.MustNewClient(c.PaymentRpc)),
		ActivityRpc: activityclient.NewActivity(zrpc.MustNewClient(c.ActivityRpc)),
		RuleRpc:     ruleclient.NewRule(zrpc.MustNewClient(c.RuleRpc)),
		WorkflowRpc: workflowclient.NewWorkflow(zrpc.MustNewClient(c.WorkflowRpc)),
		RewardRpc:   rewardclient.NewReward(zrpc.MustNewClient(c.RewardRpc)),
		RiskRpc:     riskclient.NewRisk(zrpc.MustNewClient(c.RiskRpc)),
	}
}
