# Mall 电商微服务 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a multi-repo e-commerce microservices system using go-zero with 1 API gateway + 5 RPC services + 1 shared library.

**Architecture:** Multi-repo approach with mall-common holding shared proto files and utilities, 5 independent RPC services (user, product, order, cart, payment) each with their own MySQL database, and a unified API gateway routing HTTP requests to RPC services via Etcd service discovery.

**Tech Stack:** Go, go-zero, gRPC, protobuf, MySQL, Redis, Etcd, goctl, JWT

**Toolchain:** All binaries are in `/home/carter/workspace/go/bin/`. Every command must be prefixed with `export PATH=$PATH:/home/carter/workspace/go/bin`.

**Base directory:** `/home/carter/workspace/go/go-zero/`

---

## Task 1: Initialize mall-common shared library

**Files:**
- Create: `mall-common/go.mod`
- Create: `mall-common/errorx/errorx.go`
- Create: `mall-common/result/response.go`
- Create: `mall-common/proto/user/user.proto`
- Create: `mall-common/proto/product/product.proto`
- Create: `mall-common/proto/order/order.proto`
- Create: `mall-common/proto/cart/cart.proto`
- Create: `mall-common/proto/payment/payment.proto`

- [ ] **Step 1: Create go module**

```bash
cd /home/carter/workspace/go/go-zero
mkdir -p mall-common && cd mall-common
go mod init mall-common
```

- [ ] **Step 2: Create errorx/errorx.go**

```go
package errorx

const (
	OK                 = 0
	ServerError        = 1001
	ParamError         = 1002
	AuthError          = 1003
	NotFound           = 1004
	UserNotFound       = 2001
	UserAlreadyExist   = 2002
	PasswordError      = 2003
	ProductNotFound    = 3001
	StockNotEnough     = 3002
	OrderNotFound      = 4001
	OrderStatusError   = 4002
	CartEmpty          = 5001
	PaymentNotFound    = 6001
	PaymentStatusError = 6002
)

var message = map[int]string{
	OK:                 "success",
	ServerError:        "server error",
	ParamError:         "invalid parameter",
	AuthError:          "unauthorized",
	NotFound:           "not found",
	UserNotFound:       "user not found",
	UserAlreadyExist:   "user already exists",
	PasswordError:      "wrong password",
	ProductNotFound:    "product not found",
	StockNotEnough:     "stock not enough",
	OrderNotFound:      "order not found",
	OrderStatusError:   "invalid order status",
	CartEmpty:          "cart is empty",
	PaymentNotFound:    "payment not found",
	PaymentStatusError: "invalid payment status",
}

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewCodeError(code int) *CodeError {
	return &CodeError{Code: code, Msg: message[code]}
}

func NewCodeErrorMsg(code int, msg string) *CodeError {
	return &CodeError{Code: code, Msg: msg}
}

func (e *CodeError) Error() string {
	return e.Msg
}
```

- [ ] **Step 3: Create result/response.go**

```go
package result

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"mall-common/errorx"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJsonCtx(nil, w, &Response{
		Code: errorx.OK,
		Msg:  "success",
		Data: data,
	})
}

func Fail(w http.ResponseWriter, err error) {
	if codeErr, ok := err.(*errorx.CodeError); ok {
		httpx.OkJsonCtx(nil, w, &Response{
			Code: codeErr.Code,
			Msg:  codeErr.Msg,
		})
	} else {
		httpx.OkJsonCtx(nil, w, &Response{
			Code: errorx.ServerError,
			Msg:  err.Error(),
		})
	}
}
```

- [ ] **Step 4: Create proto/user/user.proto**

```protobuf
syntax = "proto3";

package user;

option go_package = "./user";

message RegisterReq {
  string username = 1;
  string password = 2;
  string phone = 3;
}

message RegisterResp {
  int64 id = 1;
}

message LoginReq {
  string username = 1;
  string password = 2;
}

message LoginResp {
  int64 id = 1;
  string token = 2;
}

message GetUserReq {
  int64 id = 1;
}

message GetUserResp {
  int64 id = 1;
  string username = 2;
  string phone = 3;
  string avatar = 4;
  int64 create_time = 5;
}

message UpdateUserReq {
  int64 id = 1;
  string phone = 2;
  string avatar = 3;
}

message UpdateUserResp {}

service User {
  rpc Register(RegisterReq) returns (RegisterResp);
  rpc Login(LoginReq) returns (LoginResp);
  rpc GetUser(GetUserReq) returns (GetUserResp);
  rpc UpdateUser(UpdateUserReq) returns (UpdateUserResp);
}
```

- [ ] **Step 5: Create proto/product/product.proto**

```protobuf
syntax = "proto3";

package product;

option go_package = "./product";

message CreateProductReq {
  string name = 1;
  string description = 2;
  int64 price = 3;  // in cents
  int64 stock = 4;
  int64 category_id = 5;
  string images = 6;
}

message CreateProductResp {
  int64 id = 1;
}

message GetProductReq {
  int64 id = 1;
}

message GetProductResp {
  int64 id = 1;
  string name = 2;
  string description = 3;
  int64 price = 4;
  int64 stock = 5;
  int64 category_id = 6;
  string images = 7;
  int32 status = 8;
  int64 create_time = 9;
}

message ListProductsReq {
  int64 category_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message ListProductsResp {
  repeated GetProductResp products = 1;
  int64 total = 2;
}

message UpdateStockReq {
  int64 id = 1;
  int64 delta = 2;  // negative for deduction
}

message UpdateStockResp {}

message SearchProductsReq {
  string keyword = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message SearchProductsResp {
  repeated GetProductResp products = 1;
  int64 total = 2;
}

service Product {
  rpc CreateProduct(CreateProductReq) returns (CreateProductResp);
  rpc GetProduct(GetProductReq) returns (GetProductResp);
  rpc ListProducts(ListProductsReq) returns (ListProductsResp);
  rpc UpdateStock(UpdateStockReq) returns (UpdateStockResp);
  rpc SearchProducts(SearchProductsReq) returns (SearchProductsResp);
}
```

- [ ] **Step 6: Create proto/order/order.proto**

