# mall-logistics-rpc 设计稿

- **状态**：Draft（待 user 最终 review）
- **日期**：2026-05-01
- **作者**：Carter（与 Claude 协作 brainstorming 产出）
- **上游决策**：本服务是 yw-mall "C 端前端 + 评价 + 物流" 大需求被拆解后的 4 个子项目之一。mall-review-rpc 已完工。本稿覆盖 mall-logistics-rpc。

---

## 1. 目标与不目标

**目标**：在 yw-mall 微服务体系下新增 `mall-logistics-rpc`，承担"运单创建 / 轨迹查询 / 状态汇总"，对接快递100（聚合承运商 API），通过订阅 + webhook 推送获取轨迹（不轮询）。

具体能力：

- 1 个订单 → N 个 shipment，1 个 shipment → N 个 order_item（多包模式）
- 内部 7 态状态机（待揽收/已揽收/运输中/派送中/已签收/异常/退回）
- 接收 Kafka `order.shipped` 事件自动创建 shipment 并向快递100 订阅
- 快递100 webhook 接收（HMAC-MD5 签名校验）写入轨迹
- 订阅失败 3 次重试 + 商家手动重试入口
- C 端按订单查所有运单 + 轨迹时间线
- mall-api 商品详情和订单详情都不再多打 logistics 调用；只有"订单详情聚合"要并行调
- demo 模式：admin RPC 注入轨迹（绕过 webhook，方便无公网时验证全链路）

**显式不目标**：

- 真承运商 SDK 接入（顺丰/京东/中通各自 API）— 全部走快递100 聚合
- 同步轮询模式（不实现 `pollNow`）
- 真实"承运商" 选择 / 路由（直接拿快递100 返回的 carrier code 用）
- 国际物流 / 海关
- 物流时效预测 / 异常智能识别
- 商家后台 UI（仍仅留 admin token 临时入口）

---

## 2. 服务边界与依赖

### 2.1 新服务

`mall-logistics-rpc`：gRPC 服务，端口 `9016`，etcd key `logistics.rpc`。

### 2.2 出向依赖

| 依赖 | 调用 | 用途 |
|------|------|------|
| 快递100 | HTTPS POST `https://poll.kuaidi100.com/poll/3.0` | 订阅运单（subscribe） |
| 快递100 webhook | （反向）快递100 → 我们的 `/api/logistics/webhook/kuaidi100` | 推送轨迹 |
| MySQL（schema `mall_logistics`） | 直连 ProxySQL 6033 | 主存储 |
| Kafka | 消费 topic `order.shipped` | 触发自动创建 |
| Redis | 暂不需要（轨迹查询走 DB；如果性能不够再加） |

### 2.3 入向依赖

| 调用方 | 用途 |
|--------|------|
| `mall-api` | 用户查物流（订单维度）；admin 注入轨迹（demo 兜底） |
| `mall-api` | 订单详情聚合时并行调 `ListShipmentsByOrder` |

### 2.4 部署变更

- 新建 RPC 服务 + worker（Kafka 消费）。worker 跟 RPC 同进程（参考 mall-activity-rpc 已有 in-process kafka consumer 模式），或独立 worker（参考 mall-activity-async-worker）。**默认在 logistics-rpc 进程内消费**，简单。
- start.sh SERVICES 数组追加 `"mall-logistics-rpc:logistics.go:logistics-rpc:9016"`
- compose.yml 不动
- 新建 schema `mall_logistics`，bootstrap 阶段建库

### 2.5 关联：mall-order-rpc 必须扩展

为了让 `order.shipped` 事件流通，order-rpc 需要补：

- 新 RPC：`MarkShipped(orderId, trackingNo, carrierCode) → ()`（管理员/系统调用，更新 order.status 并产 Kafka event）
- order.status 枚举确认：当前需要查实际 schema，**目前已知值**：`0 pending, 1 paid, 2 shipped(?), 3 completed, 4 cancelled`。如果 status=2 不存在则需补。
- Kafka producer 发 `order.shipped` topic：`{orderId, userId, items: [{orderItemId, productId, quantity}], shippedAt}`

> 这部分变更属于 mall-order-rpc 扩展，单独一个 task，是 logistics-rpc 实现的前置依赖。

---

## 3. 数据流

### 3.1 自动创建 shipment（订单已发货）

