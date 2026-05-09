# C 端前端后端扩展实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: 使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 按任务推进。所有步骤用 `- [ ]` 跟踪。

**Goal:** 为 C 端前端铺好后端：mall-shop-rpc 新建、mall-user-rpc 地址扩展、mall-order-rpc 地址快照、mall-api 网关扩展、seed 管道（含 MinIO 托管的真实图片）。

**Architecture:** 单分支 `feat/mall-frontend-backend-prep` 上累积 20 个原子 commit。新增一个 RPC 服务（端口 9017），扩展 3 个已有服务的 schema/proto。Seed 程序从 picsum/dicebear 拉公开图，本地缓存后上传 MinIO，DB 写 MinIO public URL。

**Tech Stack:** go-zero 1.10.1 / goctl / protoc / sqlx via ProxySQL 6033 / Redis CachedConn / MinIO client / protoc-gen-go.

**Spec:** `docs/superpowers/specs/2026-05-02-mall-frontend-backend-extensions-design.md`

---

## Task 1: errorx 新增码

**Files:**
- Modify: `mall-common/errorx/errorx.go`

- [ ] **Step 1: 添加新错误码常量**

在 `mall-common/errorx/errorx.go` 适当位置（按 code 分组）追加：

```go
const (
	// ... 已有

	// User address (2010-2019)
	UserAddressNotFound  = 2010
	UserAddressForbidden = 2011
	UserAddressLimit     = 2012

	// Order extensions
	OrderAddressRequired = 5004

	// Shop (6001-6009)
	ShopNotFound             = 6001
	ShopFollowAlreadyExists  = 6002
)
```

并在文件中的 `codeMessage` map（如果存在）追加对应中文：

```go
UserAddressNotFound:     "收货地址不存在",
UserAddressForbidden:    "无权操作该收货地址",
UserAddressLimit:        "收货地址数量超出上限",
OrderAddressRequired:    "请选择收货地址",
ShopNotFound:            "店铺不存在",
ShopFollowAlreadyExists: "已关注该店铺",
```

> 先 `grep -n "codeMessage\|MessageMap\|errMessages" mall-common/errorx/errorx.go` 确认 map 名字，若 errorx 仅有常量没有 message map 就只加常量。

- [ ] **Step 2: 验证编译**

```bash
cd /home/carter/workspace/go/yw-mall/mall-common && go build ./...
```

Expected: 无输出。

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-common/errorx/errorx.go
git commit -m "feat(errorx): add user address / shop / order-address codes"
```

---

## Task 2: minioutil helper 抽取

**Files:**
- Create: `mall-common/minioutil/minioutil.go`
- Create: `mall-common/minioutil/minioutil_test.go`
- Modify: `mall-review-rpc/internal/logic/uploadreviewmedialogic.go`（替换 inline 上传为 minioutil）
- Modify: `mall-review-rpc/internal/svc/servicecontext.go`（用 minioutil.Client 替换原 raw minio.Client）

- [ ] **Step 1: 探查 review-rpc 现有 MinIO 用法**

```bash
grep -rn "minio\.\|PutObject\|minio-go" mall-review-rpc/ | head -20
```

记下：
- 包导入路径（`github.com/minio/minio-go/v7`）
- 客户端构造点
- 配置字段名

- [ ] **Step 2: 写 `mall-common/minioutil/minioutil.go`**

```go
package minioutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
	UseSSL     bool
}

type Client struct {
	raw    *minio.Client
	bucket string
	host   string
}

func New(cfg Config) (*Client, error) {
	raw, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	host := cfg.PublicHost
	if host == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		host = fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)
	}
	host = strings.TrimRight(host, "/")
	return &Client{raw: raw, bucket: cfg.Bucket, host: host}, nil
}

func (c *Client) PutFile(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := c.raw.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return c.PublicURL(key), nil
}