```protobuf
syntax = "proto3";

package order;

option go_package = "./order";

message OrderItem {
  int64 product_id = 1;
  string product_name = 2;
  int64 price = 3;
  int32 quantity = 4;
}

message CreateOrderReq {
  int64 user_id = 1;
  repeated OrderItem items = 2;
}

message CreateOrderResp {
  int64 id = 1;
  string order_no = 2;
  int64 total_amount = 3;
}

message GetOrderReq {
  int64 id = 1;
}

message GetOrderResp {
  int64 id = 1;
  string order_no = 2;
  int64 user_id = 3;
  int64 total_amount = 4;
  int32 status = 5;  // 0=pending, 1=paid, 2=shipped, 3=completed, 4=cancelled
  repeated OrderItem items = 6;
  int64 create_time = 7;
}

message ListOrdersReq {
  int64 user_id = 1;
  int32 status = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message ListOrdersResp {
  repeated GetOrderResp orders = 1;
  int64 total = 2;
}

message UpdateOrderStatusReq {
  int64 id = 1;
  int32 status = 2;
}

message UpdateOrderStatusResp {}

service Order {
  rpc CreateOrder(CreateOrderReq) returns (CreateOrderResp);
  rpc GetOrder(GetOrderReq) returns (GetOrderResp);
  rpc ListOrders(ListOrdersReq) returns (ListOrdersResp);
  rpc UpdateOrderStatus(UpdateOrderStatusReq) returns (UpdateOrderStatusResp);
}
```

- [ ] **Step 7: Create proto/cart/cart.proto**

```protobuf
syntax = "proto3";

package cart;

option go_package = "./cart";

message AddItemReq {
  int64 user_id = 1;
  int64 product_id = 2;
  int32 quantity = 3;
}

message AddItemResp {}

message RemoveItemReq {
  int64 user_id = 1;
  int64 product_id = 2;
}

message RemoveItemResp {}

message ListItemsReq {
  int64 user_id = 1;
}

message CartItem {
  int64 product_id = 1;
  int32 quantity = 2;
  bool selected = 3;
}

message ListItemsResp {
  repeated CartItem items = 1;
}

message ClearCartReq {
  int64 user_id = 1;
}

message ClearCartResp {}

message UpdateQuantityReq {
  int64 user_id = 1;
  int64 product_id = 2;
  int32 quantity = 3;
}

message UpdateQuantityResp {}

service Cart {
  rpc AddItem(AddItemReq) returns (AddItemResp);
  rpc RemoveItem(RemoveItemReq) returns (RemoveItemResp);
  rpc ListItems(ListItemsReq) returns (ListItemsResp);
  rpc ClearCart(ClearCartReq) returns (ClearCartResp);
  rpc UpdateQuantity(UpdateQuantityReq) returns (UpdateQuantityResp);
}
```

- [ ] **Step 8: Create proto/payment/payment.proto**

```protobuf
syntax = "proto3";

package payment;

option go_package = "./payment";

message CreatePaymentReq {
  string order_no = 1;
  int64 user_id = 2;
  int64 amount = 3;
  int32 pay_type = 4;  // 1=alipay, 2=wechat
}

message CreatePaymentResp {
  int64 id = 1;
  string payment_no = 2;
}

message GetPaymentReq {
  int64 id = 1;
}

message GetPaymentResp {
  int64 id = 1;
  string payment_no = 2;
  string order_no = 3;
  int64 user_id = 4;
  int64 amount = 5;
  int32 status = 6;  // 0=pending, 1=success, 2=failed
  int32 pay_type = 7;
  int64 pay_time = 8;
}

message UpdatePaymentStatusReq {
  int64 id = 1;
  int32 status = 2;
}

message UpdatePaymentStatusResp {}

service Payment {
  rpc CreatePayment(CreatePaymentReq) returns (CreatePaymentResp);
  rpc GetPayment(GetPaymentReq) returns (GetPaymentResp);
  rpc UpdatePaymentStatus(UpdatePaymentStatusReq) returns (UpdatePaymentStatusResp);
}
```

- [ ] **Step 9: Run go mod tidy and verify**

```bash
cd /home/carter/workspace/go/go-zero/mall-common
go get github.com/zeromicro/go-zero@latest
go mod tidy
```

- [ ] **Step 10: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git init
git add mall-common/
git commit -m "feat: add mall-common shared library with proto files and error codes"
```

---

## Task 2: Generate and implement user-rpc service

**Files:**
- Create: `mall-user-rpc/` (goctl generated + custom logic)
- Create: `mall-user-rpc/sql/user.sql`

- [ ] **Step 1: Create SQL schema**

Create `mall-user-rpc/sql/user.sql`:

```sql
CREATE DATABASE IF NOT EXISTS mall_user;
USE mall_user;

CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL DEFAULT '',
  `password` varchar(255) NOT NULL DEFAULT '',
  `phone` varchar(20) NOT NULL DEFAULT '',
  `avatar` varchar(255) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Generate RPC code with goctl**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-user-rpc
cd mall-user-rpc
go mod init mall-user-rpc

# Generate RPC code from proto
goctl rpc protoc ../mall-common/proto/user/user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

- [ ] **Step 3: Generate model code with goctl**

```bash
cd /home/carter/workspace/go/go-zero/mall-user-rpc
goctl model mysql ddl --src sql/user.sql --dir internal/model --cache true
```

- [ ] **Step 4: Configure etc/user.yaml**

```yaml
Name: user.rpc
ListenOn: 0.0.0.0:9001
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: user.rpc

DataSource: root:123456@tcp(127.0.0.1:3306)/mall_user?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: Update internal/config/config.go**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
}
```

- [ ] **Step 6: Update internal/svc/servicecontext.go**

```go
package svc

import (
	"mall-user-rpc/internal/config"
	"mall-user-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.Cache),
	}
}
```

- [ ] **Step 7: Implement registerlogic.go**

```go
package logic

import (
	"context"

	"mall-user-rpc/internal/model"
	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, &model.User{
		Username: in.Username,
		Password: string(hash),
		Phone:    in.Phone,
	})
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &user.RegisterResp{Id: id}, nil
}
```

- [ ] **Step 8: Implement loginlogic.go**

```go
package logic

import (
	"context"
	"time"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	u, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password))
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": u.Id,
		"iat": now,
		"exp": now + 86400*7,
	}).SignedString([]byte(l.svcCtx.Config.JwtAuth.AccessSecret))
	if err != nil {
		return nil, err
	}

	return &user.LoginResp{Id: int64(u.Id), Token: token}, nil
}
```

Note: The config needs a JwtAuth section. Update config.go to add:

