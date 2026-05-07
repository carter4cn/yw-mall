#!/bin/bash
# Verify that seed data has been applied to all mall databases.
# Exits 0 if all checks pass, 1 if any check fails.
# Run after `start.sh bootstrap` or after all services are up.

set -u

PROXY_MYSQL='mysql -h127.0.0.1 -P6033 -uproxysql -pproxysql123 -N -s'

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

PASS=0
FAIL=0

check() {
    local desc="$1"
    local db="$2"
    local query="$3"
    local min="${4:-1}"

    local count
    count=$($PROXY_MYSQL "$db" -e "$query" 2>/dev/null) || { echo -e "  ${RED}FAIL${NC} $desc (db unreachable)"; ((FAIL++)); return; }
    if [ "${count:-0}" -ge "$min" ]; then
        echo -e "  ${GREEN}OK${NC}   $desc (count=$count)"
        ((PASS++))
    else
        echo -e "  ${RED}FAIL${NC} $desc (want >=$min, got ${count:-0})"
        ((FAIL++))
    fi
}

echo "=== Mall Seed Check ==="

echo ""
echo "--- Infrastructure (tables exist) ---"
check "mall_user.user table"         mall_user     "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='mall_user' AND table_name='user'" 1
check "mall_user.user_address table" mall_user     "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='mall_user' AND table_name='user_address'" 1
check "mall_shop.shop table"         mall_shop     "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='mall_shop' AND table_name='shop'" 1
check "mall_product.product table"   mall_product  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='mall_product' AND table_name='product'" 1
check "mall_order.order table"       mall_order    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='mall_order' AND table_name='\`order\`'" 1

echo ""
echo "--- Seed data (row counts) ---"
check "shops (>=6)"         mall_shop    "SELECT COUNT(*) FROM shop WHERE status=1" 6
check "products (>=40)"     mall_product "SELECT COUNT(*) FROM product WHERE status=1" 40
check "user_addresses (>=5)" mall_user   "SELECT COUNT(*) FROM user_address" 5
check "orders (>=5)"        mall_order   "SELECT COUNT(*) FROM \`order\`" 5
check "workflow defs (>=4)" mall_workflow "SELECT COUNT(*) FROM workflow_definition" 4
check "activities (>=4)"    mall_activity "SELECT COUNT(*) FROM activity" 4

echo ""
echo "--- Product shop_id linkage ---"
check "products with shop_id set" mall_product "SELECT COUNT(*) FROM product WHERE shop_id > 0" 40

echo ""
echo "--- Order address linkage ---"
check "orders with address_id set" mall_order "SELECT COUNT(*) FROM \`order\` WHERE address_id > 0" 5

echo ""
echo "--- Order status distribution ---"
check "completed orders (>=1)"  mall_order "SELECT COUNT(*) FROM \`order\` WHERE status=3" 1
check "pending/paid orders (>=1)" mall_order "SELECT COUNT(*) FROM \`order\` WHERE status IN (0,1)" 1

echo ""
echo "=== Result: ${PASS} passed, ${FAIL} failed ==="
[ "$FAIL" -eq 0 ] && echo -e "${GREEN}All checks passed.${NC}" || echo -e "${RED}Some checks failed. Run: start.sh bootstrap${NC}"
exit $FAIL