func (c *Client) PutBytes(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return c.PutFile(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (c *Client) Exists(ctx context.Context, key string) bool {
	_, err := c.raw.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

func (c *Client) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.host, c.bucket, key)
}
```

- [ ] **Step 3: 写小型 unit test 验证 PublicURL 拼接**

`mall-common/minioutil/minioutil_test.go`:

```go
package minioutil

import "testing"

func TestPublicURL(t *testing.T) {
	c := &Client{bucket: "mall-media", host: "http://localhost:9000"}
	got := c.PublicURL("products/seed/1.jpg")
	want := "http://localhost:9000/mall-media/products/seed/1.jpg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestPublicURLTrimsSlash(t *testing.T) {
	c := &Client{bucket: "mall-media", host: "http://localhost:9000/"}
	// 需要 New 处理 trim；这里直接构造避开 New
	c.host = "http://localhost:9000"
	got := c.PublicURL("a/b.jpg")
	want := "http://localhost:9000/mall-media/a/b.jpg"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
```

```bash
cd /home/carter/workspace/go/yw-mall/mall-common && go test ./minioutil/...
```

Expected: PASS.

- [ ] **Step 4: review-rpc 接入 minioutil**

读 `mall-review-rpc/internal/svc/servicecontext.go`，把现有 `*minio.Client` 字段替换为 `*minioutil.Client`，初始化改用 `minioutil.New(...)`。

读 `mall-review-rpc/internal/logic/uploadreviewmedialogic.go`，把 `svcCtx.MinioClient.PutObject(...)` 改为：

```go
url, err := l.svcCtx.Minio.PutFile(l.ctx, key, file, header.Size, header.Header.Get("Content-Type"))
```

go.mod replace（review-rpc 已经依赖 mall-common）：

```bash
cd /home/carter/workspace/go/yw-mall/mall-review-rpc && go mod tidy && go build ./...
```

Expected: 干净 build。

- [ ] **Step 5: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-common/minioutil/ mall-review-rpc/internal/svc/servicecontext.go \
        mall-review-rpc/internal/logic/uploadreviewmedialogic.go \
        mall-review-rpc/go.mod mall-review-rpc/go.sum mall-common/go.mod mall-common/go.sum
git commit -m "refactor(common): extract minioutil; review-rpc consumes it"
```

---

## Task 3: mall-shop-rpc proto + DDL

**Files:**
- Create: `mall-common/proto/shop/shop.proto`
- Create: `mall-shop-rpc/sql/shop.sql`

- [ ] **Step 1: 写 proto**

参考 spec §2.2，完整内容写入 `mall-common/proto/shop/shop.proto`。

- [ ] **Step 2: 写 DDL**

参考 spec §2.1，完整内容写入 `mall-shop-rpc/sql/shop.sql`（含 `shop` 和 `shop_follow` 两张表）。

- [ ] **Step 3: protoc 生成**

```bash
export PATH="/home/carter/workspace/go/bin:$PATH"
cd /home/carter/workspace/go/yw-mall/mall-common
protoc --go_out=. --go-grpc_out=. proto/shop/shop.proto
```

Expected: 生成 `proto/shop/shop.pb.go` 和 `proto/shop/shop_grpc.pb.go`。

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-common/proto/shop/ mall-shop-rpc/sql/
git commit -m "feat(shop): add proto definition and DDL"
```

---

## Task 4: mall-shop-rpc 服务脚手架

**Files:**
- Create: `mall-shop-rpc/`（goctl 整套生成）

- [ ] **Step 1: 在 mall-shop-rpc 写 shop.proto 入口（`option go_package` 适配 go-zero）**

```bash
mkdir -p /home/carter/workspace/go/yw-mall/mall-shop-rpc
cd /home/carter/workspace/go/yw-mall/mall-shop-rpc
cp ../mall-common/proto/shop/shop.proto .
```

修改 `shop.proto` 顶部 `option go_package = "./shop";` 保留即可（go-zero 期望此格式）。

- [ ] **Step 2: goctl 生成 RPC 服务**

```bash
export PATH="/home/carter/workspace/go/bin:$PATH"
goctl rpc protoc shop.proto --go_out=. --go-grpc_out=. --zrpc_out=. --style=gozero
```

Expected: 生成 `etc/shop.yaml`、`internal/{config,logic,server,svc}/`、`shop.go`、`shopclient/shop.go`、`shop/shop.pb.go`、`shop/shop_grpc.pb.go`。

- [ ] **Step 3: goctl 生成 model**

```bash
goctl model mysql ddl -src "sql/shop.sql" -dir internal/model -c --style=gozero
```

> sql 目录还没有；上一 task 已创建于上层 yw-mall/mall-shop-rpc/sql。如不存在，复制：
```bash
mkdir -p sql && cp ../mall-shop-rpc/sql/shop.sql sql/ 2>/dev/null || true
```

如果脚本已经在正确目录则不需要复制。

- [ ] **Step 4: 配置 etc/shop.yaml**

替换为：

```yaml
Name: shop.rpc
ListenOn: 0.0.0.0:9017
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: shop.rpc

DataSource: proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_shop?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai

Cache:
  - Host: 127.0.0.1:6379
```

- [ ] **Step 5: 改 go.mod 用 ../mall-common replace**

```bash
go mod init mall-shop-rpc
go mod edit -require=mall-common@v0.0.0 -replace=mall-common=../mall-common
go mod tidy
```

把 `internal/svc/servicecontext.go` 中可能的 `mall-shop-rpc/...` 内部包引用调整一致。

- [ ] **Step 6: 让脚手架编译过**

```bash
go build ./...
```

logic 文件目前都是空 stub（只 return nil, nil），允许编译通过。

- [ ] **Step 7: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-shop-rpc/
git commit -m "feat(shop-rpc): scaffold service with goctl (proto + model + stubs)"
```

---

## Task 5: mall-shop-rpc ServiceContext

**Files:**
- Modify: `mall-shop-rpc/internal/config/config.go`
- Modify: `mall-shop-rpc/internal/svc/servicecontext.go`
- Modify: `mall-shop-rpc/etc/shop.yaml`

- [ ] **Step 1: Config**

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

- [ ] **Step 2: ServiceContext**

```go
package svc

import (
	"mall-shop-rpc/internal/config"
	"mall-shop-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config          config.Config
	DB              sqlx.SqlConn
	ShopModel       model.ShopModel
	ShopFollowModel model.ShopFollowModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:          c,
		DB:              conn,
		ShopModel:       model.NewShopModel(conn, c.Cache),
		ShopFollowModel: model.NewShopFollowModel(conn, c.Cache),
	}
}
```

- [ ] **Step 3: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-shop-rpc && go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-shop-rpc/internal/config mall-shop-rpc/internal/svc mall-shop-rpc/etc
git commit -m "feat(shop-rpc): wire ServiceContext (db + cache)"
```

---

## Task 6: mall-shop-rpc Get/List/Recommended/Update/Incr

**Files:**
- Modify: `mall-shop-rpc/internal/logic/getshoplogic.go`
- Modify: `mall-shop-rpc/internal/logic/listshopslogic.go`
- Modify: `mall-shop-rpc/internal/logic/listrecommendedshopslogic.go`
- Modify: `mall-shop-rpc/internal/logic/updateshoplogic.go`
- Modify: `mall-shop-rpc/internal/logic/incrproductcountlogic.go`
- Modify: `mall-shop-rpc/internal/logic/createshoplogic.go`

- [ ] **Step 1: GetShop**

```go
func (l *GetShopLogic) GetShop(in *shop.GetShopReq) (*shop.GetShopResp, error) {
	s, err := l.svcCtx.ShopModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.ShopNotFound)
		}
		return nil, err
	}
	return &shop.GetShopResp{Shop: toShopProto(s)}, nil
}
```

加 helper（放进 `internal/logic/helpers.go`，新建文件）：

```go
package logic

import (
	"mall-shop-rpc/internal/model"
	"mall-shop-rpc/shop"
)

func toShopProto(s *model.Shop) *shop.Shop {
	rating, _ := s.Rating.Float64()
	return &shop.Shop{
		Id:           int64(s.Id),
		Name:         s.Name,
		Logo:         s.Logo,
		Banner:       s.Banner,
		Description:  s.Description,
		Rating:       rating,
		ProductCount: int32(s.ProductCount),
		FollowCount:  int32(s.FollowCount),
		Status:       int32(s.Status),
		CreateTime:   s.CreateTime.Unix(),
	}
}
```

> goctl 生成的 model 类型字段名 / 类型可能略不同（`Rating sql.NullFloat64` 或 decimal.Decimal）；按实际生成结果调整。

- [ ] **Step 2: ListShops**

```go
func (l *ListShopsLogic) ListShops(in *shop.ListShopsReq) (*shop.ListShopsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size
	rows, err := l.queryList(offset, size)
	if err != nil {
		return nil, err
	}
	total, _ := l.countAll()
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: total}, nil
}

func (l *ListShopsLogic) queryList(offset, size int32) ([]*model.Shop, error) {
	var rows []*model.Shop
	q := "SELECT * FROM shop WHERE status=1 ORDER BY id DESC LIMIT ?, ?"
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, offset, size)
	return rows, err
}

func (l *ListShopsLogic) countAll() (int64, error) {
	var n int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &n, "SELECT COUNT(*) FROM shop WHERE status=1")
	return n, err
}
```

helper：

```go
func normPage(p, s int32) (int32, int32) {
	if p <= 0 {
		p = 1
	}
	if s <= 0 || s > 50 {
		s = 20
	}
	return p, s
}
```

- [ ] **Step 3: ListRecommendedShops**

```go
func (l *ListRecommendedShopsLogic) ListRecommendedShops(in *shop.ListRecommendedShopsReq) (*shop.ListShopsResp, error) {
	limit := in.Limit
	if limit <= 0 || limit > 20 {
		limit = 8
	}
	var rows []*model.Shop
	q := "SELECT * FROM shop WHERE status=1 ORDER BY rating DESC, follow_count DESC LIMIT ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, limit); err != nil {
		return nil, err
	}
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: int64(len(rows))}, nil
}
```

- [ ] **Step 4: UpdateShop / IncrProductCount / CreateShop**

```go
// UpdateShop
func (l *UpdateShopLogic) UpdateShop(in *shop.UpdateShopReq) (*shop.OkResp, error) {
	s, err := l.svcCtx.ShopModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.ShopNotFound)
		}
		return nil, err
	}
	if in.Name != "" { s.Name = in.Name }
	if in.Logo != "" { s.Logo = in.Logo }
	if in.Banner != "" { s.Banner = in.Banner }
	if in.Description != "" { s.Description = in.Description }
	if err := l.svcCtx.ShopModel.Update(l.ctx, s); err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}

// IncrProductCount
func (l *IncrProductCountLogic) IncrProductCount(in *shop.IncrProductCountReq) (*shop.OkResp, error) {
	if in.ShopId == 0 || in.Delta == 0 {
		return &shop.OkResp{Ok: true}, nil
	}
	_, err := l.svcCtx.DB.ExecCtx(l.ctx, "UPDATE shop SET product_count = product_count + ? WHERE id = ?", in.Delta, in.ShopId)
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}

