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

### 应用层（host 进程，`go run`）

| 服务 | 类型 | 端口 | 仓库 |
|------|------|------|------|
| mall-api | HTTP | **18888** | yw-mall |
| admin-api | HTTP | **18999** | yw-mall-admin |
| mall-user-rpc | gRPC | 19001 | yw-mall |
| mall-product-rpc | gRPC | 9002 | yw-mall |
| mall-order-rpc | gRPC | 9003 | yw-mall |
| mall-cart-rpc | gRPC | 9004 | yw-mall |
| mall-payment-rpc | gRPC | 9005 | yw-mall |
| mall-activity-rpc | gRPC | 9010 | yw-mall |
| mall-rule-rpc | gRPC | 9011 | yw-mall |
| mall-workflow-rpc | gRPC | 9012 | yw-mall |
| mall-reward-rpc | gRPC | 9013 | yw-mall |
| mall-risk-rpc | gRPC | 9014 | yw-mall |
| mall-review-rpc | gRPC | 9015 | yw-mall |
| mall-logistics-rpc | gRPC | 9016 | yw-mall |
| mall-shop-rpc | gRPC | 9017 | yw-mall |
| mall-activity-async-worker | worker | — | yw-mall |

### 前端（host 进程，`pnpm dev`）

| 应用 | 端口 | 仓库 |
|------|------|------|
| 后台管理 SPA (vite) | **5173 / 5174** | yw-mall-admin-fe/admin |
| C 端 H5 (uni-app) | **5173 / 5174** | yw-mall-fe |

> 谁先起谁占 5173；vite 自动 fallback 到 5174。

### infra（podman 容器，env 仓库 `compose.yml`）

| 组件 | 端口 | 备注 |
|------|------|------|
| etcd1 | 2379 | 服务发现，HA 副本进 `etcd-ha` profile |
| kafka1 | 19092 | KRaft 单节点，HA 副本进 `kafka-ha` profile |
| mysql-master1 + slave1 | 内部 | ProxySQL 后端，HA 副本进 `mysql-ha` profile |
| ProxySQL | **6033** (SQL) / 6032 (admin) | MySQL R/W split 中间件 |
| redis-master | **6379** + sentinel 26379-26381 | go-zero 直连 master |
| MinIO | **9000** (API) / 9001 (console) | 商品/店铺/KYC 图片 |
| DTM | 36789 | 分布式事务 |
| Homer | **8888** | infra 统一仪表盘 |
| Grafana / Prometheus | 3000 / 9090 | 监控 |
| Postgres (PgBouncer) | 5432 | 仅 `pg` profile，S5 预埋待用 |
| MongoDB RS | 27017-27019 | 仅 `mongo` profile，S5 预埋待用 |

详细 infra 端口/账号见 [`env/SERVICES.md`](../env/SERVICES.md)。

## 浏览器访问入口

| 入口 | URL | 说明 |
|---|---|---|
| **C 端 H5** | http://localhost:5173 | mall-fe；登录 `alice/alice123`（或 bob、demo） |
| **Admin 后台** | http://localhost:5174 | admin-fe；登录 `admin/admin123` |
| C 端 API | http://localhost:18888 | 直接调 `/api/*`，没页面 |
| Admin API | http://localhost:18999 | 直接调 `/admin/v1/*`，没页面 |
| Infra 仪表盘 | http://localhost:8888 | 聚合 Grafana / MinIO Console / Kafka UI / Bytebase |

## 快速开始

### 前置依赖
- Go 1.26+
- Node 18+ + pnpm（前端 dev server）
- podman 或 docker（infra 容器）

### 1. 启 infra（一次性）

infra 在同级仓库 `../env/`（独立 repo `yw-mall-env`，slim 模式 baseline ≈ 5.7 GB 内存）：

```bash
cd ../env
podman compose -f compose.yml -f compose.lite.yml \
  --profile pg --profile mongo up -d
```

详细 profile 说明见 `env/README.md`。

### 2. 启 yw-mall 后端（16 个服务）

```bash
cd yw-mall
./start.sh                  # 一键：infra 健康检查 → bootstrap → 启 16 个 host 进程

# 子命令
./start.sh status           # 查看进程状态
./start.sh stop             # 停 go 服务（保留 infra）
./start.sh restart          # 重启
./start.sh bootstrap        # 仅重跑 DB / 种子
./start.sh nuke             # 停服 + 清空 mall_* + 刷 Redis（重建用）
```

完成后：
- mall-api: `http://localhost:18888`
- admin-api: `http://localhost:18999`

### 3. 启前端（开发态，可选）

两个前端独立 repo：

```bash
cd ../yw-mall-fe && pnpm install && pnpm dev:h5
# → http://localhost:5173

cd ../yw-mall-admin-fe/admin && pnpm install && pnpm dev
# → http://localhost:5174
```

### 测试账号

| 角色 | 用户名 | 密码 | 入口 |
|------|--------|------|------|
| C 端用户 | alice / bob / demo | `alice123` 等同名 + 123 | http://localhost:5173 |
| 平台管理员 | admin | `admin123` | http://localhost:5174 |

种子数据：5 用户 / 3 默认地址 / 6 店铺 / 40 商品 / 5 测试订单（含 paid / shipped / cancelled 等状态）。

### 故障排查

| 现象 | 排查 |
|------|------|
| `start.sh` 报 `compose up failed` | infra 容器名漂移；按 `env/README.md` 重起 |
| gateway 启动 panic `Redis not set` | yaml 缺 `Redis:` 块（S4 后新增）|
| user-rpc panic `cryptox.MustInit` | 缺 `MALL_FIELD_ENCRYPTION_KEY` env；start.sh 已注入开发态默认 |
| 订单列表 `Unknown column 'pay_time'` | sprint 迁移没跑；`./start.sh bootstrap` 重跑 |
| product seed `frame too large` | 旧 bug：默认端口 9001 是 MinIO；已修，git pull 后重跑 |

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
