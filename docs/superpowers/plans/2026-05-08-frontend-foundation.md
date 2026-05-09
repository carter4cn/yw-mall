# C-Side Frontend Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the `mall-frontend` uni-app H5 project with scaffold, design tokens, typed API layer, Pinia stores, and 5 browsing-domain pages (Home, Shop List, Shop Detail, Product List, Product Detail).

**Architecture:** Foundation-first — Task 1 creates the compilable scaffold; Tasks 2–5 layer in tokens, types, stores, and API modules; Tasks 6–10 build one page each wired to real backend API calls; Task 11 adds the login placeholder and smoke checklist.

**Tech Stack:** uni-app (Vue 3 + TypeScript + Vite), Wot Design Uni (easycom auto-registration), Pinia 2 + pinia-plugin-persistedstate, SCSS design tokens, H5 target with dev proxy → `http://localhost:18888`.

**Spec:** `docs/superpowers/specs/2026-05-08-frontend-foundation-design.md`

**Working directory for all commands:** `/home/carter/workspace/go/yw-mall` (monorepo root) unless noted.

---

## File Map

| File | Task | Purpose |
|---|---|---|
| `mall-frontend/` | 1 | Project root (created by degit) |
| `mall-frontend/vite.config.ts` | 1 | Vite config: H5 proxy + SCSS tokens global import |
| `mall-frontend/src/manifest.json` | 1 | H5 router mode = history |
| `mall-frontend/src/pages.json` | 1 | Routes + tabBar + easycom for Wot Design Uni |
| `mall-frontend/src/main.ts` | 1 | Pinia setup |
| `mall-frontend/src/styles/tokens.scss` | 2 | Design tokens (colors, spacing, typography) |
| `mall-frontend/src/types/api.ts` | 3 | Response shape types (ShopItem, ProductDetailResp, …) |
| `mall-frontend/src/api/request.ts` | 3 | Base uni.request wrapper with JWT + 401 handler |
| `mall-frontend/src/stores/user.ts` | 4 | token, userId, setToken, clear — persisted |
| `mall-frontend/src/stores/cart.ts` | 4 | count, increment, decrement |
| `mall-frontend/src/api/shop.ts` | 5 | 8 shop endpoint functions |
| `mall-frontend/src/api/product.ts` | 5 | 3 product endpoint functions |
| `mall-frontend/src/pages/index/index.vue` | 6 | Home page |
| `mall-frontend/src/pages/shop/list.vue` | 7 | Shop List page |
| `mall-frontend/src/pages/shop/detail.vue` | 8 | Shop Detail page |
| `mall-frontend/src/pages/product/list.vue` | 9 | Product List page |
| `mall-frontend/src/pages/product/detail.vue` | 10 | Product Detail page |
| `mall-frontend/src/pages/login/index.vue` | 11 | Login placeholder shell |
| `docs/e2e-frontend-backend-checklist.md` | 11 | Append UI smoke checklist section |

---

## Task 1: Scaffold — uni-app Project + Dependencies + Base Config

**Files:**
- Create: `mall-frontend/` (via degit)
- Modify: `mall-frontend/vite.config.ts`
- Modify: `mall-frontend/src/manifest.json`
- Modify: `mall-frontend/src/pages.json`
- Modify: `mall-frontend/src/main.ts`

- [ ] **Step 1: Scaffold the uni-app project**

```bash
cd /home/carter/workspace/go/yw-mall
npx degit dcloudio/uni-preset-vue#vite-ts mall-frontend
cd mall-frontend
```

Expected: directory `mall-frontend/` created with `src/`, `vite.config.ts`, `package.json`, `tsconfig.json`.

- [ ] **Step 2: Install dependencies**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm install
pnpm add wot-design-uni pinia pinia-plugin-persistedstate
pnpm add -D @types/node
```

Expected: `node_modules/` created, `pnpm-lock.yaml` updated. No errors.

- [ ] **Step 3: Verify bare scaffold builds**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5
```

Expected: `dist/build/h5/` created, exit 0. If this fails, do not proceed — check `package.json` scripts for the correct build command name.

- [ ] **Step 4: Replace vite.config.ts**

Write `mall-frontend/vite.config.ts`:

```typescript
import path from 'path'
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'

export default defineConfig({
  plugins: [uni()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:18888',
        changeOrigin: true,
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@import "@/styles/tokens.scss";',
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
})
```

- [ ] **Step 5: Set H5 router mode to history in src/manifest.json**

Open `mall-frontend/src/manifest.json`. Find or add the `"h5"` key and set:

```json
{
  "name": "Mall",
  "appid": "",
  "description": "C端购物H5",
  "versionName": "1.0.0",
  "versionCode": "100",
  "transformPx": false,
  "h5": {
    "title": "Mall",
    "router": {
      "mode": "history"
    }
  }
}
```

Replace the entire file content with the above (keep any existing `appid` value if present).