// CreateShop
func (l *CreateShopLogic) CreateShop(in *shop.CreateShopReq) (*shop.CreateShopResp, error) {
	now := time.Now()
	rating := decimal.NewFromFloat(in.Rating)
	row := &model.Shop{
		Name:        in.Name,
		Logo:        in.Logo,
		Banner:      in.Banner,
		Description: in.Description,
		Rating:      rating,
		Status:      1,
		CreateTime:  now,
		UpdateTime:  now,
	}
	r, err := l.svcCtx.ShopModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	id, _ := r.LastInsertId()
	return &shop.CreateShopResp{Id: id}, nil
}
```

> 实际 model 字段类型按 goctl 生成调整（decimal.Decimal vs float64）。

- [ ] **Step 5: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-shop-rpc && go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-shop-rpc/internal/logic/
git commit -m "feat(shop-rpc): GetShop / ListShops / ListRecommendedShops / Update / IncrProductCount / CreateShop"
```

---

## Task 7: mall-shop-rpc Follow/Unfollow/IsFollowing/ListFollowed

**Files:**
- Modify: `mall-shop-rpc/internal/logic/followshoplogic.go`
- Modify: `mall-shop-rpc/internal/logic/unfollowshoplogic.go`
- Modify: `mall-shop-rpc/internal/logic/isfollowinglogic.go`
- Modify: `mall-shop-rpc/internal/logic/listfollowedshopslogic.go`

- [ ] **Step 1: FollowShop**

```go
func (l *FollowShopLogic) FollowShop(in *shop.FollowShopReq) (*shop.OkResp, error) {
	if in.UserId == 0 || in.ShopId == 0 {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		res, e := sess.ExecCtx(ctx,
			"INSERT IGNORE INTO shop_follow(user_id, shop_id, create_time) VALUES (?, ?, ?)",
			in.UserId, in.ShopId, time.Now().Unix())
		if e != nil {
			return e
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return errorx.NewCodeError(errorx.ShopFollowAlreadyExists)
		}
		_, e = sess.ExecCtx(ctx, "UPDATE shop SET follow_count = follow_count + 1 WHERE id = ?", in.ShopId)
		return e
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
```

- [ ] **Step 2: UnfollowShop**

```go
func (l *UnfollowShopLogic) UnfollowShop(in *shop.UnfollowShopReq) (*shop.OkResp, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		res, e := sess.ExecCtx(ctx, "DELETE FROM shop_follow WHERE user_id=? AND shop_id=?", in.UserId, in.ShopId)
		if e != nil {
			return e
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return nil // 不存在直接返回成功（幂等）
		}
		_, e = sess.ExecCtx(ctx, "UPDATE shop SET follow_count = GREATEST(follow_count - 1, 0) WHERE id = ?", in.ShopId)
		return e
	})
	if err != nil {
		return nil, err
	}
	return &shop.OkResp{Ok: true}, nil
}
```

- [ ] **Step 3: IsFollowing**

```go
func (l *IsFollowingLogic) IsFollowing(in *shop.IsFollowingReq) (*shop.IsFollowingResp, error) {
	var n int64
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &n, "SELECT COUNT(*) FROM shop_follow WHERE user_id=? AND shop_id=?", in.UserId, in.ShopId)
	if err != nil {
		return nil, err
	}
	return &shop.IsFollowingResp{IsFollowing: n > 0}, nil
}
```

- [ ] **Step 4: ListFollowedShops**

```go
func (l *ListFollowedShopsLogic) ListFollowedShops(in *shop.ListFollowedShopsReq) (*shop.ListShopsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size
	var rows []*model.Shop
	q := `SELECT s.* FROM shop s
	      INNER JOIN shop_follow f ON s.id = f.shop_id
	      WHERE f.user_id = ? AND s.status = 1
	      ORDER BY f.create_time DESC LIMIT ?, ?`
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, in.UserId, offset, size); err != nil {
		return nil, err
	}
	out := make([]*shop.Shop, 0, len(rows))
	for _, s := range rows {
		out = append(out, toShopProto(s))
	}
	return &shop.ListShopsResp{Shops: out, Total: int64(len(rows))}, nil
}
```

- [ ] **Step 5: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-shop-rpc && go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-shop-rpc/internal/logic/
git commit -m "feat(shop-rpc): FollowShop / UnfollowShop / IsFollowing / ListFollowedShops"
```

---

## Task 8: user-rpc 地址 proto + DDL

**Files:**
- Modify: `mall-common/proto/user/user.proto`
- Modify: `mall-user-rpc/user.proto`
- Modify: `mall-user-rpc/sql/user.sql`

- [ ] **Step 1: proto 追加**

参考 spec §3.2，把 7 个 RPC 和 `Address` / 各 req/resp 消息追加到两份 proto（mall-common 和 mall-user-rpc 各一份，保持同步）。

- [ ] **Step 2: SQL 追加**

`mall-user-rpc/sql/user.sql` 末尾追加 `user_address` 表（spec §3.1）。

- [ ] **Step 3: protoc 重新生成**

```bash
export PATH="/home/carter/workspace/go/bin:$PATH"
cd /home/carter/workspace/go/yw-mall/mall-common
protoc --go_out=. --go-grpc_out=. proto/user/user.proto

cd /home/carter/workspace/go/yw-mall/mall-user-rpc
goctl rpc protoc user.proto --go_out=. --go-grpc_out=. --zrpc_out=. --style=gozero
```

注意：goctl 会生成 7 个新的 logic stub 文件（addaddresslogic.go 等）；shipped/已有的 logic 文件不会被覆盖。

- [ ] **Step 4: model 重生成（覆盖加 user_address）**

```bash
goctl model mysql ddl -src "sql/user.sql" -dir internal/model -c --style=gozero
```

- [ ] **Step 5: 编译 + Commit（允许 logic stub 暂未实现）**

```bash
go build ./...
```

新 logic 是空 stub，应该编译通过。

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-common/proto/user/ mall-user-rpc/user.proto mall-user-rpc/sql/user.sql \
        mall-user-rpc/user/ mall-user-rpc/userclient/ mall-user-rpc/internal/logic/ \
        mall-user-rpc/internal/server/ mall-user-rpc/internal/model/
git commit -m "feat(user): add address schema, proto, model, scaffold"
```

---

## Task 9: user-rpc Add/Update/Delete address logic

**Files:**
- Modify: `mall-user-rpc/internal/logic/addaddresslogic.go`
- Modify: `mall-user-rpc/internal/logic/updateaddresslogic.go`
- Modify: `mall-user-rpc/internal/logic/deleteaddresslogic.go`
- Create: `mall-user-rpc/internal/logic/address_helpers.go`

- [ ] **Step 1: helpers**

`mall-user-rpc/internal/logic/address_helpers.go`:

```go
package logic

import (
	"mall-user-rpc/internal/model"
	"mall-user-rpc/user"
)

func toAddressProto(a *model.UserAddress) *user.Address {
	return &user.Address{
		Id:           int64(a.Id),
		UserId:       int64(a.UserId),
		ReceiverName: a.ReceiverName,
		Phone:        a.Phone,
		Province:     a.Province,
		City:         a.City,
		District:     a.District,
		Detail:       a.Detail,
		IsDefault:    a.IsDefault == 1,
		CreateTime:   a.CreateTime.Unix(),
	}
}
```

- [ ] **Step 2: AddAddress**（4 步事务）

