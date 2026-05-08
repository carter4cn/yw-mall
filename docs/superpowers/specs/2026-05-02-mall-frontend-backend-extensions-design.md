# C 端前端后端扩展（子项目 1/4）设计文档

**日期**：2026-05-02
**子项目**：mall 前端化项目 · 第 1 个子项目「后端扩展」
**目标**：为 C 端 H5 前端铺好后端能力——新增店铺服务、扩展用户地址、订单地址快照、统一商品图托管、可重复的 demo 数据 seed 管道。

> 这是 C 端前端整体规划的第 1 个子项目。整体分 4 个：
>
> 1. **后端扩展**（本文档）
> 2. 前端基础（uni-app 脚手架 + 设计令牌 + 浏览域）
> 3. 前端交易链路 + 账户
> 4. 前端营销活动
>
> 后续每个子项目独立 brainstorm + spec + plan + 实施。

---

## §1 架构概览

### 改动面（按服务）

| 服务 | 类型 | 说明 |
|------|------|------|
| **mall-shop-rpc** | 🆕 新建 | 店铺 CRUD + 列表 + 详情 + 关注 |
| **mall-product-rpc** | 扩展 | product 表加 `shop_id`；ListProducts/Search 透传 shop_id；新增 ListShopProducts |
| **mall-user-rpc** | 扩展 | 新表 `user_address`；5 RPC：Add/Update/Delete/SetDefault/List + GetDefault + GetAddress |
| **mall-order-rpc** | 扩展 | order 表加 receiver 快照（`receiver_name` 等 7 字段）；CreateOrder 接收 `address_id`，调 user-rpc.GetAddress 写入快照 |
| **mall-api** | 扩展 | 暴露店铺/地址新路由；OrderDetail 透出 receiver 快照 |
| **seed pipeline** | 🆕 新建 | 各 RPC 加 `cmd/seed/main.go`；从 picsum/dicebear 拉图上传 MinIO；start.sh 接入 |
| **errorx** | 扩展 | 新增 user/order/shop 错误码 |
| **minioutil** | 🆕 提取 | 把 review-rpc 中 MinIO 上传逻辑抽到 `mall-common/minioutil/` 复用 |

### 端口分配

- mall-shop-rpc → **9017**

### MinIO bucket 复用

复用现有 `mall-media` bucket，前缀拆分：
- `reviews/` （已用，review-rpc 上传媒体）
- `products/seed/` （新）
- `shops/seed/` （新）

### 错误码（mall-common/errorx）

```
2010 UserAddressNotFound
2011 UserAddressForbidden
2012 UserAddressLimit       // ≥20 条上限
5004 OrderAddressRequired   // CreateOrder 时 address_id == 0
6010 ShopNotFound           // 6001/6002 已被 payment 占用，shop 用 6010+
6011 ShopFollowAlreadyExists
```

---

## §2 mall-shop-rpc 详设

### 2.1 Schema (`mall-shop-rpc/sql/shop.sql`)

```sql
CREATE TABLE shop (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name          VARCHAR(64)  NOT NULL,
  logo          VARCHAR(255) NOT NULL DEFAULT '',
  banner        VARCHAR(255) NOT NULL DEFAULT '',
  description   VARCHAR(500) NOT NULL DEFAULT '',
  rating        DECIMAL(3,2) NOT NULL DEFAULT 5.00,
  product_count INT          NOT NULL DEFAULT 0,
  follow_count  INT          NOT NULL DEFAULT 0,
  status        TINYINT      NOT NULL DEFAULT 1,
  create_time   BIGINT       NOT NULL,
  update_time   BIGINT       NOT NULL,
  KEY idx_status (status),
  KEY idx_rating (rating DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE TABLE shop_follow (
  id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id     BIGINT UNSIGNED NOT NULL,
  shop_id     BIGINT UNSIGNED NOT NULL,
  create_time BIGINT          NOT NULL,
  UNIQUE KEY uk_user_shop (user_id, shop_id),
  KEY idx_shop (shop_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

### 2.2 Proto (`mall-common/proto/shop/shop.proto`)

```proto
syntax = "proto3";
package shop;
option go_package = "./shop";

