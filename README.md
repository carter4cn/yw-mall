# yw-mall

基于 [go-zero](https://github.com/zeromicro/go-zero) 的电商商城微服务系统。覆盖用户、商品、订单、购物车、支付，以及营销侧的活动、规则引擎、奖励、风控、工作流编排等场景，配套一键启停脚本与本地 docker-compose 基础设施。

## 技术栈

- **语言 / 框架**：Go 1.26.1，go-zero 1.10.1（HTTP API + gRPC）
- **服务发现**：etcd
- **存储 / 缓存**：MySQL（经 ProxySQL 接入）、Redis
- **消息**：Kafka（活动异步流水线）
- **分布式事务**：DTM
- **协议定义**：Protocol Buffers（`mall-common/proto/`）+ go-zero `.api`

## 仓库结构

```
yw-mall/
├── mall-api/                     HTTP API 网关（聚合所有 RPC）
├── mall-user-rpc/                用户 / 注册 / 登录 / JWT
├── mall-product-rpc/             商品 CRUD / 库存 / 搜索
├── mall-order-rpc/               订单创建 / 详情 / 列表 / 状态
├── mall-cart-rpc/                购物车增删改查 / 清空
├── mall-payment-rpc/             支付创建 / 状态查询
├── mall-activity-rpc/            活动（优惠券 / 秒杀 / 抽奖 / 签到）
├── mall-activity-async-worker/   活动异步 worker（Kafka 消费侧）
├── mall-reward-rpc/              奖励发放与查询
├── mall-risk-rpc/                风控（HMAC Token / 校验）
├── mall-rule-rpc/                规则引擎
├── mall-workflow-rpc/            工作流编排（串联活动 / 规则 / 奖励）
├── mall-common/                  共享 proto、errorx 错误码
├── docs/                         设计文档
└── start.sh                      一键启停脚本
```

## 服务端口

| 服务 | 类型 | 端口 |
|------|------|------|
| mall-api | HTTP | 18888 |
| mall-user-rpc | gRPC | 19001 |
| mall-product-rpc | gRPC | 9002 |
| mall-order-rpc | gRPC | 9003 |
| mall-cart-rpc | gRPC | 9004 |
| mall-payment-rpc | gRPC | 9005 |
| mall-activity-rpc | gRPC | 9010 |
| mall-rule-rpc | gRPC | 9011 |
| mall-workflow-rpc | gRPC | 9012 |
| mall-reward-rpc | gRPC | 9013 |
| mall-risk-rpc | gRPC | 9014 |
| mall-activity-async-worker | worker | — |

## 快速开始

依赖：`go 1.26+`、`docker` 或 `podman`、`docker compose`。

```bash
# 一键启动：拉起基础设施 → 自动 bootstrap（建库 / DDL / 种子数据）→ 启动全部 12 个服务
./start.sh

# 其它常用命令
./start.sh status      # 查看进程状态
./start.sh stop        # 停止 go 服务（保留 compose 基础设施）
./start.sh restart     # 重启
./start.sh bootstrap   # 仅重新初始化 DB / 种子数据
./start.sh nuke        # 停服 + 清空 mall_* 数据库 + 刷新 Redis（重建用）
```

启动成功后，HTTP 网关在 `http://127.0.0.1:18888`。

## 主要 HTTP 接口

| 模块 | 路径 |
|------|------|
| 用户 | `POST /api/user/register` `POST /api/user/login` `GET /api/user/info` `PUT /api/user/info` |
| 商品 | `GET /api/product/:id` `GET /api/product/list` `GET /api/product/search` |
| 购物车 | `POST /api/cart/add` `POST /api/cart/remove` `GET /api/cart/list` `POST /api/cart/clear` `POST /api/cart/update` |
| 订单 | `POST /api/order/create` `GET /api/order/:id` `GET /api/order/list` |
| 支付 | `POST /api/payment/create` `GET /api/payment/status` |
| 活动 | `GET /api/activity/list` `GET /api/activity/:id` `POST /api/activity/participate` `POST /api/activity/coupon/claim` `POST /api/activity/seckill/buy` `POST /api/activity/lottery/spin` `POST /api/activity/signin` |
| 奖励 | `GET /api/reward/my` |

## 开发约定

- proto 定义统一放在 `mall-common/proto/<domain>/`，由各 RPC 服务通过 go-zero `goctl` 生成代码。
- 错误码集中在 `mall-common/errorx/`。
- 配置文件在每个服务的 `etc/*.yaml`，DB / Redis / etcd 地址默认指向本地 docker-compose 起的实例。
- 仓库内的 `AccessSecret` 等密钥均为占位值（`*-change-in-production`），生产环境请覆盖。

## 文档

更多设计细节见 [`docs/`](./docs)。
