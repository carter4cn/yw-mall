#!/bin/bash
# One-key startup for the mall microservice stack.
#
# Layers:
#   1. infra: brings up the docker compose stack (Kafka/MySQL/ProxySQL/Redis/Etcd/DTM/...).
#   2. bootstrap: creates databases + applies DDL + seeds workflows / sample activities.
#   3. services: launches all 12 go-zero RPC + API binaries via `go run`.
#
# Usage:
#   start.sh              -> alias for `start` (infra check, bootstrap if missing, then services)
#   start.sh start        -> idempotent full startup
#   start.sh stop         -> stops only go services (compose stack stays up)
#   start.sh restart      -> stop + start
#   start.sh status       -> show go service status
#   start.sh bootstrap    -> only run db / seed bootstrap, do not (re)start services
#   start.sh nuke         -> stop services + drop all mall_* databases + flush stale Redis cache
#                            (useful when you want a clean slate; does NOT touch infra containers)

set -u

BASE_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_DIR="$(cd "$BASE_DIR/../env" && pwd)"
LOG_DIR="$BASE_DIR/logs"
PID_FILE="$BASE_DIR/.pids"
BOOTSTRAP_MARKER="$BASE_DIR/.bootstrapped"

mkdir -p "$LOG_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ---- container engine detection (docker / podman) ----
if command -v podman >/dev/null 2>&1; then
    CRI=podman
elif command -v docker >/dev/null 2>&1; then
    CRI=docker
else
    echo -e "${RED}Neither docker nor podman is installed.${NC}" >&2
    exit 1
fi
COMPOSE="$CRI compose"

PROXY_MYSQL='mysql -h127.0.0.1 -P6033 -uproxysql -pproxysql123'
REDIS_CLI() { $CRI exec -i redis-master redis-cli "$@" 2>/dev/null; }

# ---- service definitions: directory:entrypoint:name:port ----
SERVICES=(
    "mall-user-rpc:user.go:user-rpc:19001"
    "mall-product-rpc:product.go:product-rpc:9002"
    "mall-order-rpc:order.go:order-rpc:9003"
    "mall-cart-rpc:cart.go:cart-rpc:9004"
    "mall-payment-rpc:payment.go:payment-rpc:9005"
    "mall-rule-rpc:rule.go:rule-rpc:9011"
    "mall-risk-rpc:risk.go:risk-rpc:9014"
    "mall-review-rpc:review.go:review-rpc:9015"
    "mall-reward-rpc:reward.go:reward-rpc:9013"
    "mall-activity-rpc:activity.go:activity-rpc:9010"
    "mall-workflow-rpc:workflow.go:workflow-rpc:9012"
    "mall-activity-async-worker:worker.go:activity-async-worker:0"
    "mall-api:mall.go:mall-api:18888"
)

# infra containers we care about (compose-managed). probe by name.
REQUIRED_CONTAINERS=(
    etcd1 kafka1 redis-master proxysql
    mysql-master1 mysql-master2 mysql-slave1 mysql-slave2
    dtm
)

# ---------- helpers ----------
log()  { echo -e "${BLUE}[start.sh]${NC} $*"; }
ok()   { echo -e "  ${GREEN}OK${NC} $*"; }
warn() { echo -e "  ${YELLOW}WARN${NC} $*"; }
err()  { echo -e "  ${RED}ERR${NC} $*"; }

container_running() {
    $CRI ps --format '{{.Names}}' 2>/dev/null | grep -qx "$1"
}

port_listener_pid() {
    ss -tlnp 2>/dev/null | grep -E ":$1 " | grep -oE 'pid=[0-9]+' | cut -d= -f2 | head -1
}

wait_for_mysql() {
    local tries=30
    while ((tries > 0)); do
        if $PROXY_MYSQL -e 'SELECT 1' >/dev/null 2>&1; then return 0; fi
        sleep 1; ((tries--))
    done
    return 1
}

