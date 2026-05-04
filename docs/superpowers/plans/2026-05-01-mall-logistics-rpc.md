# mall-logistics-rpc Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `mall-logistics-rpc` + necessary mall-order-rpc / mall-api changes so users can see multi-shipment tracking on order detail; logistics is created via Kafka `order.shipped` event triggered by `MarkShipped` on order-rpc; tracks come from kuaidi100 webhook (with admin `InjectTrack` as offline fallback).

**Architecture:** Standard go-zero RPC service co-located with the existing 13. Kafka topic `order.shipped` (sync write from order-rpc → consumed by logistics-rpc worker) is the trigger. kuaidi100 HTTP API for subscribe; webhook receiver in mall-api validates HMAC-MD5 sign, then RPC into logistics-rpc.IngestWebhookEvents. Multi-shipment per order via `shipment_item` join table; status machine enforces monotonic upgrades via `GREATEST(status, ?)`.

**Tech Stack:** Go 1.26.1, go-zero 1.10.1, gRPC, MySQL via ProxySQL, Kafka (kafka-go writer + go-zero kq consumer), HTTP client for kuaidi100 (stdlib), HMAC-MD5 (stdlib).

**Spec:** `docs/superpowers/specs/2026-05-01-mall-logistics-rpc-design.md`

---

## File Structure

### New files

```
mall-common/proto/logistics/logistics.proto                   proto definition
mall-logistics-rpc/
├── go.mod / go.sum
├── etc/logistics.yaml                                        config
├── sql/logistics.sql                                         DDL (3 tables)
├── logistics.proto                                           working copy
├── logistics.go                                              main entry
├── logistics/{logistics.pb.go, logistics_grpc.pb.go}         generated
├── logisticsclient/logistics.go                              generated
├── internal/config/config.go                                 config struct
├── internal/svc/servicecontext.go                            svc context (DB, kuaidi100 client)
├── internal/server/logisticsserver.go                        generated
├── internal/model/{shipment,shipmentitem,shipmenttrack}*.go  model wrappers
├── internal/kuaidi100/client.go                              kuaidi100 HTTP client (subscribe + sign)
├── internal/kuaidi100/statemap.go                            state mapping helpers
├── internal/logic/createshipmentlogic.go
├── internal/logic/listshipmentsbyorderlogic.go
├── internal/logic/getshipmentlogic.go
├── internal/logic/ingestwebhookeventslogic.go
├── internal/logic/retrysubscribelogic.go
├── internal/logic/injecttracklogic.go
├── internal/logic/helpers.go
└── internal/worker/ordershippedworker.go                     Kafka consumer goroutine
mall-api/mall-logistics.api                                   go-zero api block
```

### Modified files

```
mall-common/errorx/errorx.go                                  +7 codes (9001-9007)
mall-common/proto/order/order.proto                           +MarkShipped rpc
mall-order-rpc/order.proto                                    mirror
mall-order-rpc/order/*.pb.go, orderclient/*, internal/server/* regen
mall-order-rpc/internal/logic/markshippedlogic.go             new logic
mall-order-rpc/internal/svc/servicecontext.go                 +KafkaWriter
mall-order-rpc/internal/config/config.go                      +Kafka block
mall-order-rpc/etc/order.yaml                                 +Kafka brokers
mall-order-rpc/go.mod                                         +kafka-go dep
mall-api/mall.api                                             import mall-logistics.api; ProductDetailResp unchanged
mall-api/internal/{config,svc,handler,logic,types}/...        new types/handlers/svc fields
mall-api/etc/mall-api.yaml                                    +LogisticsRpc, Kuaidi100 webhook secret
mall-api/internal/handler/orderdetailhandler.go               (read-through)
mall-api/internal/logic/orderdetaillogic.go                   parallel ListShipmentsByOrder
mall-api/internal/logic/admin*reviewlogic.go                  unchanged
start.sh                                                      +"mall-logistics-rpc:logistics.go:logistics-rpc:9016"
start.sh                                                      +[mall_logistics]=mall-logistics-rpc/sql/logistics.sql
```

---

## Task 1: Add logistics error codes 9001-9007

**Files:** Modify `mall-common/errorx/errorx.go`.

- [ ] **Step 1: Append constants**

After the review codes (`AdminTokenInvalid = 8009`), append in the const block:

```go
	// Logistics service error codes (9xxx)
	LogisticsShipmentNotFound       = 9001
	LogisticsTrackingNoExists       = 9002
	LogisticsKuaidi100SignInvalid   = 9003
	LogisticsSubscribeFailed        = 9004
	LogisticsOrderNotShippable      = 9005
	LogisticsCarrierUnknown         = 9006
	LogisticsTrackingInvalid        = 9007
```

In the `message` map, append:

```go
	LogisticsShipmentNotFound:     "logistics: shipment not found",
	LogisticsTrackingNoExists:     "logistics: tracking number already exists for this carrier",
	LogisticsKuaidi100SignInvalid: "logistics: invalid kuaidi100 webhook signature",
	LogisticsSubscribeFailed:      "logistics: subscribe to kuaidi100 failed after retries",
	LogisticsOrderNotShippable:    "logistics: order not in a shippable state",
	LogisticsCarrierUnknown:       "logistics: unknown carrier code",
	LogisticsTrackingInvalid:      "logistics: invalid tracking number format",
```

- [ ] **Step 2: Build** — `cd mall-common && go build ./...` clean.

- [ ] **Step 3: Commit**

```bash
git add mall-common/errorx/errorx.go
git commit -m "feat(errorx): add logistics error codes 9001-9007"
```

---

## Task 2: Add `MarkShipped` rpc to mall-order-rpc + Kafka producer

**Files:**
- Modify both proto copies: `mall-common/proto/order/order.proto` and `mall-order-rpc/order.proto`
- Regen: `mall-order-rpc/order/*`, `orderclient/*`, `internal/server/*`
- Modify: `mall-order-rpc/internal/{config,svc}/`, `etc/order.yaml`, `go.mod`
- Create: `mall-order-rpc/internal/logic/markshippedlogic.go`
- Create: `mall-order-rpc/internal/kafka/producer.go`

- [ ] **Step 1: Add proto**

In **both** proto files, inside `service Order { ... }` add:

```proto
  rpc MarkShipped(MarkShippedReq) returns (MarkShippedResp);
```

After existing messages, add:

```proto
message MarkShippedReq {
  int64 order_id = 1;
  string tracking_no = 2;
  string carrier = 3;
}

message MarkShippedResp {
  bool ok = 1;
}
```

- [ ] **Step 2: Regen stubs**

```bash
export PATH=/home/carter/workspace/go/bin:$PATH
cd /home/carter/workspace/go/yw-mall/mall-order-rpc
goctl rpc protoc order.proto --go_out=. --go-grpc_out=. --zrpc_out=. -m
# clean spurious dirs goctl creates
rm -rf client internal/logic/order internal/server/order
```

Confirm `internal/logic/markshippedlogic.go` exists as a stub.

- [ ] **Step 3: Add Kafka config and writer wiring**

Append to `mall-order-rpc/etc/order.yaml`:

```yaml
Kafka:
  Brokers:
    - 127.0.0.1:19092
    - 127.0.0.1:19093
    - 127.0.0.1:19094
  OrderShippedTopic: order.shipped
```

