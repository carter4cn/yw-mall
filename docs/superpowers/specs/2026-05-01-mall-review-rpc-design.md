# mall-review-rpc 设计稿

- **状态**：Draft（待 user 最终 review）
- **日期**：2026-05-01
- **作者**：Carter（与 Claude 协作 brainstorming 产出）
- **上游决策**：本服务是 yw-mall "C 端前端 + 评价 + 物流" 大需求被拆解后的 4 个子项目之一，本稿仅覆盖 mall-review-rpc。其余 3 个子项目（mall-logistics-rpc、mall-api 接入、C 端前端）后续各自走独立 spec。

---

## 1. 目标与不目标

**目标**：在 yw-mall 微服务体系下新增 `mall-review-rpc` 服务，承担商品评价的写入、查询、聚合，并支持：

- 多维评分（描述相符 / 物流速度 / 服务态度，1–5 星）+ 自动算总分
- 文字 + 图片（≤ 9 张，单张 ≤ 5MB）+ 短视频（≤ 1 个，≤ 50MB）
- 追评（原评不可改，发出后 ≥ 7 天可追评一次）
- 商家回复（数据模型完整 + 临时 admin-token 写入路径，正式商家后台后续阶段做）
- 评分聚合（avg + count + 1–5 星分布 + 带媒体数）+ Redis 缓存
- 提交评价时调 mall-risk-rpc 做行为级风控
- 一次购买只能发一条主评（DB 唯一索引保证）
- 用户不可删；管理员（admin token）可软删

**显式不目标（本期不做）**：

- 文本内容审核 / 关键词过滤
- 评价点赞 / 有用度
- 评价分享外链
- 商家后台 UI（仅留 admin token 临时入口）
- 商家身份模型（`merchant_user_id` 字段先允许任意 user_id 写入，等"商家后台"阶段再补 merchant 实体表）

---

## 2. 服务边界与依赖

### 2.1 新服务

`mall-review-rpc`：gRPC 服务，端口 `9015`，etcd key `review.rpc`。命名、目录结构、配置形态全部与现有 RPC 服务（如 `mall-reward-rpc`）保持一致。

### 2.2 出向依赖

| 依赖 | 调用方法 | 用途 |
|------|---------|------|
| `mall-order-rpc` | `GetOrderItem(orderItemId)` ← **当前 proto 未提供，需新增** | 校验"调用者买过该商品 + 订单已完成" |
| `mall-risk-rpc` | `CheckBlacklist(subject_type="user", subject_value=userId)` + `RateLimit(scope="submit_review", subject=userId)` | 行为级风控（黑名单 + 频次限制） |
| MySQL（schema `mall_review`） | 直连 ProxySQL（`proxysql:proxysql123@tcp(127.0.0.1:6033)`） | 主存储 |
| Redis | 直连本地 Redis | 评分聚合缓存 |

> **不直接接 MinIO**。媒体上传由 `mall-api` 的 `/api/upload/review-media` 接口持有 MinIO 凭据并代写，review-rpc 只接收并存储 URL 字符串。

### 2.3 入向依赖

| 调用方 | 用途 |
|--------|------|
| `mall-api` | C 端用户写评价 / 看评价 / 看汇总 / 我的评价；管理员回复 / 软删 |
| `mall-product-rpc` | **不直接调用**；商品详情聚合时由 mall-api 并行调 review-rpc 的 `GetProductRatingSummary`，避免商品服务被评价拖慢 |

### 2.4 部署变更

- `start.sh` 的 `SERVICES` 数组追加：`"mall-review-rpc:review.go:review-rpc:9015"`
- `compose.yml` **不动**：review-rpc 复用现有 MySQL / Redis / etcd 实例
- 新增数据库 schema `mall_review`（DDL 见 §4），需要 bootstrap 阶段建库

---

## 3. 数据流

### 3.1 提交评价（主评）

```
client
  └→ POST /api/review/submit  (JWT, body: order_item_id, scores, content, media[])
       └→ mall-api validate JWT, decode user_id
            └→ review-rpc.SubmitReview
                 ├─ order-rpc.GetOrderItem(order_item_id)         // 校验归属 + 订单状态=完成
                 ├─ risk-rpc.CheckBlacklist(user, userId)         // 黑名单
                 ├─ risk-rpc.RateLimit("submit_review", userId)   // 频次（默认每用户每小时 ≤ 10 条）
                 ├─ DB: INSERT review + INSERT review_media (tx)
                 ├─ Redis: DEL mall:review:summary:{product_id}
                 └─ return review_id
```