```go
type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
	JwtAuth    struct {
		AccessSecret string
	}
}
```

And add to etc/user.yaml:

```yaml
JwtAuth:
  AccessSecret: "mall-secret-key-change-in-production"
```

- [ ] **Step 9: Implement getuserlogic.go**

```go
package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *user.GetUserReq) (*user.GetUserResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	return &user.GetUserResp{
		Id:         int64(u.Id),
		Username:   u.Username,
		Phone:      u.Phone,
		Avatar:     u.Avatar,
		CreateTime: u.CreateTime.Unix(),
	}, nil
}
```

- [ ] **Step 10: Implement updateuserlogic.go**

```go
package logic

import (
	"context"

	"mall-user-rpc/internal/svc"
	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *user.UpdateUserReq) (*user.UpdateUserResp, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	u.Phone = in.Phone
	u.Avatar = in.Avatar
	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		return nil, err
	}

	return &user.UpdateUserResp{}, nil
}
```

- [ ] **Step 11: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-user-rpc
go mod tidy
go build ./...
```

- [ ] **Step 12: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-user-rpc/
git commit -m "feat: add user-rpc service with register, login, get, update"
```

---

## Task 3: Generate and implement product-rpc service

**Files:**
- Create: `mall-product-rpc/` (goctl generated + custom logic)
- Create: `mall-product-rpc/sql/product.sql`

- [ ] **Step 1: Create SQL schema**

Create `mall-product-rpc/sql/product.sql`:

```sql
CREATE DATABASE IF NOT EXISTS mall_product;
USE mall_product;

CREATE TABLE IF NOT EXISTS `category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL DEFAULT '',
  `parent_id` bigint unsigned NOT NULL DEFAULT 0,
  `sort` int NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `product` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '',
  `description` text,
  `price` bigint NOT NULL DEFAULT 0 COMMENT 'price in cents',
  `stock` bigint NOT NULL DEFAULT 0,
  `category_id` bigint unsigned NOT NULL DEFAULT 0,
  `images` varchar(1024) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1=on, 0=off',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_category_id` (`category_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Generate RPC code**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-product-rpc && cd mall-product-rpc
go mod init mall-product-rpc
goctl rpc protoc ../mall-common/proto/product/product.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

- [ ] **Step 3: Generate model code**

```bash
cd /home/carter/workspace/go/go-zero/mall-product-rpc
goctl model mysql ddl --src sql/product.sql --dir internal/model --cache true
```

- [ ] **Step 4: Configure etc/product.yaml**

```yaml
Name: product.rpc
ListenOn: 0.0.0.0:9002
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: product.rpc

DataSource: root:123456@tcp(127.0.0.1:3306)/mall_product?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: Update config.go**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
}
```

- [ ] **Step 6: Update servicecontext.go**

```go
package svc

import (
	"mall-product-rpc/internal/config"
	"mall-product-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	ProductModel model.ProductModel
	CategoryModel model.CategoryModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:       c,
		ProductModel: model.NewProductModel(conn, c.Cache),
		CategoryModel: model.NewCategoryModel(conn, c.Cache),
	}
}
```

- [ ] **Step 7: Implement createproductlogic.go**

```go
package logic

import (
	"context"

	"mall-product-rpc/internal/model"
	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProductLogic) CreateProduct(in *product.CreateProductReq) (*product.CreateProductResp, error) {
	result, err := l.svcCtx.ProductModel.Insert(l.ctx, &model.Product{
		Name:        in.Name,
		Description: in.Description,
		Price:       in.Price,
		Stock:       in.Stock,
		CategoryId:  uint64(in.CategoryId),
		Images:      in.Images,
		Status:      1,
	})
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &product.CreateProductResp{Id: id}, nil
}
```

- [ ] **Step 8: Implement getproductlogic.go**

```go
package logic

import (
	"context"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProductLogic) GetProduct(in *product.GetProductReq) (*product.GetProductResp, error) {
	p, err := l.svcCtx.ProductModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	return &product.GetProductResp{
		Id:          int64(p.Id),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		CategoryId:  int64(p.CategoryId),
		Images:      p.Images,
		Status:      int32(p.Status),
		CreateTime:  p.CreateTime.Unix(),
	}, nil
}
```

- [ ] **Step 9: Implement listproductslogic.go**

```go
package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListProductsLogic) ListProducts(in *product.ListProductsReq) (*product.ListProductsResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var products []*product.GetProductResp
	var total int64

	query := "SELECT id, name, description, price, stock, category_id, images, status, create_time FROM product WHERE status = 1"
	countQuery := "SELECT COUNT(*) FROM product WHERE status = 1"

	if in.CategoryId > 0 {
		query += fmt.Sprintf(" AND category_id = %d", in.CategoryId)
		countQuery += fmt.Sprintf(" AND category_id = %d", in.CategoryId)
	}
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT %d, %d", offset, pageSize)

	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	err := conn.QueryRowCtx(l.ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	type ProductRow struct {
		Id          uint64 `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Price       int64  `db:"price"`
		Stock       int64  `db:"stock"`
		CategoryId  uint64 `db:"category_id"`
		Images      string `db:"images"`
		Status      int64  `db:"status"`
		CreateTime  int64  `db:"create_time"`
	}
	var rows []ProductRow
	err = conn.QueryRowsCtx(l.ctx, &rows, query)
	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		products = append(products, &product.GetProductResp{
			Id:          int64(r.Id),
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
			Stock:       r.Stock,
			CategoryId:  int64(r.CategoryId),
			Images:      r.Images,
			Status:      int32(r.Status),
			CreateTime:  r.CreateTime,
		})
	}

	return &product.ListProductsResp{Products: products, Total: total}, nil
}
```

- [ ] **Step 10: Implement updatestocklogic.go**

```go
package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateStockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateStockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateStockLogic {
	return &UpdateStockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateStockLogic) UpdateStock(in *product.UpdateStockReq) (*product.UpdateStockResp, error) {
	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	result, err := conn.ExecCtx(l.ctx,
		"UPDATE product SET stock = stock + ? WHERE id = ? AND stock + ? >= 0",
		in.Delta, in.Id, in.Delta)
	if err != nil {
		return nil, err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("stock not enough or product not found")
	}

	return &product.UpdateStockResp{}, nil
}
```

- [ ] **Step 11: Implement searchproductslogic.go**

```go
package logic