service Shop {
  rpc GetShop(GetShopReq) returns (GetShopResp);
  rpc ListShops(ListShopsReq) returns (ListShopsResp);
  rpc ListRecommendedShops(ListRecommendedShopsReq) returns (ListShopsResp);
  rpc FollowShop(FollowShopReq) returns (OkResp);
  rpc UnfollowShop(UnfollowShopReq) returns (OkResp);
  rpc IsFollowing(IsFollowingReq) returns (IsFollowingResp);
  rpc ListFollowedShops(ListFollowedShopsReq) returns (ListShopsResp);

  // 仅内部用
  rpc CreateShop(CreateShopReq) returns (CreateShopResp);
  rpc UpdateShop(UpdateShopReq) returns (OkResp);
  rpc IncrProductCount(IncrProductCountReq) returns (OkResp);
}

message Shop {
  int64  id            = 1;
  string name          = 2;
  string logo          = 3;
  string banner        = 4;
  string description   = 5;
  double rating        = 6;
  int32  product_count = 7;
  int32  follow_count  = 8;
  int32  status        = 9;
  int64  create_time   = 10;
}

message OkResp { bool ok = 1; }

message GetShopReq           { int64 id = 1; }
message GetShopResp          { Shop shop = 1; }
message ListShopsReq         { int32 page = 1; int32 page_size = 2; }
message ListShopsResp        { repeated Shop shops = 1; int64 total = 2; }
message ListRecommendedShopsReq { int32 limit = 1; }
message FollowShopReq        { int64 user_id = 1; int64 shop_id = 2; }
message UnfollowShopReq      { int64 user_id = 1; int64 shop_id = 2; }
message IsFollowingReq       { int64 user_id = 1; int64 shop_id = 2; }
message IsFollowingResp      { bool is_following = 1; }
message ListFollowedShopsReq { int64 user_id = 1; int32 page = 2; int32 page_size = 3; }
message CreateShopReq        { string name = 1; string logo = 2; string banner = 3; string description = 4; double rating = 5; }
message CreateShopResp       { int64 id = 1; }
message UpdateShopReq        { int64 id = 1; string name = 2; string logo = 3; string banner = 4; string description = 5; }
message IncrProductCountReq  { int64 shop_id = 1; int32 delta = 2; }
```

### 2.3 ServiceContext

- DB：sqlx.SqlConn (ProxySQL 6033, `mall_shop` 库)
- Redis：`cache:shop:{id}` 走 go-zero CachedConn 自动缓存
- 不依赖其他 RPC

### 2.4 关键决策

- `rating` / `product_count` / `follow_count` 走**冗余字段**而非每次 JOIN 查；写路径主动维护
- FollowShop 用 `INSERT ... ON DUPLICATE KEY` + `UPDATE shop SET follow_count = follow_count + 1`，事务包裹（DUP 时不增计数）
- UnfollowShop 同样事务包裹（DELETE 命中 1 行才减计数，避免双减）
- IncrProductCount 在 product-rpc.CreateProduct 末尾调用；product 删除暂不做（demo 不删商品）
- 不做评分汇总同步（rating 在 seed 时写死，后续可加定时任务，本期不做）

---

## §3 mall-user-rpc 地址扩展

### 3.1 Schema (追加到 `mall-user-rpc/sql/user.sql`)

```sql
CREATE TABLE user_address (
  id            BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id       BIGINT UNSIGNED NOT NULL,
  receiver_name VARCHAR(32)     NOT NULL,
  phone         VARCHAR(20)     NOT NULL,
  province      VARCHAR(32)     NOT NULL,
  city          VARCHAR(32)     NOT NULL,
  district      VARCHAR(32)     NOT NULL,
  detail        VARCHAR(255)    NOT NULL,
  is_default    TINYINT         NOT NULL DEFAULT 0,
  create_time   BIGINT          NOT NULL,
  update_time   BIGINT          NOT NULL,
  KEY idx_user (user_id, is_default DESC, update_time DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

### 3.2 Proto (追加到 `mall-common/proto/user/user.proto`)

```proto
service User {
  // ... 已有
  rpc AddAddress(AddAddressReq) returns (AddAddressResp);
  rpc UpdateAddress(UpdateAddressReq) returns (OkResp);
  rpc DeleteAddress(DeleteAddressReq) returns (OkResp);
  rpc SetDefaultAddress(SetDefaultAddressReq) returns (OkResp);
  rpc ListAddresses(ListAddressesReq) returns (ListAddressesResp);
  rpc GetAddress(GetAddressReq) returns (Address);
  rpc GetDefaultAddress(GetDefaultAddressReq) returns (Address);
}

message Address {
  int64  id            = 1;
  int64  user_id       = 2;
  string receiver_name = 3;
  string phone         = 4;
  string province      = 5;
  string city          = 6;
  string district      = 7;
  string detail        = 8;
  bool   is_default    = 9;
  int64  create_time   = 10;
}

message AddAddressReq        { int64 user_id = 1; string receiver_name = 2; string phone = 3; string province = 4; string city = 5; string district = 6; string detail = 7; bool is_default = 8; }
message AddAddressResp       { int64 id = 1; }
message UpdateAddressReq     { int64 user_id = 1; int64 id = 2; string receiver_name = 3; string phone = 4; string province = 5; string city = 6; string district = 7; string detail = 8; }
message DeleteAddressReq     { int64 user_id = 1; int64 id = 2; }
message SetDefaultAddressReq { int64 user_id = 1; int64 id = 2; }
message ListAddressesReq     { int64 user_id = 1; }
message ListAddressesResp    { repeated Address addresses = 1; }
message GetAddressReq        { int64 user_id = 1; int64 id = 2; }   // user_id 用于校验
message GetDefaultAddressReq { int64 user_id = 1; }
```

### 3.3 关键逻辑

- **SetDefaultAddress**：单事务内：
  ```sql
  UPDATE user_address SET is_default=0 WHERE user_id=?;
  UPDATE user_address SET is_default=1 WHERE id=? AND user_id=?;
  ```
- **AddAddress 默认地址决策**（事务内 4 步）：
  1. SELECT COUNT(*) FROM user_address WHERE user_id=?；若 ≥20 报 `UserAddressLimit`
  2. 计算最终 `effective_default`：`req.is_default == true || count == 0`
  3. 若 `effective_default == true`：`UPDATE user_address SET is_default=0 WHERE user_id=?`
  4. INSERT 新行，`is_default = effective_default`
- **DeleteAddress 默认地址处理**：删的若是默认地址，事务内自动把"剩下最近 update_time 最大"的那条设为默认；事务包裹
- **GetAddress / DeleteAddress / UpdateAddress / SetDefaultAddress 强制 user_id 校验**：req 必带 user_id（mall-api 注入），与 address.user_id 不符直接报 `UserAddressForbidden`
- **GetDefaultAddress 可能返回空**：按 protobuf 习惯返回 `&Address{Id: 0}`（前端按 id == 0 判断"无默认地址"）；不返错

---

## §4 product-rpc & order-rpc 扩展

### 4.1 product-rpc

**Schema diff** (`mall-product-rpc/sql/product.sql` ALTER)

```sql
ALTER TABLE product
  ADD COLUMN shop_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ADD KEY idx_shop_status (shop_id, status, id DESC);
```

**Proto diff** (`mall-common/proto/product/product.proto`)

```proto
service Product {
  // ... 已有
  rpc ListShopProducts(ListShopProductsReq) returns (ListProductsResp);
}

message Product {
  // ... 已有字段
  int64 shop_id = 11;
}

message ListShopProductsReq {
  int64 shop_id   = 1;
  int32 page      = 2;
  int32 page_size = 3;
}
```

**逻辑变化**：
- `CreateProduct` 接收 `shop_id`，事务内 `INSERT product` + 调 `ShopRpc.IncrProductCount(shop_id, delta=1)`
- `GetProduct` / `ListProducts` / `Search` 透出 `shop_id`
- `ListShopProducts` 走 `idx_shop_status` 索引
- ServiceContext 注入 `ShopRpc` client（依赖 mall-shop-rpc 已启动）

### 4.2 order-rpc

**Schema diff** (`mall-order-rpc/sql/order.sql` ALTER)

```sql
ALTER TABLE `order`
  ADD COLUMN address_id        BIGINT UNSIGNED NOT NULL DEFAULT 0,
  ADD COLUMN receiver_name     VARCHAR(32)  NOT NULL DEFAULT '',
  ADD COLUMN receiver_phone    VARCHAR(20)  NOT NULL DEFAULT '',
  ADD COLUMN receiver_province VARCHAR(32)  NOT NULL DEFAULT '',
  ADD COLUMN receiver_city     VARCHAR(32)  NOT NULL DEFAULT '',
  ADD COLUMN receiver_district VARCHAR(32)  NOT NULL DEFAULT '',
  ADD COLUMN receiver_detail   VARCHAR(255) NOT NULL DEFAULT '';
```

**Proto diff**

```proto
message CreateOrderReq {
  int64 user_id           = 1;
  int64 address_id        = 2;
  repeated OrderItem items = 3;
}

message GetOrderResp {
  // 已有字段...
  int64  address_id        = 100;
  string receiver_name     = 101;
  string receiver_phone    = 102;
  string receiver_province = 103;
  string receiver_city     = 104;
  string receiver_district = 105;
  string receiver_detail   = 106;
}
```

**关键逻辑**：
- `CreateOrder` 事务前先调 `UserRpc.GetAddress(user_id, address_id)`：
  - `address_id == 0` → `OrderAddressRequired`
  - 返回 `UserAddressNotFound` / `UserAddressForbidden` 直接透传
  - 成功 → 把 6 个 receiver 字段写入快照 + 保留 `address_id` 用于回溯
- 已有订单**不补**快照（旧 demo 数据可接受空地址）
- ServiceContext 加 `UserRpc` client

### 4.3 mall-api 网关扩展

**新路由**：

```api
// shop 公开
@server (prefix: /api/shop)
service mall-api {
  @handler ShopDetail
  get /detail/:id (ShopDetailReq) returns (ShopDetailResp)

  @handler ShopList
  get /list (ShopListReq) returns (ShopListResp)

  @handler ShopRecommended
  get /recommended returns (ShopListResp)

  @handler ShopProducts
  get /products/:id (ShopProductsReq) returns (ProductListResp)
}

// shop follow + address (jwt)
@server (prefix: /api, jwt: Auth)
service mall-api {
  @handler FollowShop
  post /shop/:id/follow (FollowShopReq) returns (OkResp)
  @handler UnfollowShop
  post /shop/:id/unfollow (UnfollowShopReq) returns (OkResp)
  @handler IsFollowingShop
  get /shop/:id/is-following (IsFollowingShopReq) returns (IsFollowingShopResp)
  @handler MyFollowedShops
  get /shop/my-followed (PageReq) returns (ShopListResp)

  @handler AddressAdd
  post /address/add (AddAddressReq) returns (AddAddressResp)
  @handler AddressUpdate
  post /address/update (UpdateAddressReq) returns (OkResp)
  @handler AddressDelete
  post /address/delete (DeleteAddressReq) returns (OkResp)
  @handler AddressSetDefault
  post /address/set-default (SetDefaultAddressReq) returns (OkResp)
  @handler AddressList
  get /address/list returns (AddressListResp)
}
```

**OrderDetail 修改**：types 加 7 个 receiver 字段；CreateOrder req types 加 `addressId int64`。

---

## §5 Seed 管道

### 5.1 工程结构

```
mall-shop-rpc/cmd/seed/main.go      # 6 家 shop + logo + banner
mall-product-rpc/cmd/seed/main.go   # 40 商品 + 主图 + 详情图，按 shop_id 分配
mall-order-rpc/cmd/seed/main.go     # 20 订单（含部分已发货）
mall-review-rpc/cmd/seed/main.go    # 30 评价（含图/视频）— 已存在则扩展
mall-logistics-rpc/cmd/seed/main.go # 12 物流（含 4-6 节点）— 已存在则扩展
```

### 5.2 图片来源 + 缓存

**两阶段镜像**：
1. 构建期下载到 `/tmp/mall-seed-cache/`（带本地缓存）
2. 上传到 MinIO `mall-media` bucket
3. SQL 写 MinIO public URL（`http://localhost:9000/mall-media/<key>`）

**外链来源**（仅用稳定的）：
- 商品图：`https://picsum.photos/seed/<sku>/600/600`
- 详情大图：`https://picsum.photos/seed/<sku>-detail-<n>/800/800`
- shop logo：`https://api.dicebear.com/7.x/shapes/svg?seed=<shop-slug>`
- shop banner：`https://picsum.photos/seed/banner-<shop-id>/800/300`

**离线性**：首次需联网；缓存命中后 100% 离线可演示。

### 5.3 数据规模

| 实体 | 数量 | 备注 |
|------|------|------|
| shop | 6 | 科技数码 / 服饰潮流 / 家居生活 / 美妆个护 / 食品生鲜 / 运动户外 |
| product | 40 | 每店 6-8 个；价格 19.9 - 8999；按品类合理化命名 |
| user | 复用现有 + 补充 | alice/bob/demo 等已 seed；地址在 user-rpc seed 时给每人加 1-2 条 |
| user_address | 10-15 | 北上广深 + 几个二线 |
| order | 20 | 5 状态分布；部分有地址（新版）部分无（兼容旧） |
| review | 30 | 挂在已完成订单上；含图/视频 |
| logistics | 12 | 挂在已发货订单上；4-6 个 track 节点 |

### 5.4 幂等性

每个 seed 检查"已有数据"再决定行为：

| seed | 跳过条件 |
|------|----------|
| shop | `COUNT(*) ≥ 6` |
| product | `COUNT(*) ≥ 40` |
| user_address | `COUNT(*) ≥ 10` |
| order | `COUNT(*) ≥ 20` |
| review | `COUNT(*) ≥ 30` |
| logistics shipment | `COUNT(*) ≥ 12` |

**和 `.bootstrapped` marker 协同**：现有 marker 控制"workflow/reward/activity 是否 seed 过"；新加的 seed 也走同一 marker。

### 5.5 start.sh 接入

在现有 seed 段（workflow/reward/activity）后追加：

```bash
log "Seeding shops..."
( cd "$BASE_DIR/mall-shop-rpc"     && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
log "Seeding products..."
( cd "$BASE_DIR/mall-product-rpc"  && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
log "Seeding addresses..."
( cd "$BASE_DIR/mall-user-rpc"     && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
log "Seeding orders..."
( cd "$BASE_DIR/mall-order-rpc"    && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
log "Seeding reviews..."
( cd "$BASE_DIR/mall-review-rpc"   && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
log "Seeding logistics..."
( cd "$BASE_DIR/mall-logistics-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /'
```

DB 列表也加：`mall_shop`。

### 5.6 minioutil 提取

新建 `mall-common/minioutil/minioutil.go`：

```go
package minioutil

type Config struct {
  Endpoint   string  // "localhost:9000"
  AccessKey  string
  SecretKey  string
  Bucket     string  // "mall-media"
  PublicHost string  // "http://localhost:9000"
  UseSSL     bool
}

type Client struct{ /* private */ }

func New(cfg Config) (*Client, error)
func (c *Client) PutFile(ctx context.Context, key string, file io.Reader, contentType string) (publicURL string, err error)
func (c *Client) PutBytes(ctx context.Context, key string, data []byte, contentType string) (publicURL string, err error)
func (c *Client) Exists(ctx context.Context, key string) bool
func (c *Client) PublicURL(key string) string
```

review-rpc 现有 MinIO 上传逻辑迁移到这里；后续 seed 程序导入复用。

---

## §6 测试策略 + 风险

### 6.1 测试

| 层 | 范围 | 工具 |
|----|------|------|
| **单元** | shop-rpc Follow/Unfollow 事务、AddAddress 首条自动默认、DeleteAddress 默认地址自动转移、SetDefault 二步事务、CreateOrder 地址快照写入、IncrProductCount 原子性 | sqlx + project DB；沿用 review-rpc 测试模式 |
| **集成** | mall-api `/api/shop/*`、`/api/address/*`、CreateOrder 带 addressId | curl + jq + jwt token |
| **Seed 幂等** | seed 跑两次数据量一致 | bash 脚本 + COUNT(*) 校验 |
| **手动 E2E** | 真实操作链路 | 等子项目 2/3 接入前端时一起 |

### 6.2 风险

1. **MinIO 公网外链**：`http://localhost:9000/...` 只能本机访问；外网演示需 nginx 反代或 presigned URL（本期不做，文档提示）
2. **图片下载 ToS**：仅用 picsum + dicebear 这类开放服务；不用 unsplash（有调用限制）
3. **product.shop_id 默认 0**：现有 demo 数据 shop_id=0；seed 时把 product 数 < 40 视为"未 seed"，全清重建
4. **address 硬删**：order 已有快照，回查无依赖，硬删 OK
5. **shop.rating 脏数据**：seed 写死 4.5-5.0；后续要做评价汇总同步任务（不在本期）
6. **mall-shop-rpc 启动顺序**：start.sh SERVICES 数组要保证 shop-rpc 在 product-rpc 之前（product-rpc 启动时建立 ShopRpc 连接需要 etcd 注册到位）

### 6.3 默认决策（spec 直接定，可调整）

- 6 家 shop：`数码电器 / 潮流服饰 / 家居生活 / 美妆个护 / 优鲜食品 / 运动户外`
- 商品命名走真实化（"AirPods Pro 2 蓝牙降噪耳机" 而非 "商品 1"）
- shop_follow 仅 `uk_user_shop` + `idx_shop`（已够查）
- shop seed follow_count 50-300 随机；rating 4.5-5.0 随机
- 用户地址：每人 1-2 条；覆盖北京/上海/广州/深圳/杭州/成都
- address 上限 20（写死）
- order receiver 字段 NOT NULL DEFAULT ''
- product.shop_id 为 0 视为"无主商品"（兼容 ALTER 后未 reseed 旧数据）

---

## §7 任务列表概览

预计 **18-22 tasks**，按 7 组划分（详细列表在 plan 文档里写）：

| 组 | 数量 | 内容 |
|----|------|------|
| **A. 基础** | 2 | errorx 新增码 / minioutil helper 抽取 |
| **B. mall-shop-rpc** | 5 | proto+DDL / scaffold / svc / Get+List+Recommended / Follow+Unfollow+IsFollowing / IncrProductCount |
| **C. user-rpc 地址扩展** | 3 | proto+DDL / Add+Update+Delete logic / SetDefault+List+Get logic |
| **D. product/order-rpc 修改** | 3 | product 加 shop_id + ListShopProducts / order 加 receiver 快照 + CreateOrder 注入 |
| **E. mall-api 暴露** | 2 | shop 路由 + 4 logic / address 路由 + 5 logic |
| **F. seed + start.sh** | 4 | shop seed / product seed / order+review+logistics seed 扩展 / start.sh 接入 + DB 列表 |
| **G. E2E 手动** | 1 | 文档化验证流程（接到前端再跑） |

---

## §8 范围外（明确不做）

- 商家入驻流程 / 商家管理后台
- 多 SKU / 规格 / 库存管理（product 仍是单 SKU）
- 优惠券 / 满减规则在订单中的实际抵扣（待子项目 4）
- 物流方式选择 / 运费计算（默认包邮）
- 订单评分汇总同步任务（rating 写死）
- 订单收货地址联动地图选点（前端纯文本）
- 收藏夹 / 浏览历史
- 客服 IM