### 3.2 查询商品评价列表 + 聚合（商品详情页）

```
client
  └→ GET /api/product/:id (JWT/公开均可)
       └→ mall-api 并行调:
            ├─ product-rpc.GetProduct(id)
            └─ review-rpc.GetProductRatingSummary(id)   // Redis 命中直接返回，否则回源
       └─ 聚合返回（商品详情包含 rating_summary 字段）

client
  └→ GET /api/product/:id/reviews?sort=time&score=5&withMedia=true&page=1&pageSize=20
       └→ review-rpc.ListProductReviews(...)
```

### 3.3 追评

```
client
  └→ POST /api/review/followup  (JWT, body: review_id, content, media[])
       └→ review-rpc.SubmitFollowup
            ├─ DB: SELECT review WHERE id=? AND user_id=?
            ├─ assert: followup_content IS NULL && create_time <= NOW()-INTERVAL 7 DAY
            ├─ DB: UPDATE review SET followup_content=?, followup_time=NOW(),
            │           [+ INSERT review_media WHERE is_followup=1]
            ├─ Redis: DEL mall:review:summary:{product_id}
            └─ return ok
```

### 3.4 商家回复（demo 临时通道）

```
admin
  └→ POST /api/admin/review/:id/reply  (Header: X-Admin-Token: <static>, body: merchant_user_id, text)
       └→ mall-api validate AdminToken == config.AdminToken
            └→ review-rpc.ReplyReview
                 ├─ UPDATE review SET merchant_reply_text=?, merchant_reply_time=NOW(), merchant_user_id=?
                 ├─ Redis: 不变（聚合不含商家回复）
                 └─ return ok
```

### 3.5 软删

```
admin
  └→ DELETE /api/admin/review/:id  (Header: X-Admin-Token)
       └→ review-rpc.SoftDeleteReview
            ├─ UPDATE review SET status=1
            ├─ Redis: DEL mall:review:summary:{product_id}
            └─ return ok
```

---

## 4. 数据模型

### 4.1 DDL：`mall_review.review`