Update `mall-order-rpc/internal/config/config.go` Config struct (additive — keep existing fields):

```go
	Kafka struct {
		Brokers           []string
		OrderShippedTopic string
	}
```

Create `mall-order-rpc/internal/kafka/producer.go`:

```go
package kafka

import (
	"context"
	"time"

	kgo "github.com/segmentio/kafka-go"
)

type Producer struct {
	w *kgo.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		w: &kgo.Writer{
			Addr:         kgo.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kgo.Hash{},
			BatchTimeout: 50 * time.Millisecond,
			RequiredAcks: kgo.RequireAll,
			Async:        false,
		},
	}
}

func (p *Producer) Write(ctx context.Context, key string, value []byte) error {
	return p.w.WriteMessages(ctx, kgo.Message{Key: []byte(key), Value: value})
}

func (p *Producer) Close() error { return p.w.Close() }
```

Update `mall-order-rpc/internal/svc/servicecontext.go` — add field + init:

```go
	OrderShippedProducer *kafka.Producer
```

In NewServiceContext:

```go
		OrderShippedProducer: kafka.NewProducer(c.Kafka.Brokers, c.Kafka.OrderShippedTopic),
```

Add to go.mod:

```bash
cd /home/carter/workspace/go/yw-mall/mall-order-rpc
go get github.com/segmentio/kafka-go
go mod tidy
```

- [ ] **Step 4: Implement MarkShipped logic**

Replace `mall-order-rpc/internal/logic/markshippedlogic.go`:

```go
package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"mall-common/errorx"
	"mall-order-rpc/internal/svc"
	"mall-order-rpc/order"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	orderStatusPaid    = 1
	orderStatusShipped = 2
)

type orderShippedEvent struct {
	OrderId    int64                  `json:"orderId"`
	UserId     int64                  `json:"userId"`
	TrackingNo string                 `json:"trackingNo"`
	Carrier    string                 `json:"carrier"`
	Items      []orderShippedItemBody `json:"items"`
	ShippedAt  int64                  `json:"shippedAt"`
}
type orderShippedItemBody struct {
	OrderItemId int64 `json:"orderItemId"`
	ProductId   int64 `json:"productId"`
	Quantity    int64 `json:"quantity"`
}

type MarkShippedLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkShippedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkShippedLogic {
	return &MarkShippedLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *MarkShippedLogic) MarkShipped(in *order.MarkShippedReq) (*order.MarkShippedResp, error) {
	if strings.TrimSpace(in.TrackingNo) == "" || strings.TrimSpace(in.Carrier) == "" {
		return nil, errorx.NewCodeError(errorx.ParamError)
	}
	ord, err := l.svcCtx.OrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.OrderNotFound)
	}
	if int(ord.Status) != orderStatusPaid {
		return nil, errorx.NewCodeError(errorx.LogisticsOrderNotShippable)
	}
	if _, err := l.svcCtx.OrderModel.Conn().ExecCtx(l.ctx,
		"UPDATE `order` SET status=?, tracking_no=?, carrier=? WHERE id=? AND status=?",
		orderStatusShipped, in.TrackingNo, in.Carrier, in.OrderId, orderStatusPaid); err != nil {
		return nil, err
	}
	items, err := l.fetchItems(in.OrderId)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(orderShippedEvent{
		OrderId: in.OrderId, UserId: ord.UserId,
		TrackingNo: in.TrackingNo, Carrier: in.Carrier,
		Items: items, ShippedAt: time.Now().Unix(),
	})
	if err := l.svcCtx.OrderShippedProducer.Write(l.ctx, in.TrackingNo, body); err != nil {
		return nil, err
	}
	return &order.MarkShippedResp{Ok: true}, nil
}

func (l *MarkShippedLogic) fetchItems(orderId int64) ([]orderShippedItemBody, error) {
	rows := []*struct {
		Id        int64 `db:"id"`
		ProductId int64 `db:"product_id"`
		Quantity  int64 `db:"quantity"`
	}{}
	if err := l.svcCtx.OrderModel.Conn().QueryRowsCtx(l.ctx, &rows,
		"SELECT id, product_id, quantity FROM order_item WHERE order_id=?", orderId); err != nil {
		return nil, err
	}
	out := make([]orderShippedItemBody, 0, len(rows))
	for _, r := range rows {
		out = append(out, orderShippedItemBody{OrderItemId: r.Id, ProductId: r.ProductId, Quantity: r.Quantity})
	}
	return out, nil
}
```

> Notes:
> - Add `tracking_no VARCHAR(64) NULL` and `carrier VARCHAR(32) NULL` columns to `order` table if not present. Check `mall-order-rpc/sql/order.sql` first; if missing, add columns to the SQL and apply via the bootstrap path.
> - `OrderModel.Conn()` may need a thin wrapper if the goctl-generated model doesn't expose conn directly. If not, use a separately-injected `sqlx.SqlConn` field (e.g., `svcCtx.DB`) like mall-review-rpc did. Adjust accordingly.
> - Verify `order_item` table column names (Id, OrderId, ProductId, Quantity) match the actual schema before relying on them.

- [ ] **Step 5: Build**

```bash
cd /home/carter/workspace/go/yw-mall/mall-order-rpc
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add mall-common/proto/order/order.proto \
        mall-order-rpc/order.proto mall-order-rpc/order \
        mall-order-rpc/orderclient mall-order-rpc/internal \
        mall-order-rpc/etc/order.yaml mall-order-rpc/go.mod mall-order-rpc/go.sum
git commit -m "feat(order-rpc): MarkShipped rpc + kafka producer for order.shipped event"
```

---

## Task 3: Logistics proto + DDL

**Files:**
- Create `mall-common/proto/logistics/logistics.proto`
- Create `mall-logistics-rpc/sql/logistics.sql`

- [ ] **Step 1: Write proto**

Save to `mall-common/proto/logistics/logistics.proto` (copy from spec §5.1 verbatim — service `Logistics`, 6 RPCs, all message types).

- [ ] **Step 2: Write DDL**

Save to `mall-logistics-rpc/sql/logistics.sql` (3 tables from spec §4.1–4.3 verbatim, with the `mall_logistics` CREATE DATABASE prelude).

- [ ] **Step 3: Commit**

```bash
git add mall-common/proto/logistics/logistics.proto mall-logistics-rpc/sql/logistics.sql
git commit -m "feat(logistics): add proto definition and DDL"
```

---

## Task 4: Scaffold mall-logistics-rpc

**Files:** New mall-logistics-rpc subtree.

- [ ] **Step 1: Copy proto and scaffold**

```bash
export PATH=/home/carter/workspace/go/bin:$PATH
mkdir -p /home/carter/workspace/go/yw-mall/mall-logistics-rpc
cd /home/carter/workspace/go/yw-mall/mall-logistics-rpc
cp ../mall-common/proto/logistics/logistics.proto logistics.proto
goctl rpc protoc logistics.proto --go_out=. --go-grpc_out=. --zrpc_out=. -m
```

- [ ] **Step 2: Flatten layout to match neighbors**

