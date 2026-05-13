// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"io"
	"sync"

	"mall-activity-rpc/activityclient"
	"mall-api/internal/config"
	"mall-api/internal/middleware"
	"mall-cart-rpc/cartclient"
	"mall-common/configcenter"
	"mall-logistics-rpc/logisticsclient"
	"mall-order-rpc/orderclient"
	"mall-payment-rpc/paymentclient"
	"mall-product-rpc/productclient"
	"mall-review-rpc/reviewclient"
	"mall-reward-rpc/rewardclient"
	"mall-risk-rpc/riskclient"
	"mall-rule-rpc/ruleclient"
	"mall-shop-rpc/shopservice"
	"mall-user-rpc/userclient"
	"mall-workflow-rpc/workflowclient"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/yaml.v3"
)

type ObjectStore interface {
	PutObject(ctx context.Context, bucket, object string, r io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	BucketExists(ctx context.Context, bucket string) (bool, error)
	MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error
}

// hotMinioClient wraps an ObjectStore and allows atomic swapping on config change.
// It implements ObjectStore so existing callers need no changes.
type hotMinioClient struct {
	mu     sync.RWMutex
	client ObjectStore
}

func newHotMinioClient(c ObjectStore) *hotMinioClient { return &hotMinioClient{client: c} }

func (h *hotMinioClient) swap(c ObjectStore) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.client = c
}

func (h *hotMinioClient) cur() ObjectStore {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client
}

func (h *hotMinioClient) PutObject(ctx context.Context, bucket, object string, r io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return h.cur().PutObject(ctx, bucket, object, r, size, opts)
}
func (h *hotMinioClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return h.cur().BucketExists(ctx, bucket)
}
func (h *hotMinioClient) MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error {
	return h.cur().MakeBucket(ctx, bucket, opts)
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
	ShopRpc      shopservice.ShopService

	Minio       ObjectStore
	AdminToken  rest.Middleware
	SessionAuth rest.Middleware

	// hot-reloadable fields (updated by etcd watcher without restart)
	adminTokenHot *configcenter.HotConfig[string]
	minioHot      *hotMinioClient
}

func NewServiceContext(c config.Config, etcdHosts []string) *ServiceContext {
	mc := mustNewMinioClient(c)
	minioHot := newHotMinioClient(mc)
	adminTokenHot := configcenter.NewHotConfig(c.AdminToken)

	svc := &ServiceContext{
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
		ShopRpc:      shopservice.NewShopService(zrpc.MustNewClient(c.ShopRpc)),
		Minio:       minioHot,
		AdminToken:  middleware.NewAdminTokenMiddleware(adminTokenHot).Handle,
		adminTokenHot: adminTokenHot,
		minioHot:      minioHot,
	}
	// P0 login revamp: opaque-token Redis-session check. Needs UserRpc, which
	// is initialised above, so we wire it after the struct literal.
	svc.SessionAuth = middleware.NewSessionAuthMiddleware(svc.UserRpc)

	if len(etcdHosts) > 0 {
		go configcenter.NewWatcher(etcdHosts).Watch(configcenter.ServiceKey("yw-mall", "api-gateway"), svc.onConfigChange)
	}
	return svc
}

func (s *ServiceContext) onConfigChange(data []byte) {
	var newCfg config.Config
	if err := yaml.Unmarshal(data, &newCfg); err != nil {
		logx.Errorf("[configcenter] mall-api config parse error: %v", err)
		return
	}
	s.adminTokenHot.Set(newCfg.AdminToken)
	logx.Infof("[configcenter] mall-api: AdminToken updated")

	if mc, err := newMinioClient(newCfg); err == nil {
		s.minioHot.swap(mc)
		logx.Infof("[configcenter] mall-api: MinIO client reloaded (endpoint=%s)", newCfg.MinIO.Endpoint)
	} else {
		logx.Errorf("[configcenter] mall-api: MinIO reload failed: %v", err)
	}
}

func mustNewMinioClient(c config.Config) *minio.Client {
	mc, err := newMinioClient(c)
	if err != nil {
		panic(err)
	}
	return mc
}

func newMinioClient(c config.Config) (*minio.Client, error) {
	return minio.New(c.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIO.AccessKey, c.MinIO.SecretKey, ""),
		Secure: c.MinIO.UseSSL,
	})
}