- [ ] **Step 6: Replace src/pages.json with full route + easycom config**

Write `mall-frontend/src/pages.json`:

```json
{
  "easycom": {
    "autoscan": true,
    "custom": {
      "^wd-(.*)": "wot-design-uni/components/wd-$1/wd-$1.vue"
    }
  },
  "pages": [
    {
      "path": "pages/index/index",
      "style": { "navigationBarTitleText": "首页" }
    },
    {
      "path": "pages/shop/list",
      "style": { "navigationBarTitleText": "店铺列表" }
    },
    {
      "path": "pages/shop/detail",
      "style": { "navigationBarTitleText": "店铺详情" }
    },
    {
      "path": "pages/product/list",
      "style": { "navigationBarTitleText": "商品列表" }
    },
    {
      "path": "pages/product/detail",
      "style": { "navigationBarTitleText": "商品详情" }
    },
    {
      "path": "pages/login/index",
      "style": { "navigationBarTitleText": "登录" }
    }
  ],
  "tabBar": {
    "color": "#999999",
    "selectedColor": "#FF4B4B",
    "borderStyle": "black",
    "backgroundColor": "#FFFFFF",
    "list": [
      {
        "pagePath": "pages/index/index",
        "text": "首页"
      },
      {
        "pagePath": "pages/index/index",
        "text": "购物车"
      },
      {
        "pagePath": "pages/index/index",
        "text": "我的"
      }
    ]
  },
  "globalStyle": {
    "navigationBarTextStyle": "black",
    "navigationBarTitleText": "Mall",
    "backgroundColor": "#F5F5F5"
  }
}
```

- [ ] **Step 7: Update src/main.ts to mount Pinia**

Write `mall-frontend/src/main.ts`:

```typescript
import { createSSRApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import App from './App.vue'

export function createApp() {
  const app = createSSRApp(App)
  const pinia = createPinia()
  pinia.use(piniaPluginPersistedstate)
  app.use(pinia)
  return { app }
}
```

- [ ] **Step 8: Verify scaffold still builds after config changes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -20
```

Expected: exit 0, no TypeScript errors. If SCSS `@import` fails because `src/styles/tokens.scss` doesn't exist yet, create an empty placeholder:
```bash
mkdir -p src/styles && touch src/styles/tokens.scss
pnpm run build:h5
```

- [ ] **Step 9: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/
git commit -m "feat(frontend): scaffold uni-app H5 project with Pinia + Wot Design Uni easycom"
```

---

## Task 2: Design Tokens

**Files:**
- Create: `mall-frontend/src/styles/tokens.scss`

- [ ] **Step 1: Write the design token file**

Write `mall-frontend/src/styles/tokens.scss`:

```scss
// Brand
$color-primary:       #FF4B4B;
$color-primary-light: #FF7070;

// Neutrals
$color-text-primary:   #1A1A1A;
$color-text-secondary: #666666;
$color-text-hint:      #999999;
$color-bg-page:        #F5F5F5;
$color-bg-card:        #FFFFFF;
$color-border:         #EEEEEE;

// Semantic
$color-success: #07C160;
$color-warning: #FA8C16;
$color-error:   #FF4B4B;

// Spacing (8px base)
$space-xs:  4px;
$space-sm:  8px;
$space-md:  16px;
$space-lg:  24px;
$space-xl:  32px;

// Typography
$font-size-sm:   12px;
$font-size-base: 14px;
$font-size-md:   16px;
$font-size-lg:   18px;
$font-size-xl:   20px;
$font-weight-normal: 400;
$font-weight-medium: 500;
$font-weight-bold:   600;

// Radius
$radius-sm:  4px;
$radius-md:  8px;
$radius-lg:  12px;
$radius-full: 9999px;
```

- [ ] **Step 2: Verify tokens are available globally in a page**

Open `mall-frontend/src/pages/index/index.vue`. Add a temporary style block to check:

```vue
<style lang="scss" scoped>
.test { color: $color-primary; }
</style>
```

Run:
```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0, no "Undefined variable" SCSS error. Then remove the `.test` style block.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/styles/tokens.scss
git commit -m "feat(frontend): add SCSS design tokens (colors, spacing, typography)"
```

---

## Task 3: API Types + Base Request

**Files:**
- Create: `mall-frontend/src/types/api.ts`
- Create: `mall-frontend/src/api/request.ts`

- [ ] **Step 1: Write response types**

Create `mall-frontend/src/types/api.ts`:

```typescript
export interface ShopItem {
  id: number
  name: string
  logo: string
  banner: string
  rating: number
  productCount: number
  description: string
}

export interface ProductDetailResp {
  id: number
  name: string
  description: string
  price: number
  stock: number
  images: string[]
  shopId: number
  status: number
}

export interface ShopDetailResp {
  shop: ShopItem
}

export interface ShopListResp {
  shops: ShopItem[]
  total: number
}

export interface ProductListResp {
  products: ProductDetailResp[]
  total: number
}

export interface IsFollowingResp {
  isFollowing: boolean
}

export interface OkResp {
  ok: boolean
}

export interface ApiError {
  code: number
  message: string
}
```