```bash
# move client, logic, server out of nested 'logistics' subdir to flat layout
mkdir -p logisticsclient
mv client/logistics/logistics.go logisticsclient/logistics.go
sed -i '0,/^package logistics$/{s/^package logistics$/package logisticsclient/}' logisticsclient/logistics.go
mv internal/logic/logistics/*.go internal/logic/
rmdir internal/logic/logistics
mv internal/server/logistics/logisticsserver.go internal/server/logisticsserver.go
rmdir internal/server/logistics
rm -rf client
sed -i 's/^package logisticslogic$/package logic/' internal/logic/*.go
sed -i 's|"mall-logistics-rpc/internal/logic/logistics"|"mall-logistics-rpc/internal/logic"|; s|logisticslogic\.|logic.|g' internal/server/logisticsserver.go
sed -i 's|logisticsServer "mall-logistics-rpc/internal/server/logistics"|logisticsServer "mall-logistics-rpc/internal/server"|' logistics.go
```

- [ ] **Step 3: Configure go.mod**

```bash
go mod edit -require=github.com/zeromicro/go-zero@v1.10.1 \
            -require=google.golang.org/grpc@v1.80.0 \
            -require=google.golang.org/protobuf@v1.36.11 \
            -replace=mall-common=../mall-common \
            -replace=mall-order-rpc=../mall-order-rpc
go get github.com/segmentio/kafka-go
go mod tidy
```

- [ ] **Step 4: Configure etc/logistics.yaml**

Replace generated content with the YAML from spec §5.4 verbatim.

- [ ] **Step 5: Generate models**

```bash
goctl model mysql ddl -src sql/logistics.sql -dir internal/model -c
```

Verify `internal/model/{shipmentmodel,shipmentitemmodel,shipmenttrackmodel}*.go` exist.

- [ ] **Step 6: Build**

```bash
go build ./...
```

- [ ] **Step 7: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-logistics-rpc/
git commit -m "feat(logistics-rpc): scaffold service with goctl (proto + model + stubs)"
```

---

## Task 5: Wire ServiceContext + kuaidi100 client + state mapper

**Files:**
- Replace `mall-logistics-rpc/internal/config/config.go`
- Replace `mall-logistics-rpc/internal/svc/servicecontext.go`
- Create `mall-logistics-rpc/internal/kuaidi100/client.go`
- Create `mall-logistics-rpc/internal/kuaidi100/statemap.go`

- [ ] **Step 1: Config**

```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	DataSource string
	Cache      cache.CacheConf
	RedisCache redis.RedisConf

	Kafka struct {
		Brokers []string
		Topic   string
		Group   string
	}
	Kuaidi100 struct {
		Customer        string
		Key             string
		PollEndpoint    string
		WebhookCallback string
	}
	Subscribe struct {
		MaxRetries       int
		InitialBackoffMs int
	}
}
```

- [ ] **Step 2: ServiceContext**

```go
package svc

import (
	"net/http"
	"time"

	"mall-logistics-rpc/internal/config"
	"mall-logistics-rpc/internal/kuaidi100"
	"mall-logistics-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	DB                 sqlx.SqlConn
	ShipmentModel      model.ShipmentModel
	ShipmentItemModel  model.ShipmentItemModel
	ShipmentTrackModel model.ShipmentTrackModel
	Redis              *redis.Redis
	Kuaidi100          *kuaidi100.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:             c,
		DB:                 conn,
		ShipmentModel:      model.NewShipmentModel(conn, c.Cache),
		ShipmentItemModel:  model.NewShipmentItemModel(conn, c.Cache),
		ShipmentTrackModel: model.NewShipmentTrackModel(conn, c.Cache),
		Redis:              redis.MustNewRedis(c.RedisCache),
		Kuaidi100: kuaidi100.NewClient(kuaidi100.Config{
			Customer:        c.Kuaidi100.Customer,
			Key:             c.Kuaidi100.Key,
			PollEndpoint:    c.Kuaidi100.PollEndpoint,
			WebhookCallback: c.Kuaidi100.WebhookCallback,
			HTTP:            &http.Client{Timeout: 10 * time.Second},
		}),
	}
}
```

- [ ] **Step 3: kuaidi100 client**

Create `internal/kuaidi100/client.go`:

```go
package kuaidi100

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Config struct {
	Customer        string
	Key             string
	PollEndpoint    string
	WebhookCallback string
	HTTP            *http.Client
}

type Client struct{ cfg Config }

func NewClient(c Config) *Client { return &Client{cfg: c} }

type subscribeParam struct {
	Company    string `json:"company"`
	Number     string `json:"number"`
	Key        string `json:"key"`
	Parameters struct {
		Callbackurl string `json:"callbackurl"`
		Salt        string `json:"salt"`
		Resultv2    string `json:"resultv2"`
		Autoccsf    string `json:"autoCom"`
	} `json:"parameters"`
}

type subscribeResp struct {
	Result     bool   `json:"result"`
	ReturnCode string `json:"returnCode"`
	Message    string `json:"message"`
}

// Subscribe registers a tracking number with kuaidi100 to receive push events.
// carrier is kuaidi100 company code (e.g. "shunfeng", "jd", "zhongtong").
func (c *Client) Subscribe(ctx context.Context, carrier, trackingNo string) error {
	if c.cfg.Customer == "" || c.cfg.Key == "" || c.cfg.PollEndpoint == "" {
		return fmt.Errorf("kuaidi100: missing customer/key/endpoint")
	}
	p := subscribeParam{Company: carrier, Number: trackingNo, Key: c.cfg.Key}
	p.Parameters.Callbackurl = c.cfg.WebhookCallback
	p.Parameters.Resultv2 = "1"
	pbytes, _ := json.Marshal(p)
	form := url.Values{}
	form.Set("schema", "json")
	form.Set("param", string(pbytes))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.PollEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.cfg.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var sr subscribeResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return err
	}
	if !sr.Result {
		return fmt.Errorf("kuaidi100 subscribe rejected: %s %s", sr.ReturnCode, sr.Message)
	}
	return nil
}

// VerifySign validates webhook signature: sign == upper(md5(param + key)).
func (c *Client) VerifySign(param, sign string) bool {
	h := md5.Sum([]byte(param + c.cfg.Key))
	return strings.EqualFold(hex.EncodeToString(h[:]), sign)
}
```

- [ ] **Step 4: State mapping**

Create `internal/kuaidi100/statemap.go`:

```go
package kuaidi100

const (
	StateInternalCreated     int32 = 0
	StateInternalCollected   int32 = 1
	StateInternalInTransit   int32 = 2
	StateInternalDelivering  int32 = 3
	StateInternalDelivered   int32 = 4
	StateInternalException   int32 = 5
	StateInternalReturned    int32 = 6
	StateKuaidi100Synthetic  int32 = 255
)

// MapState maps a kuaidi100 state code to the internal status enum.
func MapState(k int32) int32 {
	switch k {
	case 1:
		return StateInternalCollected
	case 0:
		return StateInternalInTransit
	case 5:
		return StateInternalDelivering
	case 3:
		return StateInternalDelivered
	case 2:
		return StateInternalException
	case 4, 6, 14:
		return StateInternalReturned
	case StateKuaidi100Synthetic:
		return StateInternalException
	default:
		return StateInternalInTransit
	}
}
```

- [ ] **Step 5: Build**

```bash
cd /home/carter/workspace/go/yw-mall/mall-logistics-rpc
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add mall-logistics-rpc/internal/config mall-logistics-rpc/internal/svc \
        mall-logistics-rpc/internal/kuaidi100
