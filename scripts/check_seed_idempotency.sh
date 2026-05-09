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
( cd /home/carter/workspace/go/yw-mall/mall-shop-rpc    && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-product-rpc && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-user-rpc    && go run cmd/seed/main.go )
( cd /home/carter/workspace/go/yw-mall/mall-order-rpc   && go run cmd/seed/main.go )

after=$(cat <<EOF
shop:$(count mall_shop shop)
product:$(count mall_product product)
addr:$(count mall_user user_address)
order:$(count mall_order \`order\`)
EOF
)
echo "After second seed run:"
echo "$after"

if [ "$before" = "$after" ]; then
  echo "✓ idempotent"
else
  echo "✗ counts changed — seed is NOT idempotent"
  exit 1
fi