- [ ] **Step 2: Write the base request wrapper**

Create `mall-frontend/src/api/request.ts`:

```typescript
import { useUserStore } from '@/stores/user'
import type { ApiError } from '@/types/api'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown>
  auth?: boolean
}

export function request<T>(options: RequestOptions): Promise<T> {
  const { url, method = 'GET', data, auth = false } = options

  const header: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (auth) {
    const userStore = useUserStore()
    if (userStore.token) {
      header['Authorization'] = `Bearer ${userStore.token}`
    }
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url,
      method,
      data,
      header,
      success(res) {
        if (res.statusCode === 401) {
          const userStore = useUserStore()
          userStore.clear()
          uni.reLaunch({ url: '/pages/login/index' })
          reject({ code: 401, message: '请先登录' } as ApiError)
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data as T)
        } else {
          const body = res.data as { message?: string; code?: number }
          reject({
            code: body?.code ?? res.statusCode,
            message: body?.message ?? '请求失败',
          } as ApiError)
        }
      },
      fail(err) {
        reject({ code: -1, message: err.errMsg ?? '网络错误，请重试' } as ApiError)
      },
    })
  })
}

export function showError(err: unknown): void {
  const message = (err as ApiError)?.message ?? '请求失败，请重试'
  uni.showToast({ title: message, icon: 'none', duration: 2000 })
}
```

- [ ] **Step 3: Create the api directory placeholder to keep build happy**

```bash
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/api
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/stores
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/types
```

- [ ] **Step 4: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -15
```

Expected: exit 0. If TypeScript complains about `useUserStore` not found, that's expected until Task 4 — create a stub store first:

```bash
# Only if build fails due to missing stores:
cat > src/stores/user.ts << 'EOF'
import { defineStore } from 'pinia'
import { ref } from 'vue'
export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const clear = () => { token.value = '' }
  return { token, clear }
})
EOF
pnpm run build:h5
```

Remove the stub after Task 4 completes.

- [ ] **Step 5: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/types/api.ts mall-frontend/src/api/request.ts
git commit -m "feat(frontend): add API response types and base request wrapper with JWT + 401 handler"
```

---

## Task 4: Pinia Stores

**Files:**
- Create: `mall-frontend/src/stores/user.ts`
- Create: `mall-frontend/src/stores/cart.ts`

- [ ] **Step 1: Write user store**

Write `mall-frontend/src/stores/user.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userId = ref(0)

  function setToken(t: string, id: number) {
    token.value = t
    userId.value = id
  }

  function clear() {
    token.value = ''
    userId.value = 0
  }

  return { token, userId, setToken, clear }
}, {
  persist: true,
})
```

- [ ] **Step 2: Write cart store**

Write `mall-frontend/src/stores/cart.ts`:

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useCartStore = defineStore('cart', () => {
  const count = ref(0)

  function increment() {
    count.value++
  }

  function decrement() {
    if (count.value > 0) count.value--
  }

  return { count, increment, decrement }
})
```

- [ ] **Step 3: Verify build passes with real stores**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0.

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/stores/
git commit -m "feat(frontend): add user and cart Pinia stores"
```

---

## Task 5: API Modules

**Files:**
- Create: `mall-frontend/src/api/shop.ts`
- Create: `mall-frontend/src/api/product.ts`

- [ ] **Step 1: Write shop API module**

Write `mall-frontend/src/api/shop.ts`:

```typescript
import { request } from './request'
import type {
  ShopDetailResp,
  ShopListResp,
  ProductListResp,
  IsFollowingResp,
  OkResp,
} from '@/types/api'

const BASE = '/api/shop'

export const getShopDetail = (id: number) =>
  request<ShopDetailResp>({ url: `${BASE}/detail/${id}` })

export const listShops = (page: number, pageSize: number) =>
  request<ShopListResp>({ url: `${BASE}/list`, data: { page, pageSize } })

export const getRecommendedShops = (limit: number) =>
  request<ShopListResp>({ url: `${BASE}/recommended`, data: { limit } })

export const getShopProducts = (shopId: number, page: number, pageSize: number) =>
  request<ProductListResp>({ url: `${BASE}/${shopId}/products`, data: { page, pageSize } })

export const followShop = (shopId: number) =>
  request<OkResp>({ url: `${BASE}/${shopId}/follow`, method: 'POST', auth: true })

export const unfollowShop = (shopId: number) =>
  request<OkResp>({ url: `${BASE}/${shopId}/follow`, method: 'DELETE', auth: true })

export const isFollowing = (shopId: number) =>
  request<IsFollowingResp>({ url: `${BASE}/${shopId}/following`, auth: true })

export const listFollowedShops = (page: number, pageSize: number) =>
  request<ShopListResp>({ url: `${BASE}/followed`, data: { page, pageSize }, auth: true })
```