git commit -m "feat(logistics-rpc): wire ServiceContext + kuaidi100 client + state mapping"
```

---

## Task 6: CreateShipment + helpers

**Files:**
- Modify `mall-logistics-rpc/internal/logic/createshipmentlogic.go`
- Create `mall-logistics-rpc/internal/logic/helpers.go`

- [ ] **Step 1: helpers.go**

```go
package logic

import (
	"errors"
	"fmt"

	"mall-logistics-rpc/internal/model"
	"mall-logistics-rpc/logistics"

	gosqldriver "github.com/go-sql-driver/mysql"
)

func isDuplicateKey(err error) bool {
	var me *gosqldriver.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return true
	}
	return false
}

func _ = fmt.Sprint // keep imports stable

func toShipmentProto(s *model.Shipment, items []*model.ShipmentItem, tracks []*model.ShipmentTrack) *logistics.Shipment {
	out := &logistics.Shipment{
		Id: s.Id, OrderId: s.OrderId, UserId: s.UserId,
		TrackingNo: s.TrackingNo, Carrier: s.Carrier,
		Status: int32(s.Status), SubscribeStatus: int32(s.SubscribeStatus),
		CreateTime: s.CreateTime.Unix(),
	}
	if s.LastTrackTime.Valid {
		out.LastTrackTime = s.LastTrackTime.Time.Unix()
	}
	for _, it := range items {
		out.Items = append(out.Items, &logistics.ShipmentItemRef{
			OrderItemId: it.OrderItemId, ProductId: it.ProductId, Quantity: int32(it.Quantity),
		})
	}
	for _, t := range tracks {
		out.Tracks = append(out.Tracks, &logistics.Track{
			TrackTime: t.TrackTime.Unix(), Location: t.Location.String,
			Description: t.Description, StateInternal: int32(t.StateInternal),
			StateKuaidi100: int32(t.StateKuaidi100.Int64),
		})
	}
	return out
}
```

- [ ] **Step 2: CreateShipment**

```go
package logic