```sql
CREATE TABLE `review` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `order_item_id` BIGINT NOT NULL,
  `user_id` BIGINT NOT NULL,
  `product_id` BIGINT NOT NULL,
  `score_overall` TINYINT NOT NULL,
  `score_match` TINYINT NOT NULL,
  `score_logistics` TINYINT NOT NULL,
  `score_service` TINYINT NOT NULL,
  `content` VARCHAR(2000) NOT NULL,
  `has_media` TINYINT NOT NULL DEFAULT 0,
  `followup_content` VARCHAR(500) NULL,
  `followup_time` DATETIME NULL,
  `merchant_reply_text` VARCHAR(500) NULL,
  `merchant_reply_time` DATETIME NULL,
  `merchant_user_id` BIGINT NULL,
  `status` TINYINT NOT NULL DEFAULT 0
    COMMENT '0=normal, 1=admin_soft_deleted, 2=admin_hidden',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_item` (`order_item_id`),
  KEY `idx_product_status_time` (`product_id`, `status`, `create_time` DESC),
  KEY `idx_product_score` (`product_id`, `score_overall`, `status`),
  KEY `idx_user_time` (`user_id`, `create_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.2 DDL：`mall_review.review_media`

```sql
CREATE TABLE `review_media` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `review_id` BIGINT NOT NULL,
  `media_type` TINYINT NOT NULL COMMENT '1=image, 2=video',
  `media_url` VARCHAR(500) NOT NULL,
  `sort` TINYINT NOT NULL DEFAULT 0,
  `is_followup` TINYINT NOT NULL DEFAULT 0,
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_review` (`review_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 4.3 评分写入约定

`score_overall` 在写入时由服务层算好：

```
score_overall = round((score_match + score_logistics + score_service) / 3.0)
```

读路径不再做聚合。

### 4.4 Redis Key

| key | type | TTL | 失效时机 |
|-----|------|-----|---------|
| `mall:review:summary:{productId}` | string (JSON) | 300s | Submit / Followup / SoftDelete 后 DEL |

JSON 结构：
```json
{
  "avg": 4.6,
  "count": 128,
  "distribution": {"1": 2, "2": 3, "3": 8, "4": 25, "5": 90},
  "withMediaCount": 47
}
```

---

## 5. 接口契约

### 5.1 proto（`mall-common/proto/review/review.proto`）

```proto
syntax = "proto3";
package review;
option go_package = "./review";

service Review {
  rpc SubmitReview(SubmitReviewReq) returns (SubmitReviewResp);
  rpc SubmitFollowup(SubmitFollowupReq) returns (Empty);
  rpc GetReview(GetReviewReq) returns (Review);
  rpc ListProductReviews(ListProductReviewsReq) returns (ListProductReviewsResp);
  rpc ListUserReviews(ListUserReviewsReq) returns (ListProductReviewsResp);
  rpc GetProductRatingSummary(GetProductRatingSummaryReq) returns (RatingSummary);
  rpc ReplyReview(ReplyReviewReq) returns (Empty);
  rpc SoftDeleteReview(SoftDeleteReviewReq) returns (Empty);
}

message Empty {}

message Media {
  int32 type = 1;            // 1=image, 2=video
  string url = 2;
  int32 sort = 3;
}

message SubmitReviewReq {
  int64 order_item_id = 1;
  int64 user_id = 2;
  int32 score_match = 3;
  int32 score_logistics = 4;
  int32 score_service = 5;
  string content = 6;
  repeated Media media = 7;
}
message SubmitReviewResp {
  int64 review_id = 1;
}

message SubmitFollowupReq {
  int64 review_id = 1;
  int64 user_id = 2;
  string content = 3;
  repeated Media media = 4;
}

message Review {
  int64 id = 1;
  int64 order_item_id = 2;
  int64 user_id = 3;
  int64 product_id = 4;
  int32 score_overall = 5;
  int32 score_match = 6;
  int32 score_logistics = 7;
  int32 score_service = 8;
  string content = 9;
  repeated Media media = 10;
  string followup_content = 11;
  int64 followup_time = 12;
  repeated Media followup_media = 13;
  string merchant_reply_text = 14;
  int64 merchant_reply_time = 15;
  int64 merchant_user_id = 16;
  int32 status = 17;
  int64 create_time = 18;
}

message GetReviewReq { int64 review_id = 1; }

message ListProductReviewsReq {
  int64 product_id = 1;
  string sort = 2;          // "time" | "score" | "hasMedia"
  int32 score = 3;          // 0=不过滤；1..5=只看该星
  bool with_media = 4;
  int32 page = 5;
  int32 page_size = 6;
}
message ListProductReviewsResp {
  repeated Review reviews = 1;
  int64 total = 2;
}

message ListUserReviewsReq {
  int64 user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message GetProductRatingSummaryReq { int64 product_id = 1; }
message RatingSummary {
  double avg = 1;
  int64 count = 2;
  map<int32, int64> distribution = 3; // key 1..5
  int64 with_media_count = 4;
}

message ReplyReviewReq {
  int64 review_id = 1;
  int64 merchant_user_id = 2;
  string text = 3;
}

message SoftDeleteReviewReq { int64 review_id = 1; }
```

### 5.2 mall-api 路由

| Method | Path | 鉴权 | 说明 |
|--------|------|------|------|
| POST | `/api/upload/review-media` | JWT | multipart/form-data；mall-api 写 MinIO 后返回 `[{type, url}]` |
| POST | `/api/review/submit` | JWT | body: `{orderItemId, scoreMatch, scoreLogistics, scoreService, content, media:[{type,url}]}` |
| POST | `/api/review/followup` | JWT | body: `{reviewId, content, media}` |
| GET  | `/api/review/:id` | 公开 | |
| GET  | `/api/product/:productId/reviews` | 公开 | query: `sort,score,withMedia,page,pageSize` |
| GET  | `/api/product/:productId/rating-summary` | 公开 | |
| GET  | `/api/user/reviews` | JWT | "我的评价" |
| POST | `/api/admin/review/:id/reply` | `X-Admin-Token` | body: `{merchantUserId, text}` |
| DELETE | `/api/admin/review/:id` | `X-Admin-Token` | 软删 |

### 5.3 mall-api yaml 新增配置

```yaml
ReviewMedia:
  MaxImages: 9
  MaxImageSizeMB: 5
  MaxVideoSizeMB: 50
  Bucket: mall-review-media

AdminToken: "mall-admin-token-change-in-production"

MinIO:
  Endpoint: 127.0.0.1:9000
  AccessKey: admin
  SecretKey: admin123
  UseSSL: false

ReviewRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: review.rpc
```

---

## 6. 校验与错误码

### 6.1 校验规则

| 入口 | 校验 |
|------|------|
| SubmitReview | order_item.user_id == user_id；order.status == 已完成；review.order_item_id 不存在；3 个分项 ∈ [1,5]；content 长度 5–2000；media ≤ 9 图 + ≤ 1 视频；media URL 必须以 MinIO bucket 配置前缀开头；调 risk-rpc 通过 |
| SubmitFollowup | 原 review 存在 & user_id 匹配；followup_content IS NULL；create_time ≤ NOW()-7d；followup_content 长度 5–500 |
| ReplyReview | mall-api 层 `X-Admin-Token` 命中；review 存在且 status=0；text 长度 5–500 |
| SoftDeleteReview | mall-api 层 `X-Admin-Token` 命中；review 存在；幂等（重复软删返回成功） |
| 媒体上传（mall-api） | 单图 ≤ 5MB；单视频 ≤ 50MB；总图 ≤ 9 张；总视频 ≤ 1 个；MIME 类型在白名单 |

### 6.2 错误码（追加到 `mall-common/errorx`，沿用 4 位 8xxx 段）

| code | const | message |
|------|-------|---------|
| 8001 | `ReviewOrderNotFound` | 订单不存在或不属于当前用户 |
| 8002 | `ReviewOrderNotCompleted` | 订单未完成，不能评价 |
| 8003 | `ReviewAlreadyExists` | 该订单项已评价过 |
| 8004 | `ReviewNotFound` | 评价不存在或已删除 |
| 8005 | `ReviewFollowupNotAllowed` | 追评条件不满足 |
| 8006 | `ReviewRiskBlocked` | 风控拦截，无法发表评价 |
| 8007 | `ReviewMediaInvalid` | 媒体 URL 非法 |
| 8008 | `ReviewLimitExceeded` | 字数 / 媒体数量超限 |
| 8009 | `AdminTokenInvalid` | 管理员 token 无效 |

---

## 7. 缓存一致性

写路径（`SubmitReview` / `SubmitFollowup` / `SoftDeleteReview`）末尾**必须** `DEL mall:review:summary:{productId}`。  
读路径（`GetProductRatingSummary`）：缓存命中直接返回；未命中走 SQL 聚合（`SELECT score_overall, has_media WHERE product_id=? AND status=0`），算出结构后写回 Redis（TTL 300s）。  
`ReplyReview` 不影响聚合，无需 invalidate。

并发场景：极端情况下两个并发提交各自 DEL 后，下一个读触发回源 + 写回——可接受，不做分布式锁。

---

## 8. 测试策略

**review-rpc 单元测试**（`*_test.go`）

- 每个 logic 文件 1–2 个用例：mock model + mock order/risk client
- 覆盖正反路径：成功 / 订单不存在 / 订单未完成 / 已存在评价 / 风控拦截 / 媒体非法

**review-rpc 集成测试**（`integration_test.go`，build tag `integration`）

- 起 testcontainers MySQL + miniredis，跑一个端到端用例：插测试数据 → SubmitReview → ListProductReviews → GetProductRatingSummary → SubmitFollowup → 检查缓存被失效

**mall-api 测试**

- 至少 2 个 handler 用例：`POST /api/review/submit` 200 / 风控 403
- 至少 1 个上传用例：`POST /api/upload/review-media` 接受图片，拒绝超大文件

**手工 QA 清单**（demo 用）

1. 注册 → 浏览商品 → 下单 → 模拟支付完成 → 完成订单 → 提交评价（含 3 图 + 1 视频）→ 商品详情页看到该评价 + 评分聚合刷新
2. 7 天后追评（demo 时直接改 DB 时间戳模拟）
3. admin token 写商家回复 / 软删

---

## 9. 子项目落地顺序提醒

- **本 spec 范围**：仅 `mall-review-rpc` + 上述 mall-api 路由 / 配置变更（因路由是评价的天然消费方，一并落地不拆）。
- **下一个 spec**：`mall-logistics-rpc`（结构与 review 相似）。
- **再下一个 spec**：C 端前端（消费完整后端）。

---

## 10. 待办

- [ ] User 最终 review 本 spec
- [ ] 如确认无问题，进入 writing-plans 流程产出实现计划