import (
	"context"
	"fmt"

	"mall-product-rpc/internal/svc"
	"mall-product-rpc/product"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SearchProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchProductsLogic {
	return &SearchProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchProductsLogic) SearchProducts(in *product.SearchProductsReq) (*product.SearchProductsResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	keyword := "%" + in.Keyword + "%"
	query := fmt.Sprintf("SELECT id, name, description, price, stock, category_id, images, status, create_time FROM product WHERE status = 1 AND name LIKE '%s' ORDER BY id DESC LIMIT %d, %d", keyword, offset, pageSize)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product WHERE status = 1 AND name LIKE '%s'", keyword)

	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	var total int64
	err := conn.QueryRowCtx(l.ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	type ProductRow struct {
		Id          uint64 `db:"id"`
		Name        string `db:"name"`
		Description string `db:"description"`
		Price       int64  `db:"price"`
		Stock       int64  `db:"stock"`
		CategoryId  uint64 `db:"category_id"`
		Images      string `db:"images"`
		Status      int64  `db:"status"`
		CreateTime  int64  `db:"create_time"`
	}
	var rows []ProductRow
	err = conn.QueryRowsCtx(l.ctx, &rows, query)
	if err != nil {
		return nil, err
	}

	var products []*product.GetProductResp
	for _, r := range rows {
		products = append(products, &product.GetProductResp{
			Id:          int64(r.Id),
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
			Stock:       r.Stock,
			CategoryId:  int64(r.CategoryId),
			Images:      r.Images,
			Status:      int32(r.Status),
			CreateTime:  r.CreateTime,
		})
	}

	return &product.SearchProductsResp{Products: products, Total: total}, nil
}
```

- [ ] **Step 12: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-product-rpc
go mod tidy
go build ./...
```

- [ ] **Step 13: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-product-rpc/
git commit -m "feat: add product-rpc service with CRUD, stock, and search"
```

---

## Task 4: Generate and implement order-rpc service

**Files:**
- Create: `mall-order-rpc/` (goctl generated + custom logic)
- Create: `mall-order-rpc/sql/order.sql`

- [ ] **Step 1: Create SQL schema**

Create `mall-order-rpc/sql/order.sql`:

```sql
CREATE DATABASE IF NOT EXISTS mall_order;
USE mall_order;

CREATE TABLE IF NOT EXISTS `order` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_no` varchar(64) NOT NULL DEFAULT '',
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `total_amount` bigint NOT NULL DEFAULT 0 COMMENT 'in cents',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=paid,2=shipped,3=completed,4=cancelled',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `order_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL DEFAULT 0,
  `product_id` bigint unsigned NOT NULL DEFAULT 0,
  `product_name` varchar(128) NOT NULL DEFAULT '',
  `price` bigint NOT NULL DEFAULT 0,
  `quantity` int NOT NULL DEFAULT 0,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Generate RPC code**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-order-rpc && cd mall-order-rpc
go mod init mall-order-rpc
goctl rpc protoc ../mall-common/proto/order/order.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

- [ ] **Step 3: Generate model code**

```bash
cd /home/carter/workspace/go/go-zero/mall-order-rpc
goctl model mysql ddl --src sql/order.sql --dir internal/model --cache true
```

- [ ] **Step 4: Configure etc/order.yaml**

```yaml
Name: order.rpc
ListenOn: 0.0.0.0:9003
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: order.rpc

DataSource: root:123456@tcp(127.0.0.1:3306)/mall_order?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: Update config.go**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
}
```

- [ ] **Step 6: Add snowflake order number generator**

Create `mall-order-rpc/internal/util/orderno.go`:

```go
package util

import (
	"fmt"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	lastSeq int64
)

func GenerateOrderNo() string {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	seq := now.UnixNano() / 1e6
	if seq <= lastSeq {
		seq = lastSeq + 1
	}
	lastSeq = seq

	return fmt.Sprintf("%s%06d", now.Format("20060102150405"), seq%1000000)
}
```

- [ ] **Step 7: Update servicecontext.go**

```go
package svc

import (
	"mall-order-rpc/internal/config"
	"mall-order-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	OrderModel     model.OrderModel
	OrderItemModel model.OrderItemModel
	SqlConn        sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:         c,
		OrderModel:     model.NewOrderModel(conn, c.Cache),
		OrderItemModel: model.NewOrderItemModel(conn, c.Cache),
		SqlConn:        conn,
	}
}
```

- [ ] **Step 8: Implement createorderlogic.go**

```go
package logic

import (
	"context"

	"mall-order-rpc/internal/model"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/internal/util"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	orderNo := util.GenerateOrderNo()
	var totalAmount int64
	for _, item := range in.Items {
		totalAmount += item.Price * int64(item.Quantity)
	}

	var orderId int64
	err := l.svcCtx.SqlConn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		result, err := session.ExecCtx(ctx,
			"INSERT INTO `order` (order_no, user_id, total_amount, status) VALUES (?, ?, ?, 0)",
			orderNo, in.UserId, totalAmount)
		if err != nil {
			return err
		}
		orderId, _ = result.LastInsertId()

		for _, item := range in.Items {
			_, err = session.ExecCtx(ctx,
				"INSERT INTO order_item (order_id, product_id, product_name, price, quantity) VALUES (?, ?, ?, ?, ?)",
				orderId, item.ProductId, item.ProductName, item.Price, item.Quantity)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &order.CreateOrderResp{
		Id:          orderId,
		OrderNo:     orderNo,
		TotalAmount: totalAmount,
	}, nil
}
```

- [ ] **Step 9: Implement getorderlogic.go**

```go
package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOrderLogic) GetOrder(in *order.GetOrderReq) (*order.GetOrderResp, error) {
	o, err := l.svcCtx.OrderModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	type ItemRow struct {
		ProductId   int64  `db:"product_id"`
		ProductName string `db:"product_name"`
		Price       int64  `db:"price"`
		Quantity    int32  `db:"quantity"`
	}
	var items []ItemRow
	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	err = conn.QueryRowsCtx(l.ctx, &items, "SELECT product_id, product_name, price, quantity FROM order_item WHERE order_id = ?", o.Id)
	if err != nil {
		return nil, err
	}

	var orderItems []*order.OrderItem
	for _, item := range items {
		orderItems = append(orderItems, &order.OrderItem{
			ProductId:   item.ProductId,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	return &order.GetOrderResp{
		Id:          int64(o.Id),
		OrderNo:     o.OrderNo,
		UserId:      int64(o.UserId),
		TotalAmount: o.TotalAmount,
		Status:      int32(o.Status),
		Items:       orderItems,
		CreateTime:  o.CreateTime.Unix(),
	}, nil
}
```

- [ ] **Step 10: Implement listorderslogic.go**

```go
package logic

import (
	"context"
	"fmt"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrdersLogic) ListOrders(in *order.ListOrdersReq) (*order.ListOrdersResp, error) {
	page := in.Page
	if page < 1 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("SELECT id, order_no, user_id, total_amount, status, create_time FROM `order` WHERE user_id = %d", in.UserId)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `order` WHERE user_id = %d", in.UserId)

	if in.Status >= 0 {
		query += fmt.Sprintf(" AND status = %d", in.Status)
		countQuery += fmt.Sprintf(" AND status = %d", in.Status)
	}
	query += fmt.Sprintf(" ORDER BY id DESC LIMIT %d, %d", offset, pageSize)

	conn := sqlx.NewMysql(l.svcCtx.Config.DataSource)
	var total int64
	err := conn.QueryRowCtx(l.ctx, &total, countQuery)
	if err != nil {
		return nil, err
	}

	type OrderRow struct {
		Id          uint64 `db:"id"`
		OrderNo     string `db:"order_no"`
		UserId      uint64 `db:"user_id"`
		TotalAmount int64  `db:"total_amount"`
		Status      int64  `db:"status"`
		CreateTime  int64  `db:"create_time"`
	}
	var rows []OrderRow
	err = conn.QueryRowsCtx(l.ctx, &rows, query)
	if err != nil {
		return nil, err
	}

	var orders []*order.GetOrderResp
	for _, r := range rows {
		orders = append(orders, &order.GetOrderResp{
			Id:          int64(r.Id),
			OrderNo:     r.OrderNo,
			UserId:      int64(r.UserId),
			TotalAmount: r.TotalAmount,
			Status:      int32(r.Status),
			CreateTime:  r.CreateTime,
		})
	}

	return &order.ListOrdersResp{Orders: orders, Total: total}, nil
}
```

- [ ] **Step 11: Implement updateorderstatuslogic.go**

```go
package logic

import (
	"context"

	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrderStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateOrderStatusLogic) UpdateOrderStatus(in *order.UpdateOrderStatusReq) (*order.UpdateOrderStatusResp, error) {
	o, err := l.svcCtx.OrderModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	o.Status = int64(in.Status)
	err = l.svcCtx.OrderModel.Update(l.ctx, o)
	if err != nil {
		return nil, err
	}

	return &order.UpdateOrderStatusResp{}, nil
}
```

- [ ] **Step 12: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-order-rpc
go mod tidy
go build ./...
```

- [ ] **Step 13: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-order-rpc/
git commit -m "feat: add order-rpc service with create, get, list, status update"
```

---

## Task 5: Generate and implement cart-rpc service

**Files:**
- Create: `mall-cart-rpc/` (goctl generated + custom logic)
- Create: `mall-cart-rpc/sql/cart.sql`

- [ ] **Step 1: Create SQL schema**

Create `mall-cart-rpc/sql/cart.sql`:

```sql
CREATE DATABASE IF NOT EXISTS mall_cart;
USE mall_cart;

CREATE TABLE IF NOT EXISTS `cart_item` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `product_id` bigint unsigned NOT NULL DEFAULT 0,
  `quantity` int NOT NULL DEFAULT 0,
  `selected` tinyint(1) NOT NULL DEFAULT 1,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_product` (`user_id`, `product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Generate RPC code**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-cart-rpc && cd mall-cart-rpc
go mod init mall-cart-rpc
goctl rpc protoc ../mall-common/proto/cart/cart.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

- [ ] **Step 3: Generate model code**

```bash
cd /home/carter/workspace/go/go-zero/mall-cart-rpc
goctl model mysql ddl --src sql/cart.sql --dir internal/model --cache true
```

- [ ] **Step 4: Configure etc/cart.yaml**

```yaml
Name: cart.rpc
ListenOn: 0.0.0.0:9004
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: cart.rpc

DataSource: root:123456@tcp(127.0.0.1:3306)/mall_cart?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: Update config.go**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
}
```

- [ ] **Step 6: Update servicecontext.go**

```go
package svc

import (
	"mall-cart-rpc/internal/config"
	"mall-cart-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	CartItemModel model.CartItemModel
	SqlConn       sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:        c,
		CartItemModel: model.NewCartItemModel(conn, c.Cache),
		SqlConn:       conn,
	}
}
```

- [ ] **Step 7: Implement additemlogic.go**

```go
package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddItemLogic {
	return &AddItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddItemLogic) AddItem(in *cart.AddItemReq) (*cart.AddItemResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO cart_item (user_id, product_id, quantity, selected) VALUES (?, ?, ?, 1) ON DUPLICATE KEY UPDATE quantity = quantity + ?",
		in.UserId, in.ProductId, in.Quantity, in.Quantity)
	if err != nil {
		return nil, err
	}

	return &cart.AddItemResp{}, nil
}
```

- [ ] **Step 8: Implement removeitemlogic.go**

```go
package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveItemLogic {
	return &RemoveItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveItemLogic) RemoveItem(in *cart.RemoveItemReq) (*cart.RemoveItemResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"DELETE FROM cart_item WHERE user_id = ? AND product_id = ?",
		in.UserId, in.ProductId)
	if err != nil {
		return nil, err
	}

	return &cart.RemoveItemResp{}, nil
}
```

- [ ] **Step 9: Implement listitemslogic.go**

```go
package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListItemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListItemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListItemsLogic {
	return &ListItemsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListItemsLogic) ListItems(in *cart.ListItemsReq) (*cart.ListItemsResp, error) {
	type ItemRow struct {
		ProductId int64 `db:"product_id"`
		Quantity  int32 `db:"quantity"`
		Selected  bool  `db:"selected"`
	}
	var rows []ItemRow
	err := l.svcCtx.SqlConn.QueryRowsCtx(l.ctx, &rows,
		"SELECT product_id, quantity, selected FROM cart_item WHERE user_id = ?",
		in.UserId)
	if err != nil {
		return nil, err
	}

	var items []*cart.CartItem
	for _, r := range rows {
		items = append(items, &cart.CartItem{
			ProductId: r.ProductId,
			Quantity:  r.Quantity,
			Selected:  r.Selected,
		})
	}

	return &cart.ListItemsResp{Items: items}, nil
}
```

- [ ] **Step 10: Implement clearcartlogic.go**

```go
package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearCartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearCartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearCartLogic {
	return &ClearCartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearCartLogic) ClearCart(in *cart.ClearCartReq) (*cart.ClearCartResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"DELETE FROM cart_item WHERE user_id = ?", in.UserId)
	if err != nil {
		return nil, err
	}

	return &cart.ClearCartResp{}, nil
}
```

- [ ] **Step 11: Implement updatequantitylogic.go**

```go
package logic

