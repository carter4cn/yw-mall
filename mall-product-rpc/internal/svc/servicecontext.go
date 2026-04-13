package svc

import (
	"mall-product-rpc/internal/config"
	"mall-product-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	ProductModel  model.ProductModel
	CategoryModel model.CategoryModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:        c,
		ProductModel:  model.NewProductModel(conn, c.Cache),
		CategoryModel: model.NewCategoryModel(conn, c.Cache),
	}
}