```go
func (l *AddAddressLogic) AddAddress(in *user.AddAddressReq) (*user.AddAddressResp, error) {
	if in.UserId == 0 {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	var newId int64
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		// 1. 上限校验
		var count int64
		if err := sess.QueryRowCtx(ctx, &count,
			"SELECT COUNT(*) FROM user_address WHERE user_id = ?", in.UserId); err != nil {
			return err
		}
		if count >= 20 {
			return errorx.NewCodeError(errorx.UserAddressLimit)
		}
		// 2. 决定 effective_default
		effectiveDefault := in.IsDefault || count == 0
		// 3. 若 default，先清旧
		if effectiveDefault {
			if _, err := sess.ExecCtx(ctx,
				"UPDATE user_address SET is_default=0 WHERE user_id=?", in.UserId); err != nil {
				return err
			}
		}
		// 4. INSERT
		now := time.Now().Unix()
		var defaultFlag int8
		if effectiveDefault {
			defaultFlag = 1
		}
		res, err := sess.ExecCtx(ctx,
			`INSERT INTO user_address(user_id, receiver_name, phone, province, city, district, detail, is_default, create_time, update_time)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			in.UserId, in.ReceiverName, in.Phone, in.Province, in.City, in.District, in.Detail,
			defaultFlag, now, now)
		if err != nil {
			return err
		}
		newId, _ = res.LastInsertId()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user.AddAddressResp{Id: newId}, nil
}
```

- [ ] **Step 3: UpdateAddress**

```go
func (l *UpdateAddressLogic) UpdateAddress(in *user.UpdateAddressReq) (*user.OkResp, error) {
	res, err := l.svcCtx.DB.ExecCtx(l.ctx,
		`UPDATE user_address
		 SET receiver_name=?, phone=?, province=?, city=?, district=?, detail=?, update_time=?
		 WHERE id=? AND user_id=?`,
		in.ReceiverName, in.Phone, in.Province, in.City, in.District, in.Detail, time.Now().Unix(),
		in.Id, in.UserId)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// 区分 not found vs forbidden
		var ownerId int64
		_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &ownerId, "SELECT user_id FROM user_address WHERE id=?", in.Id)
		if ownerId == 0 {
			return nil, errorx.NewCodeError(errorx.UserAddressNotFound)
		}
		return nil, errorx.NewCodeError(errorx.UserAddressForbidden)
	}
	return &user.OkResp{Ok: true}, nil
}
```

- [ ] **Step 4: DeleteAddress**（含默认地址自动转移）

```go
func (l *DeleteAddressLogic) DeleteAddress(in *user.DeleteAddressReq) (*user.OkResp, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		// 1. 校验所有权 + 取 is_default
		var owner int64
		var isDefault int8
		row := sess.QueryRowCtx
		if err := row(ctx, &struct {
			UserId    int64 `db:"user_id"`
			IsDefault int8  `db:"is_default"`
		}{owner, isDefault}, "SELECT user_id, is_default FROM user_address WHERE id=?", in.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return errorx.NewCodeError(errorx.UserAddressNotFound)
			}
			return err
		}
		if owner != in.UserId {
			return errorx.NewCodeError(errorx.UserAddressForbidden)
		}
		// 2. 删除
		if _, err := sess.ExecCtx(ctx, "DELETE FROM user_address WHERE id=?", in.Id); err != nil {
			return err
		}
		// 3. 若是默认地址，自动转移
		if isDefault == 1 {
			if _, err := sess.ExecCtx(ctx,
				`UPDATE user_address SET is_default=1
				 WHERE user_id=?
				 ORDER BY update_time DESC LIMIT 1`, in.UserId); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}
```

> 注意：上面的 row scan 写法可能有问题；在实施时先用结构体 dst 然后 GetCtx 或两次 QueryRow 也可以。如果 row helper 不便用就改成两次 SELECT。

- [ ] **Step 5: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-user-rpc && go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-user-rpc/internal/logic/
git commit -m "feat(user-rpc): AddAddress / UpdateAddress / DeleteAddress with tx + ownership check"
```

---

## Task 10: user-rpc SetDefault/List/Get/GetDefault

**Files:**
- Modify: `mall-user-rpc/internal/logic/setdefaultaddresslogic.go`
- Modify: `mall-user-rpc/internal/logic/listaddresseslogic.go`
- Modify: `mall-user-rpc/internal/logic/getaddresslogic.go`
- Modify: `mall-user-rpc/internal/logic/getdefaultaddresslogic.go`

- [ ] **Step 1: SetDefaultAddress**（二步事务）

```go
func (l *SetDefaultAddressLogic) SetDefaultAddress(in *user.SetDefaultAddressReq) (*user.OkResp, error) {
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, sess sqlx.Session) error {
		// 1. 验所有权
		var owner int64
		if err := sess.QueryRowCtx(ctx, &owner, "SELECT user_id FROM user_address WHERE id=?", in.Id); err != nil {
			if err == sqlx.ErrNotFound {
				return errorx.NewCodeError(errorx.UserAddressNotFound)
			}
			return err
		}
		if owner != in.UserId {
			return errorx.NewCodeError(errorx.UserAddressForbidden)
		}
		// 2. 清旧 + 设新
		if _, err := sess.ExecCtx(ctx, "UPDATE user_address SET is_default=0 WHERE user_id=?", in.UserId); err != nil {
			return err
		}
		_, err := sess.ExecCtx(ctx, "UPDATE user_address SET is_default=1 WHERE id=?", in.Id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &user.OkResp{Ok: true}, nil
}
```

- [ ] **Step 2: ListAddresses**

```go
func (l *ListAddressesLogic) ListAddresses(in *user.ListAddressesReq) (*user.ListAddressesResp, error) {
	var rows []*model.UserAddress
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		`SELECT * FROM user_address WHERE user_id=? ORDER BY is_default DESC, update_time DESC`,
		in.UserId); err != nil {
		return nil, err
	}
	out := make([]*user.Address, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAddressProto(a))
	}
	return &user.ListAddressesResp{Addresses: out}, nil
}
```

- [ ] **Step 3: GetAddress**（强制 user_id 校验）

```go
func (l *GetAddressLogic) GetAddress(in *user.GetAddressReq) (*user.Address, error) {
	var row model.UserAddress
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row, "SELECT * FROM user_address WHERE id=?", in.Id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.UserAddressNotFound)
		}
		return nil, err
	}
	if int64(row.UserId) != in.UserId {
		return nil, errorx.NewCodeError(errorx.UserAddressForbidden)
	}
	return toAddressProto(&row), nil
}
```

- [ ] **Step 4: GetDefaultAddress**（无默认返回 id=0）

```go
func (l *GetDefaultAddressLogic) GetDefaultAddress(in *user.GetDefaultAddressReq) (*user.Address, error) {
	var row model.UserAddress
	err := l.svcCtx.DB.QueryRowCtx(l.ctx, &row,
		"SELECT * FROM user_address WHERE user_id=? AND is_default=1 LIMIT 1", in.UserId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &user.Address{Id: 0, UserId: in.UserId}, nil
		}
		return nil, err
	}
	return toAddressProto(&row), nil
}
```

