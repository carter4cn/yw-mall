// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"mall-api/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
	"mall-cart-rpc/cartclient"
	"mall-order-rpc/orderclient"
	"mall-payment-rpc/paymentclient"
	"mall-product-rpc/productclient"
	"mall-user-rpc/userclient"
)

type ServiceContext struct {
	Config     config.Config
	UserRpc    userclient.User
	ProductRpc productclient.Product
	OrderRpc   orderclient.Order
	CartRpc    cartclient.Cart
	PaymentRpc paymentclient.Payment
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		UserRpc:    userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		ProductRpc: productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		OrderRpc:   orderclient.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		CartRpc:    cartclient.NewCart(zrpc.MustNewClient(c.CartRpc)),
		PaymentRpc: paymentclient.NewPayment(zrpc.MustNewClient(c.PaymentRpc)),
	}
}