# ---------- infra ----------
infra_up() {
    log "Checking infra containers..."
    local missing=()
    for c in "${REQUIRED_CONTAINERS[@]}"; do
        if container_running "$c"; then ok "$c"; else missing+=("$c"); warn "$c not running"; fi
    done
    if ((${#missing[@]} > 0)); then
        log "Bringing up infra via compose (${#missing[@]} missing)..."
        ( cd "$ENV_DIR" && $COMPOSE up -d "${missing[@]}" ) || {
            err "compose up failed"; return 1;
        }
        sleep 3
    fi
    log "Waiting for ProxySQL to accept SQL..."
    if wait_for_mysql; then ok "ProxySQL ready"; else err "ProxySQL not reachable on 6033"; return 1; fi
}

# ---------- bootstrap ----------
bootstrap_dbs() {
    log "Creating databases..."
    for db in mall_user mall_product mall_order mall_cart mall_payment \
              mall_activity mall_rule mall_workflow mall_reward mall_risk; do
        $PROXY_MYSQL -e "CREATE DATABASE IF NOT EXISTS $db CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;" 2>/dev/null \
            && ok "$db"
    done

    log "Applying DDL..."
    declare -A DDL=(
        [mall_user]=mall-user-rpc/sql/user.sql
        [mall_product]=mall-product-rpc/sql/product.sql
        [mall_order]=mall-order-rpc/sql/order.sql
        [mall_cart]=mall-cart-rpc/sql/cart.sql
        [mall_payment]=mall-payment-rpc/sql/payment.sql
        [mall_activity]=mall-activity-rpc/sql/activity.sql
        [mall_rule]=mall-rule-rpc/sql/rule.sql
        [mall_workflow]=mall-workflow-rpc/sql/workflow.sql
        [mall_reward]=mall-reward-rpc/sql/reward.sql
        [mall_risk]=mall-risk-rpc/sql/risk.sql
        [mall_review]=mall-review-rpc/sql/review.sql
    )
    for db in "${!DDL[@]}"; do
        local f="$BASE_DIR/${DDL[$db]}"
        if [ ! -f "$f" ]; then warn "missing $f"; continue; fi
        $PROXY_MYSQL "$db" < "$f" 2>&1 | grep -v "^Warning" | grep . || true
        ok "$db <- $(basename "$f")"
    done
}

flush_stale_caches() {
    log "Flushing potentially-stale go-zero negative caches..."
    # cache:* — go-zero CachedConn (incl. negative lookups that block fresh registrations).
    # activity:* — Lua-managed seckill/coupon/lottery state (dedup sets, stock counters,
    #              user_claims, prize pools); republishing reuses the same activity_id.
    for pat in 'cache:*' 'activity:*'; do
        REDIS_CLI EVAL 'for _,k in ipairs(redis.call("keys", ARGV[1])) do redis.call("del", k) end return 1' \
            0 "$pat" >/dev/null
        ok "$pat keys flushed"
    done
}

bootstrap_seed() {
    # workflow definitions and sample activities — only run if not yet bootstrapped
    if [ -f "$BOOTSTRAP_MARKER" ] && [ "${1:-}" != "force" ]; then
        ok "seed already applied (delete $BOOTSTRAP_MARKER to re-run)"
        return 0
    fi

    log "Seeding workflow definitions..."
    if ! container_running etcd1 || ! port_listener_pid 9012 >/dev/null 2>&1; then
        warn "workflow-rpc not running yet; skipping (re-run start.sh bootstrap once services are up)"
    else
        ( cd "$BASE_DIR/mall-workflow-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "workflow seed had errors (may be benign on re-run)"
    fi

    log "Seeding reward templates..."
    if ! port_listener_pid 9013 >/dev/null 2>&1; then
        warn "reward-rpc not running yet; skipping"
    else
        ( cd "$BASE_DIR/mall-reward-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "reward seed had errors"
    fi

    log "Seeding sample activities..."
    if ! port_listener_pid 9010 >/dev/null 2>&1; then
        warn "activity-rpc not running yet; skipping"
    else
        ( cd "$BASE_DIR/mall-activity-rpc" && go run cmd/seed/main.go ) 2>&1 | sed 's/^/  /' || warn "activity seed had errors"
    fi

    touch "$BOOTSTRAP_MARKER"
}

# ---------- go services ----------
kill_port() {
    local port=$1; [ "$port" = "0" ] && return
    local pid; pid=$(port_listener_pid "$port")
    if [ -n "$pid" ]; then kill -9 "$pid" 2>/dev/null && warn "killed leftover pid=$pid on :$port"; fi
}

services_start() {
    log "Starting mall services..."
    : > "$PID_FILE"
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r dir entry name port <<< "$svc"
        echo -n "  Starting $name... "
        kill_port "$port"
        cd "$BASE_DIR/$dir"
        nohup go run "$entry" > "$LOG_DIR/$name.log" 2>&1 &
        local pid=$!
        echo "$name:$pid" >> "$PID_FILE"
        sleep 0.6
        if kill -0 "$pid" 2>/dev/null; then echo -e "${GREEN}OK${NC} (pid: $pid)"; else echo -e "${RED}FAILED${NC} (see $LOG_DIR/$name.log)"; fi
    done
    cd "$BASE_DIR"
    log "API gateway: http://localhost:18888"
    log "Logs: $LOG_DIR/"
}

services_stop() {
    log "Stopping mall services..."
    if [ -f "$PID_FILE" ]; then
        while IFS=':' read -r name pid; do
            echo -n "  Stopping $name (pid: $pid)... "
            if kill -0 "$pid" 2>/dev/null; then kill "$pid" 2>/dev/null; wait "$pid" 2>/dev/null; echo -e "${GREEN}OK${NC}"
            else echo -e "${YELLOW}already stopped${NC}"; fi
        done < "$PID_FILE"
        rm -f "$PID_FILE"
    fi
    # also reap any child binaries that survived (go run spawns a child binary)
    for svc in "${SERVICES[@]}"; do
        IFS=':' read -r _ _ _ port <<< "$svc"
        kill_port "$port"
    done
}

services_status() {
    [ -f "$PID_FILE" ] || { warn "no services running"; return; }
    echo "Service Status:"
    while IFS=':' read -r name pid; do
        if kill -0 "$pid" 2>/dev/null; then echo -e "  $name (pid: $pid) ${GREEN}running${NC}"
        else echo -e "  $name (pid: $pid) ${RED}stopped${NC}"; fi
    done < "$PID_FILE"
}

# ---------- top-level commands ----------
do_start() {
    infra_up || exit 1
    bootstrap_dbs
    flush_stale_caches
    services_start
    sleep 4
    bootstrap_seed
    log "Done. Try: curl -s -X POST http://localhost:18888/api/user/login -H 'Content-Type: application/json' -d '{\"username\":\"alice\",\"password\":\"alice123\"}'"
}

do_nuke() {
    services_stop
    log "Dropping all mall_* databases..."
    for db in mall_user mall_product mall_order mall_cart mall_payment \
              mall_activity mall_rule mall_workflow mall_reward mall_risk; do
        $PROXY_MYSQL -e "DROP DATABASE IF EXISTS $db" 2>/dev/null && ok "dropped $db"
    done
    flush_stale_caches
    rm -f "$BOOTSTRAP_MARKER"
    ok "marker reset; next start.sh start will re-bootstrap"
}

case "${1:-start}" in
    start)     do_start ;;
    stop)      services_stop ;;
    restart)   services_stop; sleep 1; do_start ;;
    status)    services_status ;;
    bootstrap) infra_up && bootstrap_dbs && flush_stale_caches && bootstrap_seed force ;;
    nuke)      do_nuke ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|bootstrap|nuke}"
        exit 1
        ;;
esac