```
admin → mall-order-rpc.MarkShipped(orderId, trackingNo, carrierCode)
         ├─ DB: UPDATE order SET status=2, tracking_no=?, carrier=?
         ├─ Kafka: produce "order.shipped" {orderId, userId, items[], shippedAt}
         └─ return ok

(消费端)
mall-logistics-rpc Kafka consumer
  ├─ on event:
  ├─ DB tx:
  │   ├─ INSERT shipment (order_id, user_id, tracking_no, carrier, status=0)
  │   └─ INSERT shipment_item × N
  ├─ POST kuaidi100/poll/3.0 subscribe (param + sign)
  │   ├─ 成功: status 仍 0（等推送）
  │   └─ 失败: 重试 3 次（指数 1s/2s/4s），仍失败则
  │     ├─ 写一条 track {state=255, desc="subscribe_failed: <error>"}
  │     └─ logx.Error 告警
  └─ commit
```

### 3.2 接收快递100 webhook（轨迹推送）

```
快递100 → POST /api/logistics/webhook/kuaidi100
                  body: param=<json>&sign=<MD5>

mall-api webhook handler
  ├─ 校验 sign = MD5(param + key)，不匹配 → 401
  ├─ 解析 param.lastResult.data[] (轨迹列表)
  ├─ 调 logistics-rpc.IngestWebhookEvents(trackingNo, events[])
  │     ├─ 按 trackingNo 找 shipment
  │     ├─ DB tx:
  │     │   ├─ INSERT shipment_track × N (去重：相同 time+desc 跳过)
  │     │   └─ UPDATE shipment SET status=<map_kuaidi100_state>
  │     └─ return ok
  └─ 返回 "result=true,returnCode=200" (快递100 期望的格式)
```

### 3.3 用户查询物流（订单详情）

```
client → GET /api/order/:id (JWT)
       └→ mall-api 并行调:
            ├─ order-rpc.GetOrder(orderId) (含 items)
            └─ logistics-rpc.ListShipmentsByOrder(orderId)
       └─ 聚合返回 order JSON 含 shipments: [{trackingNo, carrier, status, tracks: [...]}]

client → GET /api/order/:id/shipments (JWT)
       └→ logistics-rpc.ListShipmentsByOrder(orderId)
            └─ DB: SELECT shipment + LEFT JOIN shipment_track ORDER BY time DESC
```

### 3.4 demo 注入轨迹（无公网时）

```
admin → POST /api/admin/logistics/:shipmentId/inject-track
         (Header X-Admin-Token, body: state, location, description)
       └→ mall-api → logistics-rpc.InjectTrack(shipmentId, state, location, desc)
            ├─ INSERT shipment_track
            ├─ UPDATE shipment SET status = mapStateToInternal(state)
            └─ return ok
```

---

## 4. 数据模型

### 4.1 DDL：`mall_logistics.shipment`