- [ ] **Step 5: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-user-rpc && go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-user-rpc/internal/logic/
git commit -m "feat(user-rpc): SetDefault / List / Get / GetDefault address"
```

---

## Task 11: product-rpc 加 shop_id + ListShopProducts + ShopRpc 依赖

**Files:**
- Modify: `mall-common/proto/product/product.proto`
- Modify: `mall-product-rpc/product.proto`
- Modify: `mall-product-rpc/sql/product.sql`
- Modify: `mall-product-rpc/internal/config/config.go`
- Modify: `mall-product-rpc/internal/svc/servicecontext.go`
- Modify: `mall-product-rpc/internal/logic/createproductlogic.go`
- Modify: `mall-product-rpc/internal/logic/listshopproductslogic.go`（新生成）
- Modify: `mall-product-rpc/etc/product.yaml`

- [ ] **Step 1: proto + DDL**

按 spec §4.1 修改 proto（加 `shop_id` 到 `Product`、加 `ListShopProducts` RPC + req）和 SQL（ALTER TABLE）。

- [ ] **Step 2: 重生成 + 应用 ALTER**

```bash
export PATH="/home/carter/workspace/go/bin:$PATH"
cd /home/carter/workspace/go/yw-mall/mall-common
protoc --go_out=. --go-grpc_out=. proto/product/product.proto
cd /home/carter/workspace/go/yw-mall/mall-product-rpc
goctl rpc protoc product.proto --go_out=. --go-grpc_out=. --zrpc_out=. --style=gozero
goctl model mysql ddl -src "sql/product.sql" -dir internal/model -c --style=gozero
```

> ALTER 语句需要在 `start.sh nuke && start.sh start` 时通过 `mall_product` schema 重建得以应用——sql 文件加 ALTER 可能 nuke 重建时报错（because table 已经从 CREATE 创建好了带新列）。**改法**：直接把 `shop_id` 列加到 CREATE TABLE 而不是写 ALTER，新建库时一步到位。

- [ ] **Step 3: Config + svc 加 ShopRpc**

```go
// config.go
type Config struct {
	zrpc.RpcServerConf
	DataSource string
	Cache      cache.CacheConf
	ShopRpc    zrpc.RpcClientConf
}
```

```go
// servicecontext.go
import shopclient "mall-shop-rpc/shopclient"

type ServiceContext struct {
	// ... 已有
	ShopRpc shopclient.Shop
}

func NewServiceContext(c config.Config) *ServiceContext {
	// ... 已有
	return &ServiceContext{
		// ... 已有
		ShopRpc: shopclient.NewShop(zrpc.MustNewClient(c.ShopRpc)),
	}
}
```

- [ ] **Step 4: yaml**

```yaml
ShopRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: shop.rpc
```

- [ ] **Step 5: go.mod**

```bash
go mod edit -require=mall-shop-rpc@v0.0.0 -replace=mall-shop-rpc=../mall-shop-rpc
go mod tidy
```

- [ ] **Step 6: CreateProduct 调 IncrProductCount**

```go
func (l *CreateProductLogic) CreateProduct(in *product.CreateProductReq) (*product.CreateProductResp, error) {
	// ... 已有 INSERT
	id := /* ... */
	if in.ShopId > 0 {
		_, _ = l.svcCtx.ShopRpc.IncrProductCount(l.ctx, &shop.IncrProductCountReq{
			ShopId: in.ShopId, Delta: 1,
		})
		// 失败不阻塞主流程；shop 计数最终一致
	}
	return &product.CreateProductResp{Id: id}, nil
}
```

- [ ] **Step 7: ListShopProducts logic**

```go
func (l *ListShopProductsLogic) ListShopProducts(in *product.ListShopProductsReq) (*product.ListProductsResp, error) {
	page, size := normPage(in.Page, in.PageSize)
	offset := (page - 1) * size
	var rows []*model.Product
	q := "SELECT * FROM product WHERE shop_id=? AND status=1 ORDER BY id DESC LIMIT ?, ?"
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, q, in.ShopId, offset, size); err != nil {
		return nil, err
	}
	var total int64
	_ = l.svcCtx.DB.QueryRowCtx(l.ctx, &total, "SELECT COUNT(*) FROM product WHERE shop_id=? AND status=1", in.ShopId)
	out := make([]*product.Product, 0, len(rows))
	for _, p := range rows {
		out = append(out, toProductProto(p))
	}
	return &product.ListProductsResp{Products: out, Total: total}, nil
}
```

> `toProductProto` 现有；新加的 `ShopId` 字段映射别忘了。

- [ ] **Step 8: 编译 + Commit**

```bash
go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-common/proto/product/ mall-product-rpc/
git commit -m "feat(product-rpc): add shop_id, ListShopProducts, ShopRpc dependency"
```

---

## Task 12: order-rpc 收货地址快照 + UserRpc 依赖

**Files:**
- Modify: `mall-common/proto/order/order.proto`
- Modify: `mall-order-rpc/order.proto`
- Modify: `mall-order-rpc/sql/order.sql`
- Modify: `mall-order-rpc/internal/config/config.go`
- Modify: `mall-order-rpc/internal/svc/servicecontext.go`
- Modify: `mall-order-rpc/internal/logic/createorderlogic.go`
- Modify: `mall-order-rpc/internal/logic/getorderlogic.go`
- Modify: `mall-order-rpc/etc/order.yaml`

- [ ] **Step 1: proto + DDL**

按 spec §4.2 修改：CreateOrderReq 加 `address_id`，GetOrderResp 加 7 个 receiver 字段。`order` 表 CREATE TABLE 直接加 7 列（不用 ALTER；同 product 思路）。

- [ ] **Step 2: 重生成**

```bash
export PATH="/home/carter/workspace/go/bin:$PATH"
cd /home/carter/workspace/go/yw-mall/mall-common
protoc --go_out=. --go-grpc_out=. proto/order/order.proto
cd /home/carter/workspace/go/yw-mall/mall-order-rpc
goctl rpc protoc order.proto --go_out=. --go-grpc_out=. --zrpc_out=. --style=gozero
goctl model mysql ddl -src "sql/order.sql" -dir internal/model -c --style=gozero
```

- [ ] **Step 3: Config + svc 加 UserRpc**

```go
// config.go
UserRpc zrpc.RpcClientConf
```

```go
// servicecontext.go
import userclient "mall-user-rpc/userclient"

UserRpc userclient.User