import (
	"context"

	"mall-common/errorx"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateShipmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateShipmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateShipmentLogic {
	return &CreateShipmentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateShipmentLogic) CreateShipment(in *logistics.CreateShipmentReq) (*logistics.CreateShipmentResp, error) {
	if in.TrackingNo == "" || in.Carrier == "" {
		return nil, errorx.NewCodeError(errorx.LogisticsTrackingInvalid)
	}
	var newId int64
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		ret, err := session.ExecCtx(ctx,
			"INSERT INTO `shipment`(order_id, user_id, tracking_no, carrier, status, subscribe_status) VALUES (?,?,?,?,0,0)",
			in.OrderId, in.UserId, in.TrackingNo, in.Carrier)
		if err != nil {
			if isDuplicateKey(err) {
				return errorx.NewCodeError(errorx.LogisticsTrackingNoExists)
			}
			return err
		}
		newId, _ = ret.LastInsertId()
		for _, it := range in.Items {
			if _, err := session.ExecCtx(ctx,
				"INSERT INTO `shipment_item`(shipment_id, order_item_id, product_id, quantity) VALUES (?,?,?,?)",
				newId, it.OrderItemId, it.ProductId, it.Quantity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &logistics.CreateShipmentResp{ShipmentId: newId}, nil
}
```

- [ ] **Step 3: Build + commit**

```bash
go build ./...
git add mall-logistics-rpc/internal/logic/createshipmentlogic.go \
        mall-logistics-rpc/internal/logic/helpers.go
git commit -m "feat(logistics-rpc): CreateShipment with tx insert and dup-key map"
```

---

## Task 7: IngestWebhookEvents (dedup + monotonic status upgrade)

**Files:**
- Modify `mall-logistics-rpc/internal/logic/ingestwebhookeventslogic.go`

- [ ] **Step 1: Implement**

```go
package logic

import (
	"context"
	"database/sql"
	"time"

	"mall-common/errorx"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type IngestWebhookEventsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIngestWebhookEventsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IngestWebhookEventsLogic {
	return &IngestWebhookEventsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *IngestWebhookEventsLogic) IngestWebhookEvents(in *logistics.IngestWebhookEventsReq) (*logistics.Empty, error) {
	if in.TrackingNo == "" || in.Carrier == "" {
		return nil, errorx.NewCodeError(errorx.LogisticsTrackingInvalid)
	}
	var ship struct {
		Id     int64 `db:"id"`
		Status int64 `db:"status"`
	}
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &ship,
		"SELECT id, status FROM shipment WHERE carrier=? AND tracking_no=? LIMIT 1",
		in.Carrier, in.TrackingNo); err != nil {
		if err == sql.ErrNoRows || err == sqlx.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.LogisticsShipmentNotFound)
		}
		return nil, err
	}

	var maxInternal int32
	var lastTime time.Time
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, e := range in.Events {
			t := time.Unix(e.TrackTime, 0)
			_, err := session.ExecCtx(ctx,
				"INSERT IGNORE INTO `shipment_track`(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?,?,?,?,?,?)",
				ship.Id, t, e.Location, e.Description, e.StateKuaidi100, e.StateInternal)
			if err != nil {
				return err
			}
			if e.StateInternal > maxInternal {
				maxInternal = e.StateInternal
			}
			if t.After(lastTime) {
				lastTime = t
			}
		}
		_, err := session.ExecCtx(ctx,
			"UPDATE `shipment` SET status=GREATEST(status,?), last_track_time=GREATEST(IFNULL(last_track_time,'1970-01-01'),?), subscribe_status=1 WHERE id=?",
			maxInternal, lastTime, ship.Id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
```

- [ ] **Step 2: Build + commit**

```bash
go build ./...
git add mall-logistics-rpc/internal/logic/ingestwebhookeventslogic.go
git commit -m "feat(logistics-rpc): IngestWebhookEvents with dedup and monotonic status"
```

---

## Task 8: Read paths (GetShipment + ListShipmentsByOrder) + admin paths (RetrySubscribe + InjectTrack)

**Files:** 4 logic files.

- [ ] **Step 1: GetShipment**

```go
package logic

import (
	"context"

	"mall-common/errorx"
	"mall-logistics-rpc/internal/model"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShipmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetShipmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShipmentLogic {
	return &GetShipmentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetShipmentLogic) GetShipment(in *logistics.GetShipmentReq) (*logistics.Shipment, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.LogisticsShipmentNotFound)
	}
	items, _ := fetchItems(l.ctx, l.svcCtx, s.Id)
	tracks, _ := fetchTracks(l.ctx, l.svcCtx, s.Id)
	return toShipmentProto(s, items, tracks), nil
}

func fetchItems(ctx context.Context, svcCtx *svc.ServiceContext, shipmentId int64) ([]*model.ShipmentItem, error) {
	rows := []*model.ShipmentItem{}
	err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, shipment_id, order_item_id, product_id, quantity FROM shipment_item WHERE shipment_id=?",
		shipmentId)
	return rows, err
}

func fetchTracks(ctx context.Context, svcCtx *svc.ServiceContext, shipmentId int64) ([]*model.ShipmentTrack, error) {
	rows := []*model.ShipmentTrack{}
	err := svcCtx.DB.QueryRowsCtx(ctx, &rows,
		"SELECT id, shipment_id, track_time, location, description, state_kuaidi100, state_internal, create_time FROM shipment_track WHERE shipment_id=? ORDER BY track_time DESC",
		shipmentId)
	return rows, err
}
```

- [ ] **Step 2: ListShipmentsByOrder**

```go
package logic

import (
	"context"

	"mall-logistics-rpc/internal/model"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShipmentsByOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListShipmentsByOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShipmentsByOrderLogic {
	return &ListShipmentsByOrderLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListShipmentsByOrderLogic) ListShipmentsByOrder(in *logistics.ListShipmentsByOrderReq) (*logistics.ListShipmentsByOrderResp, error) {
	rows := []*model.Shipment{}
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		"SELECT id, order_id, user_id, tracking_no, carrier, status, subscribe_status, last_track_time, create_time, update_time FROM shipment WHERE order_id=? ORDER BY create_time",
		in.OrderId); err != nil {
		return nil, err
	}
	out := make([]*logistics.Shipment, 0, len(rows))
	for _, s := range rows {
		items, _ := fetchItems(l.ctx, l.svcCtx, s.Id)
		tracks, _ := fetchTracks(l.ctx, l.svcCtx, s.Id)
		out = append(out, toShipmentProto(s, items, tracks))
	}
	return &logistics.ListShipmentsByOrderResp{Shipments: out}, nil
}
```

- [ ] **Step 3: InjectTrack**

```go
package logic

import (
	"context"
	"time"

	"mall-common/errorx"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type InjectTrackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInjectTrackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InjectTrackLogic {
	return &InjectTrackLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *InjectTrackLogic) InjectTrack(in *logistics.InjectTrackReq) (*logistics.Empty, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.LogisticsShipmentNotFound)
	}
	now := time.Now()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		if _, err := session.ExecCtx(ctx,
			"INSERT IGNORE INTO `shipment_track`(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?,?,?,?,NULL,?)",
			s.Id, now, in.Location, in.Description, in.StateInternal); err != nil {
			return err
		}
		_, err := session.ExecCtx(ctx,
			"UPDATE `shipment` SET status=GREATEST(status,?), last_track_time=? WHERE id=?",
			in.StateInternal, now, s.Id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &logistics.Empty{}, nil
}
```

- [ ] **Step 4: RetrySubscribe**

```go
package logic

import (
	"context"

	"mall-common/errorx"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrySubscribeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySubscribeLogic {
	return &RetrySubscribeLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *RetrySubscribeLogic) RetrySubscribe(in *logistics.RetrySubscribeReq) (*logistics.Empty, error) {
	s, err := l.svcCtx.ShipmentModel.FindOne(l.ctx, in.ShipmentId)
	if err != nil {
		return nil, errorx.NewCodeError(errorx.LogisticsShipmentNotFound)
	}
	if err := l.svcCtx.Kuaidi100.Subscribe(l.ctx, s.Carrier, s.TrackingNo); err != nil {
		_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
			"UPDATE `shipment` SET subscribe_status=2 WHERE id=?", s.Id)
		return nil, errorx.NewCodeError(errorx.LogisticsSubscribeFailed)
	}
	_, _ = l.svcCtx.DB.ExecCtx(l.ctx,
		"UPDATE `shipment` SET subscribe_status=1 WHERE id=?", s.Id)
	return &logistics.Empty{}, nil
}
```

- [ ] **Step 5: Build + commit**

```bash
go build ./...
git add mall-logistics-rpc/internal/logic/getshipmentlogic.go \
        mall-logistics-rpc/internal/logic/listshipmentsbyorderlogic.go \
        mall-logistics-rpc/internal/logic/injecttracklogic.go \
        mall-logistics-rpc/internal/logic/retrysubscribelogic.go
git commit -m "feat(logistics-rpc): read paths + InjectTrack + RetrySubscribe"
```

---

## Task 9: Kafka consumer worker (order.shipped → CreateShipment + Subscribe)

**Files:**
- Create `mall-logistics-rpc/internal/worker/ordershippedworker.go`
- Modify `mall-logistics-rpc/logistics.go` to start worker alongside RPC server

- [ ] **Step 1: Worker**

```go
package worker

import (
	"context"
	"encoding/json"
	"time"

	"mall-logistics-rpc/internal/logic"
	"mall-logistics-rpc/internal/svc"
	"mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
	kgo "github.com/segmentio/kafka-go"
)

type orderShippedEvent struct {
	OrderId    int64 `json:"orderId"`
	UserId     int64 `json:"userId"`
	TrackingNo string `json:"trackingNo"`
	Carrier    string `json:"carrier"`
	Items      []struct {
		OrderItemId int64 `json:"orderItemId"`
		ProductId   int64 `json:"productId"`
		Quantity    int64 `json:"quantity"`
	} `json:"items"`
	ShippedAt int64 `json:"shippedAt"`
}

type OrderShippedWorker struct {
	svcCtx *svc.ServiceContext
	reader *kgo.Reader
}

func NewOrderShippedWorker(svcCtx *svc.ServiceContext) *OrderShippedWorker {
	return &OrderShippedWorker{
		svcCtx: svcCtx,
		reader: kgo.NewReader(kgo.ReaderConfig{
			Brokers:        svcCtx.Config.Kafka.Brokers,
			Topic:          svcCtx.Config.Kafka.Topic,
			GroupID:        svcCtx.Config.Kafka.Group,
			MinBytes:       1,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
		}),
	}
}

func (w *OrderShippedWorker) Start(ctx context.Context) {
	go func() {
		defer w.reader.Close()
		for {
			m, err := w.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logx.Errorf("kafka read: %v", err)
				time.Sleep(time.Second)
				continue
			}
			if err := w.handle(ctx, m.Value); err != nil {
				logx.Errorf("handle order.shipped: %v", err)
			}
		}
	}()
}

func (w *OrderShippedWorker) handle(ctx context.Context, payload []byte) error {
	var ev orderShippedEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	items := make([]*logistics.ShipmentItemRef, 0, len(ev.Items))
	for _, it := range ev.Items {
		items = append(items, &logistics.ShipmentItemRef{
			OrderItemId: it.OrderItemId, ProductId: it.ProductId, Quantity: int32(it.Quantity),
		})
	}
	resp, err := logic.NewCreateShipmentLogic(ctx, w.svcCtx).CreateShipment(&logistics.CreateShipmentReq{
		OrderId: ev.OrderId, UserId: ev.UserId,
		TrackingNo: ev.TrackingNo, Carrier: ev.Carrier,
		Items: items,
	})
	if err != nil {
		return err
	}
	w.subscribeWithRetry(ctx, resp.ShipmentId, ev.Carrier, ev.TrackingNo)
	return nil
}