- [ ] **Step 2: Write product API module**

Write `mall-frontend/src/api/product.ts`:

```typescript
import { request } from './request'
import type { ProductDetailResp, ProductListResp } from '@/types/api'

const BASE = '/api/product'

export const listProducts = (params: { page: number; pageSize: number; shopId?: number }) =>
  request<ProductListResp>({ url: `${BASE}/list`, data: params })

export const getProductDetail = (id: number) =>
  request<ProductDetailResp>({ url: `${BASE}/detail/${id}` })

export const searchProducts = (keyword: string, page = 1) =>
  request<ProductListResp>({ url: `${BASE}/search`, data: { keyword, page, pageSize: 20 } })
```

- [ ] **Step 3: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0, no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/api/
git commit -m "feat(frontend): add shop and product API modules"
```

---

## Task 6: Home Page

**Files:**
- Modify: `mall-frontend/src/pages/index/index.vue`

The Home page shows a search bar, a recommended-shops horizontal scroll, and a featured-products 2-column grid.

- [ ] **Step 1: Write the Home page**

Write `mall-frontend/src/pages/index/index.vue`:

```vue
<template>
  <view class="page">
    <!-- Search bar -->
    <view class="search-wrap">
      <wd-search
        v-model="keyword"
        placeholder="搜索商品"
        @search="onSearch"
        @clear="keyword = ''"
      />
    </view>

    <!-- Recommended shops -->
    <view class="section">
      <view class="section-title">推荐店铺</view>
      <scroll-view class="shops-scroll" scroll-x>
        <view class="shops-row">
          <view
            v-for="shop in shops"
            :key="shop.id"
            class="shop-card"
            @tap="goShopDetail(shop.id)"
          >
            <image
              class="shop-logo"
              :src="shop.logo || '/static/placeholder.png'"
              mode="aspectFill"
            />
            <text class="shop-name">{{ shop.name }}</text>
            <text class="shop-rating">★ {{ shop.rating }}</text>
          </view>
        </view>
      </scroll-view>
    </view>

    <!-- Featured products -->
    <view class="section">
      <view class="section-title">热门商品</view>
      <view v-if="productsLoading">
        <wd-skeleton :row="3" />
      </view>
      <view v-else-if="products.length === 0">
        <wd-empty description="暂无商品" />
      </view>
      <view v-else class="products-grid">
        <view
          v-for="p in products"
          :key="p.id"
          class="product-card"
          @tap="goProductDetail(p.id)"
        >
          <image
            class="product-img"
            :src="(p.images && p.images[0]) || '/static/placeholder.png'"
            mode="aspectFill"
          />
          <text class="product-name">{{ p.name }}</text>
          <text class="product-price">¥{{ (p.price / 100).toFixed(2) }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRecommendedShops } from '@/api/shop'
import { listProducts } from '@/api/product'
import { showError } from '@/api/request'
import type { ShopItem, ProductDetailResp } from '@/types/api'

const keyword = ref('')
const shops = ref<ShopItem[]>([])
const products = ref<ProductDetailResp[]>([])
const productsLoading = ref(true)

function onSearch() {
  if (!keyword.value.trim()) return
  uni.navigateTo({ url: `/pages/product/list?keyword=${encodeURIComponent(keyword.value.trim())}` })
}

function goShopDetail(id: number) {
  uni.navigateTo({ url: `/pages/shop/detail?id=${id}` })
}

function goProductDetail(id: number) {
  uni.navigateTo({ url: `/pages/product/detail?id=${id}` })
}

onMounted(async () => {
  try {
    const [shopResp, productResp] = await Promise.all([
      getRecommendedShops(5),
      listProducts({ page: 1, pageSize: 10 }),
    ])
    shops.value = shopResp.shops ?? []
    products.value = productResp.products ?? []
  } catch (err) {
    showError(err)
  } finally {
    productsLoading.value = false
  }
})
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding-bottom: $space-xl;
}

.search-wrap {
  padding: $space-sm $space-md;
  background: $color-bg-card;
}

.section {
  margin-top: $space-sm;
  background: $color-bg-card;
  padding: $space-md;
}

.section-title {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
  margin-bottom: $space-sm;
}

.shops-scroll {
  white-space: nowrap;
}

.shops-row {
  display: inline-flex;
  gap: $space-md;
}

.shop-card {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  width: 80px;
  gap: $space-xs;
}

.shop-logo {
  width: 60px;
  height: 60px;
  border-radius: $radius-full;
  background: $color-border;
}

.shop-name {
  font-size: $font-size-sm;
  color: $color-text-primary;
  text-align: center;
  max-width: 80px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shop-rating {
  font-size: $font-size-sm;
  color: $color-warning;
}

.products-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $space-sm;
}