UserRpc: userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
```

- [ ] **Step 4: yaml**

```yaml
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: user.rpc
```

- [ ] **Step 5: go.mod**

```bash
go mod edit -require=mall-user-rpc@v0.0.0 -replace=mall-user-rpc=../mall-user-rpc
go mod tidy
```

- [ ] **Step 6: CreateOrder 注入快照**

```go
func (l *CreateOrderLogic) CreateOrder(in *order.CreateOrderReq) (*order.CreateOrderResp, error) {
	if in.AddressId == 0 {
		return nil, errorx.NewCodeError(errorx.OrderAddressRequired)
	}
	addr, err := l.svcCtx.UserRpc.GetAddress(l.ctx, &user.GetAddressReq{
		UserId: in.UserId, Id: in.AddressId,
	})
	if err != nil {
		return nil, err
	}
	// ... 现有的 INSERT order：把 7 个 receiver 字段填入
	// addr.ReceiverName / Phone / Province / City / District / Detail
	// addr.Id 写入 order.address_id
}
```

- [ ] **Step 7: GetOrder 透出 7 字段**

`getorderlogic.go` 把现有 row → resp 映射加上 address_id + 6 个 receiver 字段。

- [ ] **Step 8: 编译 + Commit**

```bash
go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-common/proto/order/ mall-order-rpc/
git commit -m "feat(order-rpc): receiver address snapshot + UserRpc dependency"
```

---

## Task 13: mall-api shop 路由 + svc 接入

**Files:**
- Create: `mall-api/mall-shop.api`
- Modify: `mall-api/mall.api`
- Modify: `mall-api/etc/mall-api.yaml`
- Modify: `mall-api/internal/config/config.go`
- Modify: `mall-api/internal/svc/servicecontext.go`
- Modify: `mall-api/go.mod`
- Modify: `mall-api/internal/logic/*.go`（新生成）

- [ ] **Step 1: 写 mall-shop.api**

```api
syntax = "v1"

type (
	ShopDTO {
		Id           int64   `json:"id"`
		Name         string  `json:"name"`
		Logo         string  `json:"logo"`
		Banner       string  `json:"banner"`
		Description  string  `json:"description"`
		Rating       float64 `json:"rating"`
		ProductCount int32   `json:"productCount"`
		FollowCount  int32   `json:"followCount"`
		Status       int32   `json:"status"`
		CreateTime   int64   `json:"createTime"`
	}
	ShopDetailReq { Id int64 `path:"id"` }
	ShopDetailResp { Shop ShopDTO `json:"shop"` }
	ShopListReq { Page int32 `form:"page,optional"` PageSize int32 `form:"pageSize,optional"` }
	ShopListResp { Shops []ShopDTO `json:"shops"` Total int64 `json:"total"` }
	ShopProductsReq { Id int64 `path:"id"` Page int32 `form:"page,optional"` PageSize int32 `form:"pageSize,optional"` }

	FollowShopReq { Id int64 `path:"id"` }
	IsFollowingShopReq { Id int64 `path:"id"` }
	IsFollowingShopResp { IsFollowing bool `json:"isFollowing"` }
	PageReq { Page int32 `form:"page,optional"` PageSize int32 `form:"pageSize,optional"` }
)

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

@server (prefix: /api, jwt: Auth)
service mall-api {
	@handler FollowShop
	post /shop/:id/follow (FollowShopReq) returns (OkResp)
	@handler UnfollowShop
	post /shop/:id/unfollow (FollowShopReq) returns (OkResp)
	@handler IsFollowingShop
	get /shop/:id/is-following (IsFollowingShopReq) returns (IsFollowingShopResp)
	@handler MyFollowedShops
	get /shop/my-followed (PageReq) returns (ShopListResp)
}
```

- [ ] **Step 2: import 到 mall.api**

加一行 `import "mall-shop.api"`。

- [ ] **Step 3: yaml + config + svc + go.mod**

`mall-api/etc/mall-api.yaml` 追加：

```yaml
ShopRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: shop.rpc
```

`config.go` 加 `ShopRpc zrpc.RpcClientConf`。

`servicecontext.go` 加：

```go
import shopclient "mall-shop-rpc/shopclient"

ShopRpc shopclient.Shop

ShopRpc: shopclient.NewShop(zrpc.MustNewClient(c.ShopRpc)),
```

```bash
cd /home/carter/workspace/go/yw-mall/mall-api
go mod edit -require=mall-shop-rpc@v0.0.0 -replace=mall-shop-rpc=../mall-shop-rpc
go mod tidy
```

- [ ] **Step 4: goctl 生成 handler/logic**

```bash
goctl api go -api mall.api -dir . --style gozero
```

- [ ] **Step 5: 实现 8 个 logic（pattern 一致：取 jwt user_id → 调 ShopRpc → map proto→DTO）**

shopdetaillogic / shoplistlogic / shoprecommendedlogic / shopproductslogic / followshoplogic / unfollowshoplogic / isfollowingshoplogic / myfollowedshopslogic

helper：

```go
// mall-api/internal/logic/shop_helpers.go
func protoShopToDTO(s *shoppb.Shop) types.ShopDTO {
	return types.ShopDTO{
		Id: s.Id, Name: s.Name, Logo: s.Logo, Banner: s.Banner, Description: s.Description,
		Rating: s.Rating, ProductCount: s.ProductCount, FollowCount: s.FollowCount,
		Status: s.Status, CreateTime: s.CreateTime,
	}
}
```

每个 logic 用现成的 ProductRpc/ShopRpc 调用。FollowShop 取 jwt user_id 用 `l.ctx.Value("userId")` 模式（按现有 mall-api 其他 jwt handler 一致）。

- [ ] **Step 6: 编译 + Commit**

```bash
go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-api/mall-shop.api mall-api/mall.api \
        mall-api/etc/mall-api.yaml mall-api/internal/config mall-api/internal/svc \
        mall-api/go.mod mall-api/go.sum \
        mall-api/internal/handler mall-api/internal/types mall-api/internal/logic
git commit -m "feat(api): expose shop routes (browse + follow + my-followed)"
```

---

## Task 14: mall-api 地址路由

**Files:**
- Create: `mall-api/mall-address.api`
- Modify: `mall-api/mall.api`
- Modify: `mall-api/internal/logic/*.go`（5 个新生成）

- [ ] **Step 1: api**

```api
syntax = "v1"

type (
	AddressDTO {
		Id           int64  `json:"id"`
		ReceiverName string `json:"receiverName"`
		Phone        string `json:"phone"`
		Province     string `json:"province"`
		City         string `json:"city"`
		District     string `json:"district"`
		Detail       string `json:"detail"`
		IsDefault    bool   `json:"isDefault"`
		CreateTime   int64  `json:"createTime"`
	}
	AddAddressReq {
		ReceiverName string `json:"receiverName"`
		Phone        string `json:"phone"`
		Province     string `json:"province"`
		City         string `json:"city"`
		District     string `json:"district"`
		Detail       string `json:"detail"`
		IsDefault    bool   `json:"isDefault,optional"`
	}
	AddAddressResp { Id int64 `json:"id"` }
	UpdateAddressReq {
		Id           int64  `json:"id"`
		ReceiverName string `json:"receiverName"`
		Phone        string `json:"phone"`
		Province     string `json:"province"`
		City         string `json:"city"`
		District     string `json:"district"`
		Detail       string `json:"detail"`
	}
	DeleteAddressReq { Id int64 `json:"id"` }
	SetDefaultAddressReq { Id int64 `json:"id"` }
	AddressListResp { Addresses []AddressDTO `json:"addresses"` }
)

@server (prefix: /api, jwt: Auth)
service mall-api {
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

- [ ] **Step 2: import + 重生成**

`mall.api` 加 `import "mall-address.api"`，然后 goctl 重生成。

- [ ] **Step 3: 实现 5 个 logic**

每个 logic 取 jwt user_id，调 UserRpc 对应方法，map 结果。

helper：

```go
func protoAddressToDTO(a *userpb.Address) types.AddressDTO {
	return types.AddressDTO{
		Id: a.Id, ReceiverName: a.ReceiverName, Phone: a.Phone,
		Province: a.Province, City: a.City, District: a.District, Detail: a.Detail,
		IsDefault: a.IsDefault, CreateTime: a.CreateTime,
	}
}
```

- [ ] **Step 4: 编译 + Commit**

```bash
go build ./...
cd /home/carter/workspace/go/yw-mall
git add mall-api/mall-address.api mall-api/mall.api \
        mall-api/internal/handler mall-api/internal/types mall-api/internal/logic
git commit -m "feat(api): expose address routes"
```

---

## Task 15: shop seed (cmd/seed/main.go)

**Files:**
- Create: `mall-shop-rpc/cmd/seed/main.go`

- [ ] **Step 1: 写 seed 程序骨架**

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mall-common/minioutil"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type seedCfg struct {
	DataSource string
	Minio      minioutil.Config
}

var shops = []struct {
	Slug, Name, Description string
}{
	{"tech",     "数码电器", "正品数码 / 智能配件 / 影音娱乐"},
	{"fashion",  "潮流服饰", "时尚服装 / 鞋履箱包 / 配饰"},
	{"home",     "家居生活", "床品布艺 / 收纳整理 / 厨卫用品"},
	{"beauty",   "美妆个护", "护肤彩妆 / 香水 / 个护"},
	{"food",     "优鲜食品", "时令水果 / 肉禽蛋 / 零食饮料"},
	{"sport",    "运动户外", "运动装备 / 户外露营 / 健身器材"},
}

const cacheDir = "/tmp/mall-seed-cache"

func main() {
	var c seedCfg
	conf.MustLoad("etc/shop.yaml", &c)
	conn := sqlx.NewMysql(c.DataSource)
	mc, err := minioutil.New(c.Minio)
	if err != nil { panic(err) }
	ctx := context.Background()

	var count int64
	_ = conn.QueryRowCtx(ctx, &count, "SELECT COUNT(*) FROM shop")
	if count >= 6 {
		fmt.Println("shop seed: skip (already populated)")
		return
	}

	os.MkdirAll(cacheDir, 0o755)
	for i, s := range shops {
		shopId := i + 1
		logoURL  := fmt.Sprintf("https://api.dicebear.com/7.x/shapes/svg?seed=%s", s.Slug)
		bannerURL := fmt.Sprintf("https://picsum.photos/seed/banner-%s/800/300", s.Slug)
		logoKey   := fmt.Sprintf("shops/seed/%d-logo.svg", shopId)
		bannerKey := fmt.Sprintf("shops/seed/%d-banner.jpg", shopId)
		logoPublic   := fetchAndUpload(ctx, mc, logoURL, logoKey, "image/svg+xml")
		bannerPublic := fetchAndUpload(ctx, mc, bannerURL, bannerKey, "image/jpeg")

		now := time.Now().Unix()
		rating := 4.5 + float64(i%5)*0.1   // 4.5 ~ 4.9
		followCount := 50 + i*47           // 50, 97, 144, ...
		_, err := conn.ExecCtx(ctx,
			`INSERT INTO shop(name, logo, banner, description, rating, product_count, follow_count, status, create_time, update_time)
			 VALUES (?, ?, ?, ?, ?, 0, ?, 1, ?, ?)`,
			s.Name, logoPublic, bannerPublic, s.Description, rating, followCount, now, now)
		if err != nil {
			fmt.Println("insert shop fail:", err)
			os.Exit(1)
		}
		fmt.Printf("✓ shop %d %s\n", shopId, s.Name)
	}
}

func fetchAndUpload(ctx context.Context, mc *minioutil.Client, url, key, contentType string) string {
	cachePath := filepath.Join(cacheDir, filepath.Base(key))
	var data []byte
	if b, err := os.ReadFile(cachePath); err == nil {
		data = b
	} else {
		resp, err := http.Get(url)
		if err != nil { panic(err) }
		defer resp.Body.Close()
		data, _ = io.ReadAll(resp.Body)
		_ = os.WriteFile(cachePath, data, 0o644)
	}
	publicURL, err := mc.PutBytes(ctx, key, data, contentType)
	if err != nil { panic(err) }
	return publicURL
}
```

- [ ] **Step 2: 配置 etc/shop.yaml 加 minio 段**

```yaml
Minio:
  Endpoint: localhost:9000
  AccessKey: minioadmin
  SecretKey: minioadmin
  Bucket: mall-media
  PublicHost: http://localhost:9000
  UseSSL: false
```

- [ ] **Step 3: 编译 + Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-shop-rpc && go build ./cmd/seed
cd /home/carter/workspace/go/yw-mall
git add mall-shop-rpc/cmd/seed/ mall-shop-rpc/etc/shop.yaml
git commit -m "feat(shop-rpc): seed 6 shops with logo/banner via picsum/dicebear + minio"
```

---

## Task 16: product seed (cmd/seed/main.go)

**Files:**
- Create: `mall-product-rpc/cmd/seed/main.go`

- [ ] **Step 1: 准备 40 个商品定义**

```go
var products = []struct {
	ShopId      int64
	Name        string
	Description string
	Price       int64  // 分
	Stock       int64
	CategoryId  int64
}{
	// 数码电器（shop_id=1）
	{1, "AirPods Pro 2 蓝牙降噪耳机", "主动降噪 / H2 芯片 / 自适应通透模式", 149900, 100, 1},
	{1, "iPad mini 7 8.3 英寸", "A17 Pro / Wi-Fi 256GB", 459900, 50, 1},
	{1, "Anker 充电宝 20000mAh", "PD3.0 双向快充 / 多设备同充", 19900, 200, 1},
	{1, "Logitech MX Master 3S 无线鼠标", "8K DPI / 静音点击 / 多设备切换", 79900, 80, 1},
	{1, "Sony WH-1000XM5 头戴耳机", "DSEE Extreme / 降噪旗舰", 269900, 30, 1},
	{1, "Kindle Oasis 7 英寸", "暖光阅读灯 / 物理按键 / 防水", 229900, 40, 1},

	// 潮流服饰（shop_id=2）
	{2, "夏季纯棉短袖 T 恤", "100% 精梳棉 / 多色可选", 5900, 500, 2},
	{2, "高腰阔腿牛仔裤", "九分拖地版 / 显瘦垂坠", 19900, 300, 2},
	{2, "复古针织开衫", "羊毛混纺 / 慵懒小香风", 29900, 150, 2},
	// ... 自行扩展到 40 个
}
```

> 每店 6-7 个，总数 40。命名走真实化，不要 "商品 1"。

- [ ] **Step 2: 主体逻辑（参考 shop seed）**

幂等检查 `COUNT(*) FROM product >= 40` 时跳过。每个商品：
- 拉 `https://picsum.photos/seed/p<idx>/600/600` 当主图（4 张：主图 + 3 张详情）
- 上传 MinIO `products/seed/<idx>-<n>.jpg`
- DB 插入：images 字段写 `["url1","url2",...]` JSON 数组（按现有 product 表 images 字段格式）

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall/mall-product-rpc && go build ./cmd/seed
cd /home/carter/workspace/go/yw-mall
git add mall-product-rpc/cmd/seed/ mall-product-rpc/etc/product.yaml
git commit -m "feat(product-rpc): seed 40 products with images via minio"
```

---

## Task 17: user-rpc 地址 seed + order seed + 扩展 review/logistics seed

**Files:**
- Create: `mall-user-rpc/cmd/seed/main.go`（地址；如已存在用户 seed 则扩展）
- Create: `mall-order-rpc/cmd/seed/main.go`
- Modify: `mall-review-rpc/cmd/seed/main.go`（如存在）— 补充挂订单的逻辑
- Modify: `mall-logistics-rpc/cmd/seed/main.go`（如存在）

- [ ] **Step 1: address seed**

`mall-user-rpc/cmd/seed/main.go`（如果该文件已经做用户 seed 就追加 address 部分）：

```go
// 给 user_id 1, 2, 3 各加 1-2 条地址
addrs := []struct { UserId int64; Receiver, Phone, Prov, City, Dist, Detail string; IsDefault bool }{
	{1, "Alice 张", "13800000001", "北京市", "北京市", "朝阳区", "三里屯太古里 1 号", true},
	{1, "Alice 张（公司）", "13800000001", "北京市", "北京市", "海淀区", "中关村大街 1 号", false},
	{2, "Bob 李", "13800000002", "上海市", "上海市", "浦东新区", "陆家嘴环路 100 号", true},
	{3, "Carol 王", "13800000003", "广东省", "深圳市", "南山区", "深南大道 1000 号", true},
	// ... 共 10-15 条
}
```

幂等：`COUNT(*) FROM user_address >= 10` 时跳过。

- [ ] **Step 2: order seed**

`mall-order-rpc/cmd/seed/main.go`：
- 创建 20 条订单分散在 user 1-3、各种 status (1/2/3/4/0 比例 4/4/4/4/4)
- 已发货的订单（status >= 3）写 receiver 快照（取该 user 的默认地址）
- 幂等：`COUNT(*) FROM \`order\` >= 20` 时跳过

> 因为 order 表 receiver 字段允许 ''，旧数据可不设；新 seed 数据全部带快照。

- [ ] **Step 3: 扩展 review/logistics seed**

确保 review seed 挂到 order_id（要求 order 已存在）；logistics seed 挂到已发货 order_id。
顺序：order seed 必须在 review/logistics seed 之前——start.sh 里编排好。

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
go build ./mall-user-rpc/cmd/seed ./mall-order-rpc/cmd/seed
git add mall-user-rpc/cmd/seed/ mall-order-rpc/cmd/seed/ \
        mall-review-rpc/cmd/seed/ mall-logistics-rpc/cmd/seed/
git commit -m "feat(seed): user addresses + orders + review/logistics linkage"
```

---

## Task 18: start.sh 接入 mall-shop-rpc + mall_shop DB + seed pipeline

**Files:**
- Modify: `start.sh`

- [ ] **Step 1: SERVICES 数组加 mall-shop-rpc（在 product-rpc 之前）**

```bash
SERVICES=(
    "mall-user-rpc:user.go:user-rpc:19001"
    "mall-shop-rpc:shop.go:shop-rpc:9017"            # ← 新增（在 product 之前）
    "mall-product-rpc:product.go:product-rpc:9002"
    # ... 其余不变
)
```

- [ ] **Step 2: bootstrap_dbs 加 mall_shop**

```bash
for db in mall_user mall_product mall_order mall_cart mall_payment \
          mall_activity mall_rule mall_workflow mall_reward mall_risk \
          mall_review mall_logistics mall_shop; do
```

DDL map 加：

```bash
[mall_shop]=mall-shop-rpc/sql/shop.sql
```

- [ ] **Step 3: do_nuke 同步加 mall_shop**

```bash
for db in mall_user mall_product ... mall_review mall_logistics mall_shop; do
```

- [ ] **Step 4: bootstrap_seed 追加 6 个 seed 调用**

在现有 workflow/reward/activity seed 之后追加：

```bash
log "Seeding shops..."
( cd "$BASE_DIR/mall-shop-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "shop seed had errors"
log "Seeding products..."
( cd "$BASE_DIR/mall-product-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "product seed had errors"
log "Seeding addresses..."
( cd "$BASE_DIR/mall-user-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "user seed had errors"
log "Seeding orders..."
( cd "$BASE_DIR/mall-order-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "order seed had errors"
log "Seeding reviews..."
( cd "$BASE_DIR/mall-review-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "review seed had errors"
log "Seeding logistics..."
( cd "$BASE_DIR/mall-logistics-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "logistics seed had errors"
```

- [ ] **Step 5: 语法检查 + Commit**

```bash
bash -n start.sh
cd /home/carter/workspace/go/yw-mall
git add start.sh
git commit -m "ops: register mall-shop-rpc + mall_shop schema + extended seed pipeline"
```

---

## Task 19: Seed 幂等性脚本

**Files:**
- Create: `scripts/check_seed_idempotency.sh`

- [ ] **Step 1: 写脚本**

```bash
#!/bin/bash
# 检查 seed 跑两次后数据量保持不变（幂等性）
set -e
PROXY_MYSQL='mysql -h127.0.0.1 -P6033 -uproxysql -pproxysql123'

count() {
  $PROXY_MYSQL "$1" -e "SELECT COUNT(*) FROM $2" 2>/dev/null | tail -1
}

before=$(cat <<EOF
shop:$(count mall_shop shop)
product:$(count mall_product product)
addr:$(count mall_user user_address)
order:$(count mall_order \`order\`)
EOF
)
echo "Before second seed run:"
echo "$before"

# 重跑 seed
( cd /home/carter/workspace/go/yw-mall/mall-shop-rpc     && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-product-rpc  && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-user-rpc     && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-order-rpc    && go run cmd/seed/main.go )

after=$(cat <<EOF
shop:$(count mall_shop shop)
product:$(count mall_product product)
addr:$(count mall_user user_address)
order:$(count mall_order \`order\`)
EOF
)
echo "After:"
echo "$after"

if [ "$before" = "$after" ]; then
  echo "✓ idempotent"
else
  echo "✗ counts changed — seed is NOT idempotent"
  exit 1
fi
```

- [ ] **Step 2: chmod + Commit**

```bash
chmod +x scripts/check_seed_idempotency.sh
cd /home/carter/workspace/go/yw-mall
git add scripts/
git commit -m "ops: add seed idempotency check script"
```

---

## Task 20: 文档化 E2E 验证流程（手动，不实际跑）

**Files:**
- Create: `docs/superpowers/notes/2026-05-02-frontend-prep-e2e-checklist.md`

- [ ] **Step 1: 写 checklist**

```markdown
# 子项目 1 / E2E 验证清单

> 该子项目的代码改动需要前端接入才能完整验证。本清单列出 curl 命令，可在 start.sh start 之后单独跑。

## 前置

```bash
cd /home/carter/workspace/go/env && docker compose up -d
cd /home/carter/workspace/go/yw-mall && ./start.sh nuke && ./start.sh start
sleep 15
./start.sh status | grep -E 'shop|product|user|order'
```

## 1. shop 列表

```bash
curl -s http://127.0.0.1:18888/api/shop/list | jq
curl -s http://127.0.0.1:18888/api/shop/recommended | jq
curl -s http://127.0.0.1:18888/api/shop/detail/1 | jq
curl -s http://127.0.0.1:18888/api/shop/products/1 | jq '.products | length'   # 应 ≥ 6
```

## 2. 用户登录 + 地址

```bash
TOKEN=$(curl -s -X POST http://127.0.0.1:18888/api/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"alice123"}' | jq -r .token)

curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/address/list | jq
ADDR_ID=$(curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/address/list | jq -r '.addresses[0].id')

curl -s -X POST http://127.0.0.1:18888/api/address/add \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"receiverName":"测试","phone":"13900000099","province":"北京","city":"北京","district":"朝阳","detail":"望京 SOHO"}'
```

## 3. 关注店铺

```bash
curl -s -X POST -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/shop/1/follow
curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/shop/1/is-following | jq
curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/shop/my-followed | jq
```

## 4. 下单（含地址快照）

```bash
ORDER_RESP=$(curl -s -X POST http://127.0.0.1:18888/api/order/create \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"addressId\":$ADDR_ID,\"items\":[{\"productId\":1,\"productName\":\"AirPods\",\"price\":149900,\"quantity\":1}]}")
ORDER_ID=$(echo $ORDER_RESP | jq -r .id)

curl -s -H "Authorization: Bearer $TOKEN" http://127.0.0.1:18888/api/order/detail/$ORDER_ID | jq
# 应有 receiverName / receiverPhone / receiverDetail 字段
```

## 5. 幂等校验

```bash
bash scripts/check_seed_idempotency.sh
```

## 6. MinIO 图片可访问性

```bash
curl -sI http://127.0.0.1:9000/mall-media/shops/seed/1-banner.jpg | head -1
# HTTP/1.1 200 OK
```
```

- [ ] **Step 2: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add docs/superpowers/notes/
git commit -m "docs: e2e checklist for frontend-prep subproject"
```

---

## 自审备注

- **Spec 覆盖**：spec §1-§7 每节都有对应 task。§5 seed → tasks 15-19；§6.1 测试 → 各 logic task 内含编译验证 + task 19 幂等脚本；§6.2 风险 → task 18 SERVICES 顺序写明 product 之前。
- **占位符**：无 TBD/TODO；商品命名 task 16 步骤 1 标了 "自行扩展到 40 个" — **这是允许的扩展点而非未填空**，提示实施者按 6 家店各 6-7 个的规模补足。
- **类型一致性**：`ShopDTO`（api types）↔ `Shop`（proto）；`AddressDTO` ↔ `Address`；`protoShopToDTO` / `protoAddressToDTO` 是命名一致的转换 helper。
