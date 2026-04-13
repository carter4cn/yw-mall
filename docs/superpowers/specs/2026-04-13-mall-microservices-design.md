# Mall 电商微服务系统设计

## 概述

基于 go-zero 框架的电商微服务系统，采用多 Repo 架构。包含 1 个 HTTP API 网关 + 5 个 gRPC RPC 服务 + 1 个共享库，使用 MySQL + Redis + Etcd。

## 架构

### 仓库列表

| 仓库 | 职责 | 端口 |
|------|------|------|
| `mall-api` | HTTP API 网关，接收客户端请求，转发给 RPC 服务 | 8888 |
| `mall-user-rpc` | 用户注册、登录、信息管理 | 9001 |
| `mall-product-rpc` | 商品 CRUD、分类、搜索 | 9002 |
| `mall-order-rpc` | 订单创建、查询、状态管理 | 9003 |
| `mall-cart-rpc` | 购物车增删改查 | 9004 |
| `mall-payment-rpc` | 支付创建、回调、状态查询 | 9005 |
| `mall-common` | 共享 proto 文件、公共工具库、错误码定义 | — |

### 服务调用关系

```
客户端 → mall-api → user-rpc
                   → product-rpc
                   → order-rpc → product-rpc (校验库存)
                                → cart-rpc (清空购物车)
                                → payment-rpc (发起支付)
                   → cart-rpc → product-rpc (查询商品信息)
                   → payment-rpc
```

### 基础设施

- **Etcd**: 服务注册与发现
- **MySQL**: 每个 RPC 服务独立数据库（mall_user、mall_product、mall_order、mall_cart、mall_payment）
- **Redis**: 缓存、会话、分布式锁

## mall-common 共享库

```
mall-common/
├── proto/
│   ├── user/user.proto
│   ├── product/product.proto
│   ├── order/order.proto
│   ├── cart/cart.proto
│   └── payment/payment.proto
├── errorx/           # 统一错误码定义
│   └── errorx.go
├── result/           # 统一 HTTP 响应格式
│   └── response.go
└── interceptor/      # gRPC 拦截器（日志、鉴权等）
    └── auth.go
```

Proto 管理策略：所有 .proto 文件集中在 mall-common，各 RPC 服务引用并生成代码。

## 数据模型

### user-rpc（数据库: mall_user）

| 表 | 关键字段 |
|---|---------|
| `user` | id, username, password (bcrypt), phone, avatar, create_time |

核心接口：Register、Login、GetUser、UpdateUser

### product-rpc（数据库: mall_product）

| 表 | 关键字段 |
|---|---------|
| `category` | id, name, parent_id, sort |
| `product` | id, name, description, price, stock, category_id, images, status, create_time |

核心接口：CreateProduct、GetProduct、ListProducts、UpdateStock、SearchProducts

### order-rpc（数据库: mall_order）

| 表 | 关键字段 |
|---|---------|
| `order` | id, order_no (雪花算法), user_id, total_amount, status, create_time |
| `order_item` | id, order_id, product_id, product_name, price, quantity |

核心接口：CreateOrder、GetOrder、ListOrders、UpdateOrderStatus

订单状态流转：待支付 → 已支付 → 已发货 → 已完成 / 已取消

### cart-rpc（数据库: mall_cart）

| 表 | 关键字段 |
|---|---------|
| `cart_item` | id, user_id, product_id, quantity, selected, create_time |

核心接口：AddItem、RemoveItem、ListItems、ClearCart、UpdateQuantity

### payment-rpc（数据库: mall_payment）

| 表 | 关键字段 |
|---|---------|
| `payment` | id, payment_no, order_no, user_id, amount, status, pay_type, pay_time |

核心接口：CreatePayment、GetPayment、UpdatePaymentStatus

支付状态：待支付 → 支付成功 / 支付失败

## mall-api 网关

### 路由

```
/api/user/      → register, login, info, update
/api/product/   → list, detail, search
/api/order/     → create, list, detail
/api/cart/      → add, remove, list, clear
/api/payment/   → create, status
```

### 中间件

- JWT 鉴权（登录/注册接口除外）
- 统一响应格式 `{ code, msg, data }`

网关只做参数校验 + 转发，不含业务逻辑。

## 各服务目录结构（以 user-rpc 为例）

```
mall-user-rpc/
├── etc/
│   └── user.yaml          # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go      # 配置结构
│   ├── logic/
│   │   ├── registerlogic.go
│   │   ├── loginlogic.go
│   │   ├── getuserlogic.go
│   │   └── updateuserlogic.go
│   ├── model/
│   │   └── usermodel.go   # 数据模型
│   ├── server/
│   │   └── userserver.go  # gRPC server 实现
│   └── svc/
│       └── servicecontext.go
├── user.go                 # 入口
├── go.mod
└── go.sum
```

## 技术要点

- goctl 生成 API 和 RPC 代码骨架
- go-zero 内置 sqlx 操作 MySQL，支持缓存
- go-zero 内置 Redis 客户端
- JWT 使用 go-zero 内置 jwt 中间件
- 密码使用 bcrypt 加密
- 订单号使用雪花算法生成
- 各服务通过 Etcd 注册，API 网关通过 Etcd 发现 RPC 服务