import (
	"context"

	"mall-cart-rpc/cart"
	"mall-cart-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateQuantityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateQuantityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateQuantityLogic {
	return &UpdateQuantityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateQuantityLogic) UpdateQuantity(in *cart.UpdateQuantityReq) (*cart.UpdateQuantityResp, error) {
	_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"UPDATE cart_item SET quantity = ? WHERE user_id = ? AND product_id = ?",
		in.Quantity, in.UserId, in.ProductId)
	if err != nil {
		return nil, err
	}

	return &cart.UpdateQuantityResp{}, nil
}
```

- [ ] **Step 12: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-cart-rpc
go mod tidy
go build ./...
```

- [ ] **Step 13: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-cart-rpc/
git commit -m "feat: add cart-rpc service with add, remove, list, clear, update quantity"
```

---

## Task 6: Generate and implement payment-rpc service

**Files:**
- Create: `mall-payment-rpc/` (goctl generated + custom logic)
- Create: `mall-payment-rpc/sql/payment.sql`

- [ ] **Step 1: Create SQL schema**

Create `mall-payment-rpc/sql/payment.sql`:

```sql
CREATE DATABASE IF NOT EXISTS mall_payment;
USE mall_payment;

CREATE TABLE IF NOT EXISTS `payment` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `payment_no` varchar(64) NOT NULL DEFAULT '',
  `order_no` varchar(64) NOT NULL DEFAULT '',
  `user_id` bigint unsigned NOT NULL DEFAULT 0,
  `amount` bigint NOT NULL DEFAULT 0 COMMENT 'in cents',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0=pending,1=success,2=failed',
  `pay_type` tinyint NOT NULL DEFAULT 0 COMMENT '1=alipay,2=wechat',
  `pay_time` timestamp NULL DEFAULT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_payment_no` (`payment_no`),
  KEY `idx_order_no` (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: Generate RPC code**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-payment-rpc && cd mall-payment-rpc
go mod init mall-payment-rpc
goctl rpc protoc ../mall-common/proto/payment/payment.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

- [ ] **Step 3: Generate model code**

```bash
cd /home/carter/workspace/go/go-zero/mall-payment-rpc
goctl model mysql ddl --src sql/payment.sql --dir internal/model --cache true
```

- [ ] **Step 4: Configure etc/payment.yaml**

```yaml
Name: payment.rpc
ListenOn: 0.0.0.0:9005
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: payment.rpc