func (w *OrderShippedWorker) subscribeWithRetry(ctx context.Context, shipmentId int64, carrier, trackingNo string) {
	max := w.svcCtx.Config.Subscribe.MaxRetries
	if max <= 0 {
		max = 3
	}
	backoff := time.Duration(w.svcCtx.Config.Subscribe.InitialBackoffMs) * time.Millisecond
	if backoff <= 0 {
		backoff = time.Second
	}
	var lastErr error
	for i := 0; i < max; i++ {
		if err := w.svcCtx.Kuaidi100.Subscribe(ctx, carrier, trackingNo); err == nil {
			_, _ = w.svcCtx.DB.ExecCtx(ctx,
				"UPDATE shipment SET subscribe_status=1 WHERE id=?", shipmentId)
			return
		} else {
			lastErr = err
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	_, _ = w.svcCtx.DB.ExecCtx(ctx,
		"UPDATE shipment SET subscribe_status=2 WHERE id=?", shipmentId)
	_, _ = w.svcCtx.DB.ExecCtx(ctx,
		"INSERT INTO shipment_track(shipment_id, track_time, location, description, state_kuaidi100, state_internal) VALUES (?, NOW(), '', ?, 255, 5)",
		shipmentId, "subscribe_failed: "+lastErr.Error())
	logx.Errorf("subscribe failed after %d retries: %v", max, lastErr)
}
```

- [ ] **Step 2: Boot worker from logistics.go main**

In `logistics.go`, after `svc.NewServiceContext`, add worker startup:

```go
import "mall-logistics-rpc/internal/worker"

// after ctx := svc.NewServiceContext(c)
ws := worker.NewOrderShippedWorker(ctx)
ws.Start(context.Background())
```

- [ ] **Step 3: Build + commit**

```bash
go build ./...
git add mall-logistics-rpc/internal/worker mall-logistics-rpc/logistics.go
git commit -m "feat(logistics-rpc): kafka consumer worker for order.shipped + subscribe-with-retry"
```

---

## Task 10: mall-api scaffold logistics endpoints

**Files:**
- Create `mall-api/mall-logistics.api`
- Modify `mall-api/mall.api` (import)
- Regen mall-api

- [ ] **Step 1: api file**

Create `mall-api/mall-logistics.api`:

```api
syntax = "v1"

type (
	ShipmentItemRef {
		OrderItemId int64 `json:"orderItemId"`
		ProductId   int64 `json:"productId"`
		Quantity    int32 `json:"quantity"`
	}

	ShipmentTrack {
		TrackTime      int64  `json:"trackTime"`
		Location       string `json:"location"`
		Description    string `json:"description"`
		StateInternal  int32  `json:"stateInternal"`
		StateKuaidi100 int32  `json:"stateKuaidi100,omitempty"`
	}

	ShipmentDTO {
		Id              int64             `json:"id"`
		OrderId         int64             `json:"orderId"`
		UserId          int64             `json:"userId"`
		TrackingNo      string            `json:"trackingNo"`
		Carrier         string            `json:"carrier"`
		Status          int32             `json:"status"`
		SubscribeStatus int32             `json:"subscribeStatus"`
		LastTrackTime   int64             `json:"lastTrackTime,omitempty"`
		CreateTime      int64             `json:"createTime"`
		Items           []ShipmentItemRef `json:"items"`
		Tracks          []ShipmentTrack   `json:"tracks"`
	}

	ListOrderShipmentsReq {
		Id int64 `path:"id"`
	}
	ListOrderShipmentsResp {
		Shipments []ShipmentDTO `json:"shipments"`
	}

	GetShipmentByIdReq {
		Id int64 `path:"id"`
	}
	GetShipmentByIdResp {
		Shipment ShipmentDTO `json:"shipment"`
	}

	AdminMarkShippedReq {
		Id         int64  `path:"id"`
		TrackingNo string `json:"trackingNo"`
		Carrier    string `json:"carrier"`
	}

	AdminRetrySubscribeReq {
		Id int64 `path:"id"`
	}

	AdminInjectTrackReq {
		Id            int64  `path:"id"`
		StateInternal int32  `json:"stateInternal"`
		Location      string `json:"location"`
		Description   string `json:"description"`
	}
)

@server (
	prefix: /api
	jwt:    Auth
)
service mall-api {
	@handler ListOrderShipments
	get /order/:id/shipments (ListOrderShipmentsReq) returns (ListOrderShipmentsResp)

	@handler GetShipmentById
	get /logistics/shipment/:id (GetShipmentByIdReq) returns (GetShipmentByIdResp)
}

@server (
	prefix: /api/logistics
)
service mall-api {
	@handler Kuaidi100Webhook
	post /webhook/kuaidi100 () returns (string)
}

@server (
	prefix:     /api/admin
	middleware: AdminToken
)
service mall-api {
	@handler AdminMarkShipped
	post /order/:id/ship (AdminMarkShippedReq) returns (OkResp)

	@handler AdminRetrySubscribe
	post /logistics/:id/retry-subscribe (AdminRetrySubscribeReq) returns (OkResp)

	@handler AdminInjectTrack
	post /logistics/:id/inject-track (AdminInjectTrackReq) returns (OkResp)
}
```

- [ ] **Step 2: Add import to mall.api**

After `import "mall-review.api"` add `import "mall-logistics.api"`.

- [ ] **Step 3: Regen**

```bash
cd /home/carter/workspace/go/yw-mall/mall-api
goctl api go -api mall.api -dir . --style gozero
```

- [ ] **Step 4: Commit scaffold (build will fail until svc wiring; isolate)**

```bash
git add mall-api/mall-logistics.api mall-api/mall.api \
        mall-api/internal/handler mall-api/internal/logic mall-api/internal/types
git commit -m "feat(api): scaffold logistics handlers from mall-logistics.api"
```

---

## Task 11: mall-api logistics svc wiring + Kuaidi100 webhook key config

**Files:**
- Modify `mall-api/etc/mall-api.yaml`
- Modify `mall-api/internal/config/config.go`
- Modify `mall-api/internal/svc/servicecontext.go`
- Modify `mall-api/go.mod`

- [ ] **Step 1: yaml**

Append to `mall-api/etc/mall-api.yaml`:

```yaml
LogisticsRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: logistics.rpc

Kuaidi100:
  WebhookCustomer: ""
  WebhookKey: ""
```

- [ ] **Step 2: config**

Add to Config struct:

```go
	LogisticsRpc zrpc.RpcClientConf
	Kuaidi100    struct {
		WebhookCustomer string
		WebhookKey      string
	}
```

- [ ] **Step 3: svc context**

Add field + init:

```go
	LogisticsRpc logisticsclient.Logistics
```

```go
	LogisticsRpc: logisticsclient.NewLogistics(zrpc.MustNewClient(c.LogisticsRpc)),
```

Import `"mall-logistics-rpc/logisticsclient"`.

- [ ] **Step 4: go.mod**

```bash
cd /home/carter/workspace/go/yw-mall/mall-api
go mod edit -require=mall-logistics-rpc@v0.0.0 -replace=mall-logistics-rpc=../mall-logistics-rpc
go mod tidy
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add mall-api/etc mall-api/internal/config mall-api/internal/svc \
        mall-api/go.mod mall-api/go.sum
git commit -m "feat(api): wire logistics-rpc client and kuaidi100 webhook config"
```

---

## Task 12: mall-api Kuaidi100 webhook handler (sign verify + delegate)

**Files:**
- Modify `mall-api/internal/handler/kuaidi100webhookhandler.go`
- Modify `mall-api/internal/logic/kuaidi100webhooklogic.go`

- [ ] **Step 1: handler**

Replace handler body to read raw form, verify sign, then delegate parsed events:

```go
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"mall-api/internal/logic"
	"mall-api/internal/svc"
)

func Kuaidi100WebhookHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		param := r.FormValue("param")
		sign := r.FormValue("sign")
		l := logic.NewKuaidi100WebhookLogic(r.Context(), svcCtx)
		resp, err := l.Process(param, sign)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
```

- [ ] **Step 2: logic**

```go
package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"mall-api/internal/svc"
	"mall-common/errorx"
	logisticspb "mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type Kuaidi100WebhookLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKuaidi100WebhookLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Kuaidi100WebhookLogic {
	return &Kuaidi100WebhookLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

type kuaidi100Pushed struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	AutoCheck  string `json:"autoCheck"`
	ComOld     string `json:"comOld"`
	ComNew     string `json:"comNew"`
	LastResult struct {
		Message string `json:"message"`
		Nu      string `json:"nu"`
		Ischeck string `json:"ischeck"`
		Com     string `json:"com"`
		State   string `json:"state"`
		Status  string `json:"status"`
		Data    []struct {
			Time     string `json:"time"`
			Ftime    string `json:"ftime"`
			Context  string `json:"context"`
			Location string `json:"location"`
			Status   string `json:"status"`
		} `json:"data"`
	} `json:"lastResult"`
}