.product-card {
  background: $color-bg-card;
  border-radius: $radius-md;
  overflow: hidden;
  border: 1px solid $color-border;
}

.product-img {
  width: 100%;
  height: 150px;
  background: $color-border;
}

.product-name {
  display: block;
  font-size: $font-size-base;
  color: $color-text-primary;
  padding: $space-xs $space-sm;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-price {
  display: block;
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-primary;
  padding: 0 $space-sm $space-sm;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0, no TypeScript errors.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/index/index.vue
git commit -m "feat(frontend): add Home page with recommended shops and featured products"
```

---

## Task 7: Shop List Page

**Files:**
- Create: `mall-frontend/src/pages/shop/list.vue`

- [ ] **Step 1: Create page directory and write the file**

```bash
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/pages/shop
```

Write `mall-frontend/src/pages/shop/list.vue`:

```vue
<template>
  <view class="page">
    <view v-if="loading && shops.length === 0">
      <wd-skeleton :row="5" />
    </view>
    <view v-else-if="shops.length === 0">
      <wd-empty description="暂无店铺" />
    </view>
    <view v-else>
      <view
        v-for="shop in shops"
        :key="shop.id"
        class="shop-card"
        @tap="goDetail(shop.id)"
      >
        <image
          class="shop-logo"
          :src="shop.logo || '/static/placeholder.png'"
          mode="aspectFill"
        />
        <view class="shop-info">
          <text class="shop-name">{{ shop.name }}</text>
          <text class="shop-desc">{{ shop.description }}</text>
          <view class="shop-meta">
            <text class="shop-rating">★ {{ shop.rating }}</text>
            <text class="shop-count">{{ shop.productCount }} 件商品</text>
          </view>
        </view>
      </view>
    </view>

    <wd-load-more
      :status="loadMoreStatus"
      @loadmore="loadMore"
    />
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listShops } from '@/api/shop'
import { showError } from '@/api/request'
import type { ShopItem } from '@/types/api'

const shops = ref<ShopItem[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const loadMoreStatus = ref<'loadmore' | 'loading' | 'nomore'>('loadmore')

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/shop/detail?id=${id}` })
}

async function fetchPage(p: number) {
  loading.value = true
  try {
    const resp = await listShops(p, 10)
    shops.value = p === 1 ? (resp.shops ?? []) : [...shops.value, ...(resp.shops ?? [])]
    total.value = resp.total ?? 0
    loadMoreStatus.value = shops.value.length >= total.value ? 'nomore' : 'loadmore'
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadMoreStatus.value !== 'loadmore' || loading.value) return
  loadMoreStatus.value = 'loading'
  page.value++
  await fetchPage(page.value)
}

onMounted(() => fetchPage(1))
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding: $space-sm;
}

.shop-card {
  display: flex;
  align-items: center;
  gap: $space-md;
  background: $color-bg-card;
  border-radius: $radius-md;
  padding: $space-md;
  margin-bottom: $space-sm;
}

.shop-logo {
  width: 64px;
  height: 64px;
  border-radius: $radius-full;
  flex-shrink: 0;
  background: $color-border;
}

.shop-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: $space-xs;
  min-width: 0;
}

.shop-name {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.shop-desc {
  font-size: $font-size-sm;
  color: $color-text-secondary;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shop-meta {
  display: flex;
  gap: $space-md;
}

.shop-rating {
  font-size: $font-size-sm;
  color: $color-warning;
}

.shop-count {
  font-size: $font-size-sm;
  color: $color-text-hint;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/shop/list.vue
git commit -m "feat(frontend): add Shop List page with infinite scroll"
```

---

## Task 8: Shop Detail Page

**Files:**
- Create: `mall-frontend/src/pages/shop/detail.vue`

- [ ] **Step 1: Write the file**

Write `mall-frontend/src/pages/shop/detail.vue`:

```vue
<template>
  <view class="page">
    <view v-if="loading">
      <wd-skeleton :row="6" />
    </view>
    <view v-else-if="!shop">
      <wd-empty description="店铺不存在" />
    </view>
    <template v-else>
      <!-- Shop header -->
      <image
        class="banner"
        :src="shop.banner || '/static/placeholder.png'"
        mode="aspectFill"
      />
      <view class="shop-header">
        <image
          class="logo"
          :src="shop.logo || '/static/placeholder.png'"
          mode="aspectFill"
        />
        <view class="shop-info">
          <text class="shop-name">{{ shop.name }}</text>
          <text class="shop-rating">★ {{ shop.rating }}</text>
          <text class="shop-desc">{{ shop.description }}</text>
        </view>
        <wd-button
          class="follow-btn"
          :type="following ? 'info' : 'primary'"
          size="small"
          @tap="toggleFollow"
        >
          {{ following ? '已关注' : '关注' }}
        </wd-button>
      </view>

      <!-- Products section -->
      <view class="section-title">店铺商品</view>
      <view v-if="productsLoading && products.length === 0">
        <wd-skeleton :row="4" />
      </view>
      <view v-else-if="products.length === 0">
        <wd-empty description="暂无商品" />
      </view>
      <view v-else class="products-grid">
        <view
          v-for="p in products"
          :key="p.id"
          class="product-card"
          @tap="goProduct(p.id)"
        >
          <image
            class="product-img"
            :src="(p.images && p.images[0]) || '/static/placeholder.png'"
            mode="aspectFill"
          />
          <text class="product-name">{{ p.name }}</text>
          <text class="product-price">¥{{ (p.price / 100).toFixed(2) }}</text>
        </view>
      </view>
      <wd-load-more :status="loadMoreStatus" @loadmore="loadMoreProducts" />
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getShopDetail, isFollowing, followShop, unfollowShop, getShopProducts } from '@/api/shop'
import { showError } from '@/api/request'
import { useUserStore } from '@/stores/user'
import type { ShopItem, ProductDetailResp } from '@/types/api'

const userStore = useUserStore()

const shopId = ref(0)
const shop = ref<ShopItem | null>(null)
const loading = ref(true)
const following = ref(false)

const products = ref<ProductDetailResp[]>([])
const productPage = ref(1)
const productTotal = ref(0)
const productsLoading = ref(false)
const loadMoreStatus = ref<'loadmore' | 'loading' | 'nomore'>('loadmore')

function goProduct(id: number) {
  uni.navigateTo({ url: `/pages/product/detail?id=${id}` })
}

async function toggleFollow() {
  if (!userStore.token) {
    uni.navigateTo({ url: '/pages/login/index' })
    return
  }
  try {
    if (following.value) {
      await unfollowShop(shopId.value)
      following.value = false
    } else {
      await followShop(shopId.value)
      following.value = true
    }
  } catch (err) {
    showError(err)
  }
}

async function fetchProducts(p: number) {
  productsLoading.value = true
  try {
    const resp = await getShopProducts(shopId.value, p, 10)
    products.value = p === 1 ? (resp.products ?? []) : [...products.value, ...(resp.products ?? [])]
    productTotal.value = resp.total ?? 0
    loadMoreStatus.value = products.value.length >= productTotal.value ? 'nomore' : 'loadmore'
  } catch (err) {
    showError(err)
  } finally {
    productsLoading.value = false
  }
}

async function loadMoreProducts() {
  if (loadMoreStatus.value !== 'loadmore' || productsLoading.value) return
  loadMoreStatus.value = 'loading'
  productPage.value++
  await fetchProducts(productPage.value)
}

onMounted(async () => {
  const pages = getCurrentPages()
  const current = pages[pages.length - 1] as unknown as { options: { id?: string } }
  shopId.value = Number(current.options?.id ?? 0)

  try {
    const resp = await getShopDetail(shopId.value)
    shop.value = resp.shop
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }

  if (userStore.token) {
    try {
      const resp = await isFollowing(shopId.value)
      following.value = resp.isFollowing
    } catch { /* ignore */ }
  }

  await fetchProducts(1)
})
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding-bottom: $space-xl;
}

.banner {
  width: 100%;
  height: 180px;
  background: $color-border;
}

.shop-header {
  display: flex;
  align-items: flex-start;
  gap: $space-md;
  padding: $space-md;
  background: $color-bg-card;
}

.logo {
  width: 56px;
  height: 56px;
  border-radius: $radius-full;
  flex-shrink: 0;
  background: $color-border;
}

.shop-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: $space-xs;
  min-width: 0;
}

.shop-name {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.shop-rating {
  font-size: $font-size-sm;
  color: $color-warning;
}

.shop-desc {
  font-size: $font-size-sm;
  color: $color-text-secondary;
}

.follow-btn {
  flex-shrink: 0;
}

.section-title {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
  padding: $space-md;
  background: $color-bg-card;
  margin-top: $space-sm;
}

.products-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $space-sm;
  padding: 0 $space-sm;
  margin-top: $space-sm;
}

.product-card {
  background: $color-bg-card;
  border-radius: $radius-md;
  overflow: hidden;
  border: 1px solid $color-border;
}

.product-img {
  width: 100%;
  height: 140px;
  background: $color-border;
}

.product-name {
  display: block;
  font-size: $font-size-base;
  color: $color-text-primary;
  padding: $space-xs $space-sm;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-price {
  display: block;
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-primary;
  padding: 0 $space-sm $space-sm;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/shop/detail.vue
git commit -m "feat(frontend): add Shop Detail page with follow toggle and product grid"
```

---

## Task 9: Product List Page

**Files:**
- Create: `mall-frontend/src/pages/product/list.vue`

- [ ] **Step 1: Create directory and write the file**

```bash
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/pages/product
```

Write `mall-frontend/src/pages/product/list.vue`:

```vue
<template>
  <view class="page">
    <view v-if="loading && products.length === 0">
      <wd-skeleton :row="6" />
    </view>
    <view v-else-if="products.length === 0">
      <wd-empty description="暂无商品" />
    </view>
    <view v-else class="products-grid">
      <view
        v-for="p in products"
        :key="p.id"
        class="product-card"
        @tap="goDetail(p.id)"
      >
        <image
          class="product-img"
          :src="(p.images && p.images[0]) || '/static/placeholder.png'"
          mode="aspectFill"
        />
        <text class="product-name">{{ p.name }}</text>
        <text class="product-price">¥{{ (p.price / 100).toFixed(2) }}</text>
      </view>
    </view>

    <wd-load-more :status="loadMoreStatus" @loadmore="loadMore" />
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listProducts, searchProducts } from '@/api/product'
import { showError } from '@/api/request'
import type { ProductDetailResp } from '@/types/api'

const products = ref<ProductDetailResp[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const loadMoreStatus = ref<'loadmore' | 'loading' | 'nomore'>('loadmore')

let keyword = ''
let shopId = 0

function goDetail(id: number) {
  uni.navigateTo({ url: `/pages/product/detail?id=${id}` })
}

async function fetchPage(p: number) {
  loading.value = true
  try {
    let resp
    if (keyword) {
      resp = await searchProducts(keyword, p)
    } else {
      resp = await listProducts({ page: p, pageSize: 20, shopId: shopId || undefined })
    }
    products.value = p === 1 ? (resp.products ?? []) : [...products.value, ...(resp.products ?? [])]
    total.value = resp.total ?? 0
    loadMoreStatus.value = products.value.length >= total.value ? 'nomore' : 'loadmore'
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadMoreStatus.value !== 'loadmore' || loading.value) return
  loadMoreStatus.value = 'loading'
  page.value++
  await fetchPage(page.value)
}

onMounted(() => {
  const pages = getCurrentPages()
  const current = pages[pages.length - 1] as unknown as { options: { keyword?: string; shopId?: string } }
  keyword = current.options?.keyword ? decodeURIComponent(current.options.keyword) : ''
  shopId = Number(current.options?.shopId ?? 0)
  fetchPage(1)
})
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding: $space-sm;
  padding-bottom: $space-xl;
}

.products-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $space-sm;
}

.product-card {
  background: $color-bg-card;
  border-radius: $radius-md;
  overflow: hidden;
  border: 1px solid $color-border;
}

.product-img {
  width: 100%;
  height: 150px;
  background: $color-border;
}

.product-name {
  display: block;
  font-size: $font-size-base;
  color: $color-text-primary;
  padding: $space-xs $space-sm;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-price {
  display: block;
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-primary;
  padding: 0 $space-sm $space-sm;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/product/list.vue
git commit -m "feat(frontend): add Product List page (keyword/shopId/browse modes)"
```

---

## Task 10: Product Detail Page

**Files:**
- Create: `mall-frontend/src/pages/product/detail.vue`

- [ ] **Step 1: Write the file**

Write `mall-frontend/src/pages/product/detail.vue`:

```vue
<template>
  <view class="page">
    <view v-if="loading">
      <wd-skeleton :row="8" />
    </view>
    <view v-else-if="!product">
      <wd-empty description="商品不存在" />
    </view>
    <template v-else>
      <!-- Main image -->
      <image
        class="main-img"
        :src="(product.images && product.images[0]) || '/static/placeholder.png'"
        mode="aspectFill"
      />

      <!-- Product info -->
      <view class="info-card">
        <text class="product-price">¥{{ (product.price / 100).toFixed(2) }}</text>
        <text class="product-name">{{ product.name }}</text>
        <text class="product-stock">库存: {{ product.stock }} 件</text>
      </view>

      <!-- Description -->
      <view class="desc-card">
        <text class="desc-title">商品详情</text>
        <text class="desc-text">{{ product.description }}</text>
      </view>

      <!-- Sticky bottom bar -->
      <view class="bottom-bar">
        <wd-button type="primary" block @tap="addToCart">加入购物车</wd-button>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProductDetail } from '@/api/product'
import { showError } from '@/api/request'
import { useCartStore } from '@/stores/cart'
import type { ProductDetailResp } from '@/types/api'

const cartStore = useCartStore()
const product = ref<ProductDetailResp | null>(null)
const loading = ref(true)

function addToCart() {
  cartStore.increment()
  uni.showToast({ title: '已加入购物车', icon: 'success', duration: 1500 })
}

onMounted(async () => {
  const pages = getCurrentPages()
  const current = pages[pages.length - 1] as unknown as { options: { id?: string } }
  const id = Number(current.options?.id ?? 0)
  try {
    product.value = await getProductDetail(id)
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
})
</script>

<style lang="scss" scoped>
.page {
  background: $color-bg-page;
  min-height: 100vh;
  padding-bottom: 80px;
}

.main-img {
  width: 100%;
  height: 300px;
  background: $color-border;
}

.info-card {
  background: $color-bg-card;
  padding: $space-md;
  margin-top: $space-sm;
  display: flex;
  flex-direction: column;
  gap: $space-xs;
}

.product-price {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.product-name {
  font-size: $font-size-md;
  font-weight: $font-weight-medium;
  color: $color-text-primary;
  line-height: 1.4;
}

.product-stock {
  font-size: $font-size-sm;
  color: $color-text-hint;
}

.desc-card {
  background: $color-bg-card;
  padding: $space-md;
  margin-top: $space-sm;
  display: flex;
  flex-direction: column;
  gap: $space-sm;
}

.desc-title {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.desc-text {
  font-size: $font-size-base;
  color: $color-text-secondary;
  line-height: 1.6;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: $space-sm $space-md;
  background: $color-bg-card;
  border-top: 1px solid $color-border;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -10
```

Expected: exit 0.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/product/detail.vue
git commit -m "feat(frontend): add Product Detail page with add-to-cart action"
```

---

## Task 11: Login Placeholder + E2E Smoke Checklist

**Files:**
- Create: `mall-frontend/src/pages/login/index.vue`
- Modify: `docs/e2e-frontend-backend-checklist.md`

- [ ] **Step 1: Create the login placeholder shell**

```bash
mkdir -p /home/carter/workspace/go/yw-mall/mall-frontend/src/pages/login
```

Write `mall-frontend/src/pages/login/index.vue`:

```vue
<template>
  <view class="page">
    <wd-empty description="登录功能即将上线" image-type="search" />
  </view>
</template>

<script setup lang="ts">
</script>

<style lang="scss" scoped>
.page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: $color-bg-page;
}
</style>
```

- [ ] **Step 2: Append UI smoke checklist to the e2e doc**

Append the following section to `docs/e2e-frontend-backend-checklist.md`:

```markdown

---

## 7. Frontend UI Smoke (H5)

Start the dev server:
```bash
cd mall-frontend
pnpm dev:h5
# Opens at http://localhost:5173
```

- [ ] Dev server starts without errors; browser opens at `http://localhost:5173`
- [ ] **Home page:** Search bar visible; recommended shops horizontal strip loads (≥1 shop card); featured products grid loads (≥1 product card)
- [ ] **Home → search:** Type "手机" in search bar, press Enter → navigates to Product List with results
- [ ] **Home → Shop Detail:** Tap a shop card → Shop Detail page loads with banner, name, rating, follow button, product grid
- [ ] **Shop Detail follow:** Tap "关注" without login → redirects to login placeholder page
- [ ] **Shop List:** Navigate to `/pages/shop/list` directly → list of shops visible; scroll to bottom → more shops load
- [ ] **Product List (browse):** Navigate to `/pages/product/list` → 2-column product grid visible
- [ ] **Product Detail:** Tap a product card from home → detail page loads; price shown as ¥xx.xx; "加入购物车" button visible
- [ ] **Add to cart:** Tap "加入购物车" → toast "已加入购物车" appears
- [ ] **Build gate:** `pnpm run build:h5` exits 0 with zero TypeScript errors
```

- [ ] **Step 3: Verify final build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -15
```

Expected: exit 0, zero TypeScript errors, `dist/build/h5/` produced.

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/login/index.vue
git add docs/e2e-frontend-backend-checklist.md
git commit -m "feat(frontend): add login placeholder page and UI smoke checklist"
```

---

## Self-Review Notes

**Spec coverage:**
- §1 Scaffold + Wot Design Uni + Pinia → Tasks 1–4 ✓
- §3.3 Design tokens → Task 2 ✓
- §4.1 Base request + JWT + 401 → Task 3 ✓
- §4.2 Shop API (8 fns) → Task 5 ✓
- §4.3 Product API (3 fns) → Task 5 ✓
- §4.4 Response types → Task 3 ✓
- §5.1 User store (persisted) → Task 4 ✓
- §5.2 Cart store → Task 4 ✓
- §6.1 Home → Task 6 ✓
- §6.2 Shop List → Task 7 ✓
- §6.3 Shop Detail + follow → Task 8 ✓
- §6.4 Product List (keyword/shopId/browse) → Task 9 ✓
- §6.5 Product Detail + sticky add-to-cart → Task 10 ✓
- §8 Build gate + smoke checklist → Task 11 ✓
- §9 Login placeholder → Task 11 ✓

**Type consistency:** All types defined in `src/types/api.ts` (Task 3) and referenced consistently across Tasks 5–10. `ShopItem`, `ProductDetailResp`, `ShopListResp`, `ProductListResp`, `IsFollowingResp`, `OkResp` used exactly as defined.

**No placeholders:** All steps contain complete code blocks, exact commands, and expected output.
