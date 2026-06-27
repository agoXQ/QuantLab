#!/usr/bin/env bash
# ============================================================
# QuantLab — 一键启动所有后端服务与 API 网关
# ============================================================
# 用法:
#   ./scripts/dev.sh start    # 构建并后台启动基础设施 + 全部服务 + 网关
#   ./scripts/dev.sh stop     # 停止全部服务/网关 (不停止 Docker 基础设施)
#   ./scripts/dev.sh restart  # stop && start
#   ./scripts/dev.sh status   # 查看各进程存活状态
#   ./scripts/dev.sh logs [svc]  # 查看全部日志, 或指定服务: market/user/...
#   ./scripts/dev.sh down     # 停止服务并关闭 Docker 基础设施
#
# 选项 (仅对 start/restart 生效):
#   --no-build   跳过重新编译, 复用 bin/ 下已有二进制
#   --no-infra   跳过 Docker 基础设施就绪检查 (假定 etcd/redis/postgres 已就绪)
# ============================================================
set -euo pipefail

# ---------- 路径与全局变量 ----------
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

BIN_DIR="$ROOT/bin"
LOG_DIR="$ROOT/logs"
PID_DIR="$ROOT/.run"
INFRA_DIR="$ROOT/infrastructure/docker"

mkdir -p "$BIN_DIR" "$LOG_DIR" "$PID_DIR"

# 加载 .env (若存在), 让 TUSHARE_TOKEN 等环境变量对服务可见
if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

# 服务清单: name|pkg|config
# 顺序大致按依赖: 基础数据/计算在前, 网关最后启动
SERVICES=(
  "market|./app/market|app/market/etc/market.yaml"
  "formula|./app/formula|app/formula/etc/formula.yaml"
  "strategy|./app/strategy|app/strategy/etc/strategy.yaml"
  "backtest|./app/backtest|app/backtest/etc/backtest.yaml"
  "ranking|./app/ranking|app/ranking/etc/ranking.yaml"
  "portfolio|./app/portfolio|app/portfolio/etc/portfolio.yaml"
  "community|./app/community|app/community/etc/community.yaml"
  "billing|./app/billing|app/billing/etc/billing.yaml"
  "user|./app/user|app/user/etc/user.yaml"
  "notification|./app/notification|app/notification/etc/notification.yaml"
  "ai|./app/ai|app/ai/etc/ai.yaml"
)
GATEWAY="gateway|./app/gateway|app/gateway/etc/gateway.yaml"

# ---------- 颜色 ----------
if [[ -t 1 ]]; then
  C_GREEN=$'\033[32m'; C_RED=$'\033[31m'; C_YELLOW=$'\033[33m'
  C_CYAN=$'\033[36m'; C_DIM=$'\033[2m'; C_RESET=$'\033[0m'
else
  C_GREEN=""; C_RED=""; C_YELLOW=""; C_CYAN=""; C_DIM=""; C_RESET=""
fi

log()  { echo "${C_CYAN}[dev]${C_RESET} $*"; }
ok()   { echo "${C_GREEN}[ok]${C_RESET}  $*"; }
warn() { echo "${C_YELLOW}[warn]${C_RESET} $*"; }
err()  { echo "${C_RED}[err]${C_RESET} $*" >&2; }

# ---------- 选项解析 ----------
DO_BUILD=1
DO_INFRA=1
ACTION="${1:-start}"; shift || true
while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-build) DO_BUILD=0; shift;;
    --no-infra) DO_INFRA=0; shift;;
    *) err "未知参数: $1"; exit 2;;
  esac
done

# ---------- 工具函数 ----------
pid_file() { echo "$PID_DIR/$1.pid"; }
log_file() { echo "$LOG_DIR/$1.log"; }