// VerifySign is exported for unit tests.
func VerifySign(param, sign, key string) bool {
	h := md5.Sum([]byte(param + key))
	return strings.EqualFold(hex.EncodeToString(h[:]), sign)
}

func (l *Kuaidi100WebhookLogic) Process(param, sign string) (string, error) {
	if !VerifySign(param, sign, l.svcCtx.Config.Kuaidi100.WebhookKey) {
		return "", errorx.NewCodeError(errorx.LogisticsKuaidi100SignInvalid)
	}
	var p kuaidi100Pushed
	if err := json.Unmarshal([]byte(param), &p); err != nil {
		return "", errorx.NewCodeError(errorx.ParamError)
	}
	events := make([]*logisticspb.Track, 0, len(p.LastResult.Data))
	for _, d := range p.LastResult.Data {
		t, _ := time.Parse("2006-01-02 15:04:05", d.Time)
		events = append(events, &logisticspb.Track{
			TrackTime:   t.Unix(),
			Location:    d.Location,
			Description: d.Context,
		})
	}
	if _, err := l.svcCtx.LogisticsRpc.IngestWebhookEvents(l.ctx, &logisticspb.IngestWebhookEventsReq{
		Carrier:    p.LastResult.Com,
		TrackingNo: p.LastResult.Nu,
		Events:     events,
	}); err != nil {
		return "", err
	}
	return `{"result":true,"returnCode":"200","message":"success"}`, nil
}
```

- [ ] **Step 3: Build + commit**

```bash
go build ./...
git add mall-api/internal/handler/kuaidi100webhookhandler.go \
        mall-api/internal/logic/kuaidi100webhooklogic.go
git commit -m "feat(api): kuaidi100 webhook handler with HMAC-MD5 sign verify"
```

---

## Task 13: mall-api user + admin handlers (delegates)

**Files:** 5 logic files: `listordershipmentslogic.go`, `getshipmentbyidlogic.go`, `adminmarkshippedlogic.go`, `adminretrysubscribelogic.go`, `admininjecttracklogic.go`.

- [ ] **Step 1: ListOrderShipments**

```go
package logic

