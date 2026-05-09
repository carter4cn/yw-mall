# 前端后端扩展 / E2E 验证清单

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
curl -s "http://127.0.0.1:18888/api/shop/products/1" | jq '.products | length'   # 应 ≥ 6
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
curl -s -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:18888/api/shop/1/is-following" | jq
curl -s -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:18888/api/shop/my-followed" | jq
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