is_running() {
  local pid
  pid="$(cat "$(pid_file "$1")" 2>/dev/null || true)"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

wait_port() {
  # 等待 host:port 可连, 最多 ~$2 秒
  local hostport="$1" max="${2:-20}" i=0
  while (( i < max )); do
    if (echo > "/dev/tcp/${hostport%%:*}/${hostport##*:}") 2>/dev/null; then
      return 0
    fi
    sleep 1; ((i++))
  done
  return 1
}

# ---------- 基础设施 ----------
ensure_infra() {
  if [[ "$DO_INFRA" -eq 0 ]]; then
    warn "跳过基础设施检查 (--no-infra)"
    return 0
  fi
  if ! command -v docker >/dev/null 2>&1; then
    warn "未检测到 docker, 假定基础设施已在外部就绪"
    return 0
  fi
  log "启动 Docker 基础设施 (etcd / redis / postgres / kafka / minio) ..."
  ( cd "$INFRA_DIR" && docker compose up -d ) >/dev/null 2>&1 || {
    err "docker compose up 失败, 请检查 $INFRA_DIR/docker-compose.yml"
    exit 1
  }
  log "等待 etcd / redis / postgres 就绪 ..."
  wait_port "127.0.0.1:2379" 30 || warn "etcd(2379) 未就绪, 依赖服务注册的服务可能启动失败"
  wait_port "127.0.0.1:6379" 30 || warn "redis(6379) 未就绪"
  wait_port "127.0.0.1:5432" 40 || warn "postgres(5432) 未就绪"
  ok "基础设施就绪"
}

# ---------- 编译 ----------
build_all() {
  if [[ "$DO_BUILD" -eq 0 ]]; then
    warn "跳过编译 (--no-build)"
    return 0
  fi
  log "编译全部服务与网关 ..."
  local pkgs=()
  for s in "${SERVICES[@]}" "$GATEWAY"; do
    pkgs+=("${s#*|}")
  done
  # 串行构建, 任一失败立即中止
  for s in "${SERVICES[@]}" "$GATEWAY"; do
    local name="${s%%|*}"; local pkg="${s#*|}"; pkg="${pkg%|*}"
    printf "  -> %-14s %s\n" "$name" "$pkg"
    go build -o "$BIN_DIR/$name" "$pkg" || { err "编译 $name 失败"; exit 1; }
  done
  ok "编译完成"
}

# ---------- 启停 ----------
start_one() {
  local entry="$1"
  local name="${entry%%|*}"; local rest="${entry#*|}"
  local pkg="${rest%|*}"; local cfg="${rest##*|}"

  if is_running "$name"; then
    warn "$name 已在运行 (pid $(cat "$(pid_file "$name")"))"
    return 0
  fi

  local bin="$BIN_DIR/$name"
  if [[ ! -x "$bin" ]]; then
    err "$name 二进制不存在: $bin (请去掉 --no-build)"
    return 1
  fi

  log "启动 $name (config: $cfg)"
  nohup "$bin" -f "$cfg" >"$(log_file "$name")" 2>&1 &
  echo $! >"$(pid_file "$name")"
}

start_all() {
  ensure_infra
  build_all

  log "启动后端服务 ..."
  for s in "${SERVICES[@]}"; do
    start_one "$s"
  done

  # 给核心 gRPC 服务一点时间完成 etcd 注册, 再启动网关
  log "等待服务绑定端口 ..."
  local checks=(
    "market:127.0.0.1:8082" "formula:127.0.0.1:8098" "strategy:127.0.0.1:8084"
    "backtest:127.0.0.1:8092" "user:127.0.0.1:8086" "notification:127.0.0.1:8088"
  )
  for c in "${checks[@]}"; do
    local n="${c%%:*}" p="${c##*:}"
    wait_port "127.0.0.1:$p" 25 || warn "$n ($p) 端口未就绪, 网关可能连不上该服务"
  done

  log "启动 gateway ..."
  start_one "$GATEWAY"
  wait_port "127.0.0.1:8000" 25 || warn "gateway(8000) 未就绪, 查看日志: $LOG_DIR/gateway.log"

  echo
  ok "全部已启动. 网关: http://127.0.0.1:8000  健康检查: /healthz"
  echo "${C_DIM}查看状态: ./scripts/dev.sh status | 查看日志: ./scripts/dev.sh logs [svc]${C_RESET}"
}

stop_one() {
  local name="$1"
  if is_running "$name"; then
    local pid; pid="$(cat "$(pid_file "$name")")"
    kill "$pid" 2>/dev/null || true
    # 最多等 5s 让其优雅退出
    for _ in 1 2 3 4 5; do
      kill -0 "$pid" 2>/dev/null || break
      sleep 1
    done
    kill -9 "$pid" 2>/dev/null || true
    ok "已停止 $name (pid $pid)"
  else
    rm -f "$(pid_file "$name")"
  fi
}

stop_all() {
  log "停止服务与网关 ..."
  stop_one "gateway"
  for s in "${SERVICES[@]}"; do
    stop_one "${s%%|*}"
  done
  ok "全部已停止"
}

down_all() {
  stop_all
  if command -v docker >/dev/null 2>&1; then
    log "关闭 Docker 基础设施 ..."
    ( cd "$INFRA_DIR" && docker compose down ) >/dev/null 2>&1 || true
    ok "基础设施已关闭"
  fi
}

status_all() {
  printf "%-14s %-8s %-8s\n" "SERVICE" "STATUS" "PID"
  local all=( "$GATEWAY" "${SERVICES[@]}" )
  for s in "${all[@]}"; do
    local name="${s%%|*}"
    if is_running "$name"; then
      printf "%-14s ${C_GREEN}%-8s${C_RESET} %-8s\n" "$name" "running" "$(cat "$(pid_file "$name")")"
    else
      printf "%-14s ${C_DIM}%-8s${C_RESET} %-8s\n" "$name" "stopped" "-"
    fi
  done
}

show_logs() {
  local svc="${1:-}"
  if [[ -n "$svc" ]]; then
    local f; f="$(log_file "$svc")"
    [[ -f "$f" ]] || { err "无日志文件: $f"; exit 1; }
    tail -f "$f"
  else
    tail -f "$LOG_DIR"/*.log
  fi
}

# ---------- 分发 ----------
case "$ACTION" in
  start)   start_all;;
  stop)    stop_all;;
  restart) stop_all; start_all;;
  status)  status_all;;
  logs)    show_logs "${1:-}";;
  down)    down_all;;
  *) err "未知命令: $ACTION"; echo "用法: $0 {start|stop|restart|status|logs [svc]|down} [--no-build] [--no-infra]"; exit 2;;
esac
