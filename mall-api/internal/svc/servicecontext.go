// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"io"

	"mall-activity-rpc/activityclient"
	"mall-api/internal/config"
	"mall-api/internal/middleware"
	"mall-cart-rpc/cartclient"
	"mall-order-rpc/orderclient"
	"mall-payment-rpc/paymentclient"
	"mall-product-rpc/productclient"
	"mall-logistics-rpc/logisticsclient"
	"mall-review-rpc/reviewclient"
	"mall-reward-rpc/rewardclient"
	"mall-risk-rpc/riskclient"
	"mall-rule-rpc/ruleclient"
	"mall-user-rpc/userclient"
	"mall-workflow-rpc/workflowclient"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ObjectStore interface {
	PutObject(ctx context.Context, bucket, object string, r io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	BucketExists(ctx context.Context, bucket string) (bool, error)
	MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error
}

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
	ReviewRpc    reviewclient.Review
	LogisticsRpc logisticsclient.Logistics

	Minio      ObjectStore
	AdminToken rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	mc, err := minio.New(c.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIO.AccessKey, c.MinIO.SecretKey, ""),
		Secure: c.MinIO.UseSSL,
	})
	if err != nil {
		panic(err)
	}
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
		ReviewRpc:    reviewclient.NewReview(zrpc.MustNewClient(c.ReviewRpc)),
		LogisticsRpc: logisticsclient.NewLogistics(zrpc.MustNewClient(c.LogisticsRpc)),
		Minio:       mc,
		AdminToken:  middleware.NewAdminTokenMiddleware(c.AdminToken).Handle,
	}
}