DataSource: root:123456@tcp(127.0.0.1:3306)/mall_payment?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: Update config.go**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
}
```

- [ ] **Step 6: Update servicecontext.go**

```go
package svc

import (
	"mall-payment-rpc/internal/config"
	"mall-payment-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	PaymentModel model.PaymentModel
	SqlConn      sqlx.SqlConn
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:       c,
		PaymentModel: model.NewPaymentModel(conn, c.Cache),
		SqlConn:      conn,
	}
}
```

- [ ] **Step 7: Implement createpaymentlogic.go**

```go
package logic

import (
	"context"
	"fmt"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePaymentLogic) CreatePayment(in *payment.CreatePaymentReq) (*payment.CreatePaymentResp, error) {
	paymentNo := fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)

	result, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
		"INSERT INTO payment (payment_no, order_no, user_id, amount, status, pay_type) VALUES (?, ?, ?, ?, 0, ?)",
		paymentNo, in.OrderNo, in.UserId, in.Amount, in.PayType)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &payment.CreatePaymentResp{Id: id, PaymentNo: paymentNo}, nil
}
```

- [ ] **Step 8: Implement getpaymentlogic.go**

```go
package logic

import (
	"context"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaymentLogic) GetPayment(in *payment.GetPaymentReq) (*payment.GetPaymentResp, error) {
	p, err := l.svcCtx.PaymentModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	var payTime int64
	if p.PayTime.Valid {
		payTime = p.PayTime.Time.Unix()
	}

	return &payment.GetPaymentResp{
		Id:        int64(p.Id),
		PaymentNo: p.PaymentNo,
		OrderNo:   p.OrderNo,
		UserId:    int64(p.UserId),
		Amount:    p.Amount,
		Status:    int32(p.Status),
		PayType:   int32(p.PayType),
		PayTime:   payTime,
	}, nil
}
```

- [ ] **Step 9: Implement updatepaymentstatuslogic.go**

```go
package logic

import (
	"context"
	"time"

	"mall-payment-rpc/internal/svc"
	"mall-payment-rpc/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePaymentStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePaymentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePaymentStatusLogic {
	return &UpdatePaymentStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePaymentStatusLogic) UpdatePaymentStatus(in *payment.UpdatePaymentStatusReq) (*payment.UpdatePaymentStatusResp, error) {
	if in.Status == 1 {
		// Payment success - set pay_time
		_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE payment SET status = ?, pay_time = ? WHERE id = ?",
			in.Status, time.Now(), in.Id)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := l.svcCtx.SqlConn.ExecCtx(l.ctx,
			"UPDATE payment SET status = ? WHERE id = ?",
			in.Status, in.Id)
		if err != nil {
			return nil, err
		}
	}

	return &payment.UpdatePaymentStatusResp{}, nil
}
```

- [ ] **Step 10: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-payment-rpc
go mod tidy
go build ./...
```

- [ ] **Step 11: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-payment-rpc/
git commit -m "feat: add payment-rpc service with create, get, status update"
```

---

## Task 7: Create mall-api gateway

**Files:**
- Create: `mall-api/mall.api`
- Create: `mall-api/` (goctl generated + custom logic)

- [ ] **Step 1: Create the API definition file**

Create `mall-api/mall.api`:

```api
syntax = "v1"

info (
	title:   "Mall API Gateway"
	desc:    "E-commerce API gateway"
	version: "1.0"
)