```sql
CREATE TABLE `shipment` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `tracking_no` VARCHAR(64) NOT NULL,
  `carrier` VARCHAR(32) NOT NULL COMMENT 'kuaidi100 carrier code (sf/jd/zto/...)',
  `status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=created, 1=collected, 2=in_transit, 3=delivering, 4=delivered, 5=exception, 6=returned',
  `subscribe_status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=pending, 1=ok, 2=failed',
  `last_track_time` DATETIME NULL,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_carrier_tracking` (`carrier`, `tracking_no`),
  KEY `idx_order` (`order_id`),
  KEY `idx_user_time` (`user_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.2 DDL：`mall_logistics.shipment_item`

```sql
CREATE TABLE `shipment_item` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `shipment_id` BIGINT NOT NULL,
  `order_item_id` BIGINT NOT NULL,
  `product_id` BIGINT NOT NULL,
  `quantity` INT NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_shipment` (`shipment_id`),
  KEY `idx_order_item` (`order_item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.3 DDL：`mall_logistics.shipment_track`

```sql
CREATE TABLE `shipment_track` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `shipment_id` BIGINT NOT NULL,
  `track_time` DATETIME NOT NULL,
  `location` VARCHAR(255) NULL,
  `description` VARCHAR(500) NOT NULL,
  `state_kuaidi100` SMALLINT NULL COMMENT 'raw kuaidi100 state code (0..14, 255 for synthetic)',
  `state_internal` TINYINT NOT NULL COMMENT 'mapped internal status 0..6',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_shipment_time_desc` (`shipment_id`, `track_time`, `description`(50)),
  KEY `idx_shipment_time` (`shipment_id`, `track_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

> `uk_shipment_time_desc` 是去重保护：webhook 重发同一条轨迹不会重复入库。`description(50)` 前缀索引避免索引超限。

### 4.4 状态映射

快递100 状态码 → 内部 status：

| kuaidi100 | 描述 | 内部 status |
|---|---|---|
| 0 | 在途 | 2 (in_transit) |
| 1 | 揽收 | 1 (collected) |
| 2 | 疑难 | 5 (exception) |
| 3 | 签收 | 4 (delivered) |
| 4 | 退签 | 6 (returned) |
| 5 | 派件 | 3 (delivering) |
| 6 | 退回 | 6 (returned) |
| 14 | 拒签 | 6 (returned) |
| 255 | 系统合成（subscribe_failed） | 5 (exception) |

写入 `shipment.status` 时只升档不降档（避免乱序推送把"已签收"覆盖回"在途"）：用 `UPDATE shipment SET status=GREATEST(status, ?) WHERE id=?`。

---

## 5. 接口契约

### 5.1 proto（`mall-common/proto/logistics/logistics.proto`）

```proto
syntax = "proto3";
package logistics;
option go_package = "./logistics";

service Logistics {
  rpc CreateShipment(CreateShipmentReq) returns (CreateShipmentResp);
  rpc ListShipmentsByOrder(ListShipmentsByOrderReq) returns (ListShipmentsByOrderResp);
  rpc GetShipment(GetShipmentReq) returns (Shipment);
  rpc IngestWebhookEvents(IngestWebhookEventsReq) returns (Empty);
  rpc RetrySubscribe(RetrySubscribeReq) returns (Empty);
  rpc InjectTrack(InjectTrackReq) returns (Empty);
}

message Empty {}

message CreateShipmentReq {
  int64 order_id = 1;
  int64 user_id = 2;
  string tracking_no = 3;
  string carrier = 4;
  repeated ShipmentItemRef items = 5;
}
message CreateShipmentResp { int64 shipment_id = 1; }

message ShipmentItemRef {
  int64 order_item_id = 1;
  int64 product_id = 2;
  int32 quantity = 3;
}

message Track {
  int64 track_time = 1;
  string location = 2;
  string description = 3;
  int32 state_internal = 4;
  int32 state_kuaidi100 = 5;
}

message Shipment {
  int64 id = 1;
  int64 order_id = 2;
  int64 user_id = 3;
  string tracking_no = 4;
  string carrier = 5;
  int32 status = 6;
  int32 subscribe_status = 7;
  int64 last_track_time = 8;
  int64 create_time = 9;
  repeated ShipmentItemRef items = 10;
  repeated Track tracks = 11;
}

message ListShipmentsByOrderReq { int64 order_id = 1; }
message ListShipmentsByOrderResp { repeated Shipment shipments = 1; }

message GetShipmentReq { int64 shipment_id = 1; }

message IngestWebhookEventsReq {
  string carrier = 1;
  string tracking_no = 2;
  repeated Track events = 3;
}

message RetrySubscribeReq { int64 shipment_id = 1; }

message InjectTrackReq {
  int64 shipment_id = 1;
  int32 state_internal = 2;
  string location = 3;
  string description = 4;
}
```

### 5.2 mall-api 路由

| Method | Path | 鉴权 | 说明 |
|---|---|---|---|
| GET | `/api/order/:id/shipments` | JWT | 列出该订单所有 shipment + 轨迹 |
| GET | `/api/logistics/shipment/:id` | JWT | 单个 shipment 详情 |
| POST | `/api/logistics/webhook/kuaidi100` | sign校验 | 快递100 推送入口（公开） |
| POST | `/api/admin/order/:id/ship` | `X-Admin-Token` | 调 order-rpc.MarkShipped（demo 兜底） |
| POST | `/api/admin/logistics/:id/retry-subscribe` | `X-Admin-Token` | 手动重试订阅 |
| POST | `/api/admin/logistics/:id/inject-track` | `X-Admin-Token` | demo 注入轨迹 |

商品详情聚合**不**调 logistics（评分汇总走 review-rpc，物流跟商品无关）。订单详情聚合**改**：mall-api 在 GET `/api/order/:id` 里并行调 `ListShipmentsByOrder` + 既有 `GetOrder`，合并到响应。

### 5.3 mall-api yaml 新增

```yaml
LogisticsRpc:
  Etcd:
    Hosts: [127.0.0.1:2379]
    Key: logistics.rpc

Kuaidi100:
  WebhookCustomer: "<your customer id>"
  WebhookKey: "<your secret key>"
```

> mall-api 持有 webhook 校验所需的 key（不让 logistics-rpc 也持一份）。webhook 校验通过后再 RPC 调 logistics-rpc.IngestWebhookEvents（已被 mall-api 信任）。

### 5.4 mall-logistics-rpc yaml

```yaml
Name: logistics.rpc
ListenOn: 0.0.0.0:9016
Etcd:
  Hosts: [127.0.0.1:2379]
  Key: logistics.rpc
DataSource: proxysql:proxysql123@tcp(127.0.0.1:6033)/mall_logistics?charset=utf8mb4&parseTime=true&loc=Local
Cache:
  - Host: 127.0.0.1:6379
RedisCache:
  Host: 127.0.0.1:6379
  Type: node

Kafka:
  Brokers: [127.0.0.1:19092, 127.0.0.1:19093, 127.0.0.1:19094]
  Topic: order.shipped
  Group: logistics-rpc

Kuaidi100:
  Customer: "<your customer id>"
  Key: "<your secret key>"
  PollEndpoint: "https://poll.kuaidi100.com/poll/3.0"
  WebhookCallback: "http://<your-public-host>/api/logistics/webhook/kuaidi100"

Subscribe:
  MaxRetries: 3
  InitialBackoffMs: 1000
```

---

## 6. 校验与错误码

### 6.1 错误码（`mall-common/errorx`，9xxx 段）

| code | const | message |
|---|---|---|
| 9001 | `LogisticsShipmentNotFound` | shipment not found |
| 9002 | `LogisticsTrackingNoExists` | tracking number already exists for this carrier |
| 9003 | `LogisticsKuaidi100SignInvalid` | invalid kuaidi100 webhook signature |
| 9004 | `LogisticsSubscribeFailed` | subscribe to kuaidi100 failed after retries |
| 9005 | `LogisticsOrderNotShippable` | order is not in a state that can be shipped |
| 9006 | `LogisticsCarrierUnknown` | unknown carrier code |
| 9007 | `LogisticsTrackingInvalid` | invalid tracking number format |

### 6.2 校验

| 入口 | 校验 |
|---|---|
| CreateShipment | order 存在 + 不存在同 carrier+tracking_no（命中 UK 直接 ErrTrackingNoExists） |
| MarkShipped (order-rpc) | order.status ∈ {paid}；非法状态返回 LogisticsOrderNotShippable |
| webhook | mall-api 层校验 sign；通过后 logistics-rpc 信任入参 |
| InjectTrack | mall-api 层 X-Admin-Token；shipment 存在 |
| RetrySubscribe | mall-api 层 X-Admin-Token；shipment 存在且 subscribe_status != 1 |

---

## 7. 一致性 / 幂等

- **订阅幂等**：相同 trackingNo 重复订阅快递100 会返回成功（其文档允许），DB UK 兜底。
- **轨迹去重**：`uk_shipment_time_desc` 索引保证同一条事件 webhook 重发不会插重复。
- **状态升档**：`UPDATE shipment SET status=GREATEST(status, ?)` 防乱序覆盖。
- **Kafka 消费幂等**：用同 carrier+tracking_no UK 兜底；消费失败 throw 让 Kafka 重投。

---

## 8. 测试策略

**logistics-rpc 单元测试**：
- 状态机映射 + GREATEST 升档逻辑：纯函数测试
- 订阅重试退避：用 fake HTTP server
- IngestWebhookEvents 去重：sqlmock

**mall-api webhook 测试**：
- 签名校验通过/失败两条路径
- payload 解析

**手工 QA 清单**（需要本地基础设施）：
1. 启动全栈
2. 下单 → 支付完成 → admin 调 `POST /api/admin/order/:id/ship` 给 trackingNo（demo: SF1234567890）
3. 看 logistics-rpc log 确认收到 Kafka event 并尝试订阅快递100（如果没配真 key，订阅会失败 → subscribe_status=2，写一条 synthetic track）
4. demo 模式：admin 调 `POST /api/admin/logistics/:id/inject-track` 注入 "已揽收/运输中/已签收" 轨迹
5. 用户查 `GET /api/order/:id` 看到 shipments 字段
6. 用户查 `GET /api/order/:id/shipments` 看到完整轨迹时间线

---

## 9. 子项目落地顺序

- 本 spec：`mall-logistics-rpc` + mall-order-rpc 扩展 + mall-api 路由（订单详情聚合 + webhook + admin 接口）
- 下一个：C 端前端（终于到了）

---

## 10. 待办

- [ ] User 最终 review 本 spec
- [ ] 确认 mall-order-rpc 当前 `order.status` 值映射（特别是 status=2=shipped 是否已存在）
- [ ] 进入 writing-plans 流程
