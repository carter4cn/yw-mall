package svc

import (
	"mall-product-rpc/internal/config"
	"mall-product-rpc/internal/model"
	"mall-shop-rpc/shopservice"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	DB            sqlx.SqlConn
	ProductModel  model.ProductModel
	CategoryModel model.CategoryModel
	ShopRpc       shopservice.ShopService
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:        c,
		DB:            conn,
		ProductModel:  model.NewProductModel(conn, c.Cache),
		CategoryModel: model.NewCategoryModel(conn, c.Cache),
		ShopRpc:       shopservice.NewShopService(zrpc.MustNewClient(c.ShopRpc)),
	}
}