// ===== Types =====

type (
	RegisterReq {
		Username string `json:"username"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
	}
	RegisterResp {
		Id int64 `json:"id"`
	}

	LoginReq {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	LoginResp {
		Id    int64  `json:"id"`
		Token string `json:"token"`
	}

	UserInfoResp {
		Id         int64  `json:"id"`
		Username   string `json:"username"`
		Phone      string `json:"phone"`
		Avatar     string `json:"avatar"`
		CreateTime int64  `json:"createTime"`
	}

	UpdateUserReq {
		Phone  string `json:"phone"`
		Avatar string `json:"avatar"`
	}
)

type (
	ProductDetailReq {
		Id int64 `path:"id"`
	}
	ProductDetailResp {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       int64  `json:"price"`
		Stock       int64  `json:"stock"`
		CategoryId  int64  `json:"categoryId"`
		Images      string `json:"images"`
		Status      int32  `json:"status"`
		CreateTime  int64  `json:"createTime"`
	}

	ProductListReq {
		CategoryId int64 `form:"categoryId,optional"`
		Page       int32 `form:"page,optional"`
		PageSize   int32 `form:"pageSize,optional"`
	}
	ProductListResp {
		Products []ProductDetailResp `json:"products"`
		Total    int64               `json:"total"`
	}

	ProductSearchReq {
		Keyword  string `form:"keyword"`
		Page     int32  `form:"page,optional"`
		PageSize int32  `form:"pageSize,optional"`
	}
)

type (
	CreateOrderItem {
		ProductId   int64  `json:"productId"`
		ProductName string `json:"productName"`
		Price       int64  `json:"price"`
		Quantity    int32  `json:"quantity"`
	}
	CreateOrderReq {
		Items []CreateOrderItem `json:"items"`
	}
	CreateOrderResp {
		Id          int64  `json:"id"`
		OrderNo     string `json:"orderNo"`
		TotalAmount int64  `json:"totalAmount"`
	}

	OrderDetailReq {
		Id int64 `path:"id"`
	}
	OrderDetailResp {
		Id          int64             `json:"id"`
		OrderNo     string            `json:"orderNo"`
		UserId      int64             `json:"userId"`
		TotalAmount int64             `json:"totalAmount"`
		Status      int32             `json:"status"`
		Items       []CreateOrderItem `json:"items"`
		CreateTime  int64             `json:"createTime"`
	}

	OrderListReq {
		Status   int32 `form:"status,optional"`
		Page     int32 `form:"page,optional"`
		PageSize int32 `form:"pageSize,optional"`
	}
	OrderListResp {
		Orders []OrderDetailResp `json:"orders"`
		Total  int64             `json:"total"`
	}
)

type (
	CartAddReq {
		ProductId int64 `json:"productId"`
		Quantity  int32 `json:"quantity"`
	}
	CartRemoveReq {
		ProductId int64 `json:"productId"`
	}
	CartItem {
		ProductId int64 `json:"productId"`
		Quantity  int32 `json:"quantity"`
		Selected  bool  `json:"selected"`
	}
	CartListResp {
		Items []CartItem `json:"items"`
	}
	CartUpdateQuantityReq {
		ProductId int64 `json:"productId"`
		Quantity  int32 `json:"quantity"`
	}
)

type (
	CreatePaymentReq {
		OrderNo string `json:"orderNo"`
		Amount  int64  `json:"amount"`
		PayType int32  `json:"payType"`
	}
	CreatePaymentResp {
		Id        int64  `json:"id"`
		PaymentNo string `json:"paymentNo"`
	}

	PaymentStatusReq {
		Id int64 `path:"id"`
	}
	PaymentStatusResp {
		Id        int64  `json:"id"`
		PaymentNo string `json:"paymentNo"`
		OrderNo   string `json:"orderNo"`
		Amount    int64  `json:"amount"`
		Status    int32  `json:"status"`
		PayType   int32  `json:"payType"`
		PayTime   int64  `json:"payTime"`
	}
)

// ===== Routes =====

@server (
	prefix: /api/user
)
service mall-api {
	@handler Register
	post /register (RegisterReq) returns (RegisterResp)

	@handler Login
	post /login (LoginReq) returns (LoginResp)
}

@server (
	prefix: /api/user
	jwt:    Auth
)
service mall-api {
	@handler UserInfo
	get /info returns (UserInfoResp)

	@handler UpdateUser
	put /update (UpdateUserReq) returns ()
}

@server (
	prefix: /api/product
)
service mall-api {
	@handler ProductDetail
	get /detail/:id (ProductDetailReq) returns (ProductDetailResp)

	@handler ProductList
	get /list (ProductListReq) returns (ProductListResp)

	@handler ProductSearch
	get /search (ProductSearchReq) returns (ProductListResp)
}

@server (
	prefix: /api/order
	jwt:    Auth
)
service mall-api {
	@handler CreateOrder
	post /create (CreateOrderReq) returns (CreateOrderResp)

	@handler OrderDetail
	get /detail/:id (OrderDetailReq) returns (OrderDetailResp)

	@handler OrderList
	get /list (OrderListReq) returns (OrderListResp)
}

@server (
	prefix: /api/cart
	jwt:    Auth
)
service mall-api {
	@handler CartAdd
	post /add (CartAddReq) returns ()

	@handler CartRemove
	post /remove (CartRemoveReq) returns ()

	@handler CartList
	get /list returns (CartListResp)

	@handler CartClear
	post /clear returns ()

	@handler CartUpdateQuantity
	post /update-quantity (CartUpdateQuantityReq) returns ()
}

@server (
	prefix: /api/payment
	jwt:    Auth
)
service mall-api {
	@handler CreatePayment
	post /create (CreatePaymentReq) returns (CreatePaymentResp)

	@handler PaymentStatus
	get /status/:id (PaymentStatusReq) returns (PaymentStatusResp)
}
```

- [ ] **Step 2: Generate API code with goctl**

```bash
export PATH=$PATH:/home/carter/workspace/go/bin
cd /home/carter/workspace/go/go-zero
mkdir -p mall-api
cd mall-api
go mod init mall-api
goctl api go --api mall.api --dir .
```

- [ ] **Step 3: Configure etc/mall-api.yaml**

```yaml
Name: mall-api
Host: 0.0.0.0
Port: 8888

Auth:
  AccessSecret: "mall-secret-key-change-in-production"
  AccessExpire: 604800

UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: user.rpc

ProductRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: product.rpc

OrderRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: order.rpc

CartRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: cart.rpc

PaymentRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: payment.rpc
```

- [ ] **Step 4: Update internal/config/config.go**

```go
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
	UserRpc    zrpc.RpcClientConf
	ProductRpc zrpc.RpcClientConf
	OrderRpc   zrpc.RpcClientConf
	CartRpc    zrpc.RpcClientConf
	PaymentRpc zrpc.RpcClientConf
}
```

- [ ] **Step 5: Update internal/svc/servicecontext.go**

```go
package svc

import (
	"mall-api/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
	"mall-user-rpc/user"
	"mall-product-rpc/product"
	"mall-order-rpc/order"
	"mall-cart-rpc/cart"
	"mall-payment-rpc/payment"
)

type ServiceContext struct {
	Config     config.Config
	UserRpc    user.UserClient
	ProductRpc product.ProductClient
	OrderRpc   order.OrderClient
	CartRpc    cart.CartClient
	PaymentRpc payment.PaymentClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		UserRpc:    user.NewUserClient(zrpc.MustNewClient(c.UserRpc).Conn()),
		ProductRpc: product.NewProductClient(zrpc.MustNewClient(c.ProductRpc).Conn()),
		OrderRpc:   order.NewOrderClient(zrpc.MustNewClient(c.OrderRpc).Conn()),
		CartRpc:    cart.NewCartClient(zrpc.MustNewClient(c.CartRpc).Conn()),
		PaymentRpc: payment.NewPaymentClient(zrpc.MustNewClient(c.PaymentRpc).Conn()),
	}
}
```

Note: The API gateway's go.mod must use `replace` directives to point to the local RPC service directories for the generated gRPC client code:

```
replace (
    mall-user-rpc => ../mall-user-rpc
    mall-product-rpc => ../mall-product-rpc
    mall-order-rpc => ../mall-order-rpc
    mall-cart-rpc => ../mall-cart-rpc
    mall-payment-rpc => ../mall-payment-rpc
)
```

- [ ] **Step 6: Implement handler logic — registerlogic.go**

```go
package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	res, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
	})
	if err != nil {
		return nil, err
	}

	return &types.RegisterResp{Id: res.Id}, nil
}
```

- [ ] **Step 7: Implement loginlogic.go**

```go
package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	res, err := l.svcCtx.UserRpc.Login(l.ctx, &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &types.LoginResp{Id: res.Id, Token: res.Token}, nil
}
```

- [ ] **Step 8: Implement userinfologic.go (extracts uid from JWT)**

```go
package logic

import (
	"context"
	"encoding/json"

	"mall-api/internal/svc"
	"mall-api/internal/types"

	"mall-user-rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserInfoLogic {
	return &UserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserInfoLogic) UserInfo() (resp *types.UserInfoResp, err error) {
	uid, _ := l.ctx.Value("uid").(json.Number)
	userId, _ := uid.Int64()

	res, err := l.svcCtx.UserRpc.GetUser(l.ctx, &user.GetUserReq{Id: userId})
	if err != nil {
		return nil, err
	}

	return &types.UserInfoResp{
		Id:         res.Id,
		Username:   res.Username,
		Phone:      res.Phone,
		Avatar:     res.Avatar,
		CreateTime: res.CreateTime,
	}, nil
}
```

- [ ] **Step 9: Implement remaining handler logic files**

Each handler follows the same pattern: extract JWT uid from context where needed, call the corresponding RPC client method, and map the response to the API types. The remaining logic files are:

- `updateuserlogic.go` — calls `UserRpc.UpdateUser`
- `productdetaillogic.go` — calls `ProductRpc.GetProduct`
- `productlistlogic.go` — calls `ProductRpc.ListProducts`
- `productsearchlogic.go` — calls `ProductRpc.SearchProducts`
- `createorderlogic.go` — calls `OrderRpc.CreateOrder` with uid from JWT
- `orderdetaillogic.go` — calls `OrderRpc.GetOrder`
- `orderlistlogic.go` — calls `OrderRpc.ListOrders` with uid from JWT
- `cartaddlogic.go` — calls `CartRpc.AddItem` with uid from JWT
- `cartremovelogic.go` — calls `CartRpc.RemoveItem` with uid from JWT
- `cartlistlogic.go` — calls `CartRpc.ListItems` with uid from JWT
- `cartclearlogic.go` — calls `CartRpc.ClearCart` with uid from JWT
- `cartupdatequantitylogic.go` — calls `CartRpc.UpdateQuantity` with uid from JWT
- `createpaymentlogic.go` — calls `PaymentRpc.CreatePayment` with uid from JWT
- `paymentstatuslogic.go` — calls `PaymentRpc.GetPayment`

Each one follows the same pattern as registerlogic/loginlogic/userinfologic. For each file:
1. Extract uid from JWT context (for auth-required endpoints): `uid, _ := l.ctx.Value("uid").(json.Number); userId, _ := uid.Int64()`
2. Call the RPC method with request params
3. Map RPC response to API types and return

- [ ] **Step 10: Run go mod tidy and verify build**

```bash
cd /home/carter/workspace/go/go-zero/mall-api
go mod tidy
go build ./...
```

- [ ] **Step 11: Commit**

```bash
cd /home/carter/workspace/go/go-zero
git add mall-api/
git commit -m "feat: add mall-api gateway with routes for all services"
```

---

## Task 8: Verify full project structure and cross-service build

- [ ] **Step 1: Verify directory structure**

```bash
cd /home/carter/workspace/go/go-zero
find . -name "go.mod" -exec echo {} \; -exec head -1 {} \;
```

Expected: 7 go.mod files (mall-common, mall-user-rpc, mall-product-rpc, mall-order-rpc, mall-cart-rpc, mall-payment-rpc, mall-api).

- [ ] **Step 2: Build all services**

```bash
cd /home/carter/workspace/go/go-zero
for dir in mall-common mall-user-rpc mall-product-rpc mall-order-rpc mall-cart-rpc mall-payment-rpc mall-api; do
  echo "=== Building $dir ==="
  (cd $dir && go build ./...)
done
```

Expected: all 7 directories build without errors.

- [ ] **Step 3: Commit final state**

```bash
cd /home/carter/workspace/go/go-zero
git add -A
git commit -m "feat: complete mall microservices project with all services verified"
```