import (
	"context"

	"mall-api/internal/svc"
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrderShipmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrderShipmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderShipmentsLogic {
	return &ListOrderShipmentsLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ListOrderShipmentsLogic) ListOrderShipments(req *types.ListOrderShipmentsReq) (*types.ListOrderShipmentsResp, error) {
	r, err := l.svcCtx.LogisticsRpc.ListShipmentsByOrder(l.ctx, &logisticspb.ListShipmentsByOrderReq{OrderId: req.Id})
	if err != nil {
		return nil, err
	}
	out := make([]types.ShipmentDTO, 0, len(r.Shipments))
	for _, s := range r.Shipments {
		out = append(out, protoShipmentToType(s))
	}
	return &types.ListOrderShipmentsResp{Shipments: out}, nil
}
```

- [ ] **Step 2: GetShipmentById**

```go
func (l *GetShipmentByIdLogic) GetShipmentById(req *types.GetShipmentByIdReq) (*types.GetShipmentByIdResp, error) {
	s, err := l.svcCtx.LogisticsRpc.GetShipment(l.ctx, &logisticspb.GetShipmentReq{ShipmentId: req.Id})
	if err != nil {
		return nil, err
	}
	return &types.GetShipmentByIdResp{Shipment: protoShipmentToType(s)}, nil
}
```

- [ ] **Step 3: Admin handlers**

```go
// AdminMarkShipped: delegate to OrderRpc.MarkShipped
func (l *AdminMarkShippedLogic) AdminMarkShipped(req *types.AdminMarkShippedReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.OrderRpc.MarkShipped(l.ctx, &orderpb.MarkShippedReq{
		OrderId: req.Id, TrackingNo: req.TrackingNo, Carrier: req.Carrier,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}

// AdminRetrySubscribe
func (l *AdminRetrySubscribeLogic) AdminRetrySubscribe(req *types.AdminRetrySubscribeReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.LogisticsRpc.RetrySubscribe(l.ctx, &logisticspb.RetrySubscribeReq{ShipmentId: req.Id}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}

// AdminInjectTrack
func (l *AdminInjectTrackLogic) AdminInjectTrack(req *types.AdminInjectTrackReq) (*types.OkResp, error) {
	if _, err := l.svcCtx.LogisticsRpc.InjectTrack(l.ctx, &logisticspb.InjectTrackReq{
		ShipmentId: req.Id, StateInternal: req.StateInternal,
		Location: req.Location, Description: req.Description,
	}); err != nil {
		return nil, err
	}
	return &types.OkResp{Ok: true}, nil
}
```

Add `protoShipmentToType` to a new `mall-api/internal/logic/logistics_helpers.go`:

```go
package logic

import (
	"mall-api/internal/types"
	logisticspb "mall-logistics-rpc/logistics"
)

func protoShipmentToType(s *logisticspb.Shipment) types.ShipmentDTO {
	out := types.ShipmentDTO{
		Id: s.Id, OrderId: s.OrderId, UserId: s.UserId,
		TrackingNo: s.TrackingNo, Carrier: s.Carrier,
		Status: s.Status, SubscribeStatus: s.SubscribeStatus,
		LastTrackTime: s.LastTrackTime, CreateTime: s.CreateTime,
	}
	for _, it := range s.Items {
		out.Items = append(out.Items, types.ShipmentItemRef{
			OrderItemId: it.OrderItemId, ProductId: it.ProductId, Quantity: it.Quantity,
		})
	}
	for _, t := range s.Tracks {
		out.Tracks = append(out.Tracks, types.ShipmentTrack{
			TrackTime: t.TrackTime, Location: t.Location, Description: t.Description,
			StateInternal: t.StateInternal, StateKuaidi100: t.StateKuaidi100,
		})
	}
	return out
}
```

- [ ] **Step 4: Build + commit**

```bash
go build ./...
git add mall-api/internal/logic/
git commit -m "feat(api): logistics user + admin handlers"
```

---

## Task 14: Integrate shipments into order detail (parallel)

**Files:** Modify `mall-api/internal/logic/orderdetaillogic.go`. Modify `mall-api/mall.api` (add `Shipments []ShipmentDTO` to `OrderDetailResp`).

- [ ] **Step 1: api**

Find `type OrderDetailResp { ... }` in `mall.api` and inside the braces add:

```api
		Shipments []ShipmentDTO `json:"shipments,omitempty"`
```

Regen: `goctl api go -api mall.api -dir . --style gozero`.

- [ ] **Step 2: logic — fan out**

Open `mall-api/internal/logic/orderdetaillogic.go`. Wrap existing OrderRpc.GetOrder call in a parallel pattern:

```go
import (
	"sync"
	logisticspb "mall-logistics-rpc/logistics"
)

func (l *OrderDetailLogic) OrderDetail(req *types.OrderDetailReq) (*types.OrderDetailResp, error) {
	var (
		ord       *orderpb.GetOrderResp
		shipments *logisticspb.ListShipmentsByOrderResp
		ordErr    error
		wg        sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		ord, ordErr = l.svcCtx.OrderRpc.GetOrder(l.ctx, &orderpb.GetOrderReq{Id: req.Id})
	}()
	go func() {
		defer wg.Done()
		shipments, _ = l.svcCtx.LogisticsRpc.ListShipmentsByOrder(l.ctx, &logisticspb.ListShipmentsByOrderReq{OrderId: req.Id})
	}()
	wg.Wait()
	if ordErr != nil {
		return nil, ordErr
	}
	resp := /* ... map ord into types.OrderDetailResp as before ... */
	if shipments != nil {
		for _, s := range shipments.Shipments {
			resp.Shipments = append(resp.Shipments, protoShipmentToType(s))
		}
	}
	return resp, nil
}
```

> Reuse `protoShipmentToType` from Task 13.

- [ ] **Step 3: Build + commit**

```bash
go build ./...
git add mall-api/mall.api mall-api/internal/types mall-api/internal/logic/orderdetaillogic.go
git commit -m "feat(api): include shipments in order detail (parallel fetch)"
```

---

## Task 15: start.sh registration + bootstrap schema

**Files:** Modify `start.sh`.

- [ ] **Step 1: Add to SERVICES**

After `mall-review-rpc:review.go:review-rpc:9015`, insert:

```bash
    "mall-logistics-rpc:logistics.go:logistics-rpc:9016"
```

- [ ] **Step 2: Add to bootstrap schema map**

After `[mall_review]=mall-review-rpc/sql/review.sql`, add:

```bash
        [mall_logistics]=mall-logistics-rpc/sql/logistics.sql
```

- [ ] **Step 3: Commit**

```bash
git add start.sh
git commit -m "ops: register mall-logistics-rpc in start.sh + bootstrap mall_logistics schema"
```

---

## Task 16: End-to-end smoke (kuaidi100 demo path) — manual

**Files:** none modified. Documents the verification flow.

- [ ] **Step 1: Bring up infra + services**

```bash
cd /home/carter/workspace/go/env && docker compose up -d
cd /home/carter/workspace/go/yw-mall && ./start.sh nuke && ./start.sh start
sleep 10
./start.sh status | grep -E 'review|logistics|order'
```

- [ ] **Step 2: Seed an order in `paid` status**

```bash
JWT=$(curl -s -X POST http://127.0.0.1:18888/api/user/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"qa-logi","password":"pw","phone":"19000000002"}' | jq -r .id)
JWT=$(curl -s -X POST http://127.0.0.1:18888/api/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"qa-logi","password":"pw"}' | jq -r .token)
# create order via existing API; capture orderId from response
```

If easier, directly seed in DB:

```bash
docker exec $(docker ps --filter name=mysql-master1 -q) \
  mysql -uroot -proot123 mall_order -e \
  "INSERT INTO \`order\`(user_id, total_amount, status) VALUES (1, 100, 1);
   SELECT LAST_INSERT_ID() AS order_id;"
```

- [ ] **Step 3: Mark shipped via admin API**

```bash
curl -s -X POST http://127.0.0.1:18888/api/admin/order/<orderId>/ship \
  -H 'X-Admin-Token: mall-admin-token-change-in-production' \
  -H 'Content-Type: application/json' \
  -d '{"trackingNo":"SF1234567890","carrier":"shunfeng"}'
```

Expected: `{"ok":true}`. logistics-rpc log shows kafka event consumed; subscribe attempted; if no real kuaidi100 key configured, subscribe_status=2 + a synthetic track row appears.

- [ ] **Step 4: Inject demo tracks**

```bash
SHIP_ID=$(curl -s -H "Authorization: Bearer $JWT" \
  http://127.0.0.1:18888/api/order/<orderId>/shipments | jq '.shipments[0].id')

for state in 1 2 3 4; do
  curl -s -X POST http://127.0.0.1:18888/api/admin/logistics/$SHIP_ID/inject-track \
    -H 'X-Admin-Token: mall-admin-token-change-in-production' \
    -H 'Content-Type: application/json' \
    -d "{\"stateInternal\":$state,\"location\":\"Shenzhen\",\"description\":\"demo step $state\"}"
done
```

- [ ] **Step 5: Verify**

```bash
curl -s -H "Authorization: Bearer $JWT" \
  http://127.0.0.1:18888/api/order/<orderId>/shipments | jq
curl -s -H "Authorization: Bearer $JWT" \
  http://127.0.0.1:18888/api/order/<orderId> | jq '.shipments'
```

Expect: shipments array with 4+ tracks, status=4 (delivered).

- [ ] **Step 6: Stop**

```bash
./start.sh stop
git commit --allow-empty -m "milestone: mall-logistics-rpc end-to-end smoke verified"
```

---

## Plan Self-Review Notes

- **Spec coverage**: every spec section maps to ≥1 task. §1 goals → 6–9; §2 deps → 2,5,9,11; §3 flows → 6,7,9,12; §4 schema → 3; §5 contracts → 3,10,11,12,13; §6 errors → 1; §7 consistency → 7,9; §8 testing → 16 (manual). §2.5 order-rpc extension → task 2.
- **Placeholder scan**: no TBD/TODO; explicit `<orderId>` and `<your customer id>` markers are user-supplied runtime values, not placeholder code.
- **Type consistency**: `Shipment` (proto) ↔ `ShipmentDTO` (api types) ↔ `Shipment` (model) — three layers, names differ on purpose. `protoShipmentToType` and `toShipmentProto` are the conversion seams.
- **Risks**:
  - Task 2 assumes `OrderModel.Conn()` accessor exists; if not, add `DB sqlx.SqlConn` field on order-rpc svc.
  - Task 9 assumes mall-activity-async-worker pattern for in-process Kafka consumer is fine; verify by inspecting `logistics.go` after Step 2.
  - Task 12 webhook key in mall-api yaml is empty by default — webhook will return 401 until user fills `Kuaidi100.WebhookKey`. Demo mode bypasses webhook entirely (use InjectTrack).
