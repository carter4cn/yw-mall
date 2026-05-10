# Changelog

All notable changes to yw-mall are documented here.

---

## [Unreleased] — feature/etcd-config-center

### 2026-05-10

#### feat: etcd 配置中心 + 热更新

**背景**

所有微服务原本在启动时直接用 `conf.MustLoad` 读取本地 YAML 文件，运行时配置无法更新，每次改动都需重启服务。本次迭代引入 etcd 作为统一配置中心，服务启动时优先从 etcd 加载，etcd 不可用时回退到本地 YAML；关键服务（mall-api、mall-user-rpc）还支持运行时热更新，无需重启即可生效。

---

#### 新增 — mall-common/configcenter 包

新增三个文件，构成配置中心的核心能力：

**`mall-common/configcenter/loader.go`**
- `EtcdHostsFromEnv()` — 读取环境变量 `ETCD_HOSTS`（逗号分隔），容器和本地开发均适用；未设置时返回 nil，触发本地文件回退
- `MustLoadWithFallback(etcdHosts, key, localPath, dest)` — 统一加载入口：有 etcd 主机时从 etcd 拉取并反序列化 YAML，否则退回 `conf.MustLoad` 读本地文件；任一方式失败均 panic（与原有行为一致）
- `LoadWithFallback(...)` — 同上，返回 error 而非 panic，供有错误处理需求的调用方使用

**`mall-common/configcenter/watcher.go`**
- `Watcher` 结构体，封装 etcd clientv3 Watch API
- `Watch(key, onChange)` — 在独立 goroutine 中持续监听 key 变化；网络断开时以 5 秒退避自动重连，保证长运行稳定性
- `onChange(newValue []byte)` 回调由调用方实现，负责解析新配置并原子替换运行时状态

**`mall-common/configcenter/hotreload.go`**
- `HotConfig[T any]` 泛型结构体，内置 `sync.RWMutex`
- `Get()` 持读锁返回当前值（高并发安全）
- `Set(v T)` 持写锁原子替换（Watch 回调中调用）

**`mall-common/go.mod`** — 新增直接依赖：
- `go.etcd.io/etcd/client/v3 v3.5.21`
- `gopkg.in/yaml.v3 v3.0.1`

---

#### 修改 — mall-api（热更新：AdminToken + MinIO）

**`mall-api/mall.go`**
- 替换 `conf.MustLoad` → `configcenter.MustLoadWithFallback`，启动时优先从 `/mall/config/mall-api` 拉取配置

**`mall-api/internal/svc/servicecontext.go`**
- 新增 `hotMinioClient`：实现 `ObjectStore` 接口，内部用 `atomic.Value` 存储实际客户端，支持运行时 `swap()` 无锁切换
- `ServiceContext` 新增 `adminTokenHot *configcenter.HotConfig[string]`
- `NewServiceContext` 新增 `etcdHosts []string` 参数，启动后台 goroutine 监听 `/mall/config/mall-api`
- `onConfigChange()` 收到新配置后：解析 YAML → 更新 `adminTokenHot` → 用新凭据重建 MinIO 客户端并 swap

**`mall-api/internal/middleware/admintokenmiddleware.go`**
- `AdminTokenMiddleware.token` 类型从 `string` 改为 `*configcenter.HotConfig[string]`
- `Handle()` 每次请求调用 `m.token.Get()`，实时读取最新 token，热更新后立即生效

**`mall-api/go.mod`** — 引入 `mall-common/configcenter`（通过 replace 指令）

---

#### 修改 — mall-user-rpc（热更新：JWT AccessSecret）

**`mall-user-rpc/user.go`**
- 替换 `conf.MustLoad` → `configcenter.MustLoadWithFallback`，从 `/mall/config/user-rpc` 加载

**`mall-user-rpc/internal/svc/servicecontext.go`**
- 新增 `JwtSecretHot *configcenter.HotConfig[string]`
- 启动后台 goroutine 监听 `/mall/config/user-rpc`，收到变更后更新 `JwtSecretHot`

**`mall-user-rpc/internal/logic/loginlogic.go`**
- 生成 JWT token 时从 `l.svcCtx.JwtSecretHot.Get()` 读取密钥，确保热更新后新签发的 token 使用最新密钥

**`mall-user-rpc/go.mod`** — 引入 `mall-common/configcenter`

---

#### 修改 — 其余 13 个 RPC 服务（启动时从 etcd 加载，无热更新）

以下服务均按相同模式改造入口文件，替换 `conf.MustLoad`：

| 服务 | etcd key | 入口文件 |
|------|----------|---------|
| mall-product-rpc | `/mall/config/product-rpc` | `product.go` |
| mall-order-rpc | `/mall/config/order-rpc` | `order.go` |
| mall-cart-rpc | `/mall/config/cart-rpc` | `cart.go` |
| mall-payment-rpc | `/mall/config/payment-rpc` | `payment.go` |
| mall-activity-rpc | `/mall/config/activity-rpc` | `activity.go` |
| mall-workflow-rpc | `/mall/config/workflow-rpc` | `workflow.go` |
| mall-reward-rpc | `/mall/config/reward-rpc` | `reward.go` |
| mall-risk-rpc | `/mall/config/risk-rpc` | `risk.go` |
| mall-review-rpc | `/mall/config/review-rpc` | `review.go` |
| mall-logistics-rpc | `/mall/config/logistics-rpc` | `logistics.go` |
| mall-shop-rpc | `/mall/config/shop-rpc` | `shop.go` |
| mall-rule-rpc | `/mall/config/rule-rpc` | `rule.go` |
| mall-activity-async-worker | `/mall/config/activity-async-worker` | `worker.go` |

各服务对应 `go.mod` 均新增 `mall-common` 依赖。

---

#### 设计决策

- **回退策略**：`ETCD_HOSTS` 未设置 → 直接读本地 YAML，本地开发零改造
- **etcd key 命名**：`/mall/config/{service-short-name}`，与 yw-mall-deploy 的 push/pull 脚本保持一致
- **热更新范围**：仅 mall-api（AdminToken、MinIO）和 mall-user-rpc（JWT 密钥）做热更新；其他服务配置变更频率低，重建镜像即可，不增加复杂度
- **线程安全**：`HotConfig[T]` 读写均加锁；MinIO 客户端通过 `hotMinioClient`（内部 `atomic.Value`）实现无锁原子 swap

---

## 历史版本

### 2026-05-09

#### fix: 种子数据与部署修复

- fix(seed): 绕过 gRPC 超时，修复动态用户 ID 解析
- fix(docker): 所有镜像名改为全限定格式（兼容 Podman）
- fix(mall-reward-rpc): go mod tidy 清理过期依赖
- chore: 添加 Dockerfile 和 docker-entrypoint，支持容器化部署
