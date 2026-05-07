# C-Side Frontend Foundation Design

**Sub-project 2 of the C-side frontend-backend system.**
Builds the uni-app H5 scaffold, design tokens, and the complete browsing domain
(Home, Shop List, Shop Detail, Product List, Product Detail).

---

## 1. Goals & Scope

**In scope:**
- uni-app H5 project scaffold with TypeScript, Vite, Wot Design Uni, Pinia
- Design token system (colors, spacing, typography)
- API layer with typed request wrapper and JWT injection
- 5 browsing-domain pages: Home, Shop List, Shop Detail, Product List, Product Detail
- Pinia stores: user (token/profile), cart (count badge)
- UI smoke checklist appended to the existing e2e checklist doc

**Out of scope:**
- Login / registration flows (auth pages are a separate sub-project)
- Cart, checkout, order, address pages
- WeChat Mini Program target (H5 only for this sub-project)
- Admin or merchant-facing UI

---

## 2. Architecture

### 2.1 Project Location

```
yw-mall/
  mall-frontend/        ← new directory (sibling to mall-api/, mall-user-rpc/, …)
```

### 2.2 Stack

| Concern | Choice | Version |
|---|---|---|
| Framework | uni-app + Vue 3 | latest @dcloudio |
| Language | TypeScript | ~5.x |
| Build | Vite | bundled with uni-app |
| UI components | Wot Design Uni | ^1.x |
| Component auto-import | @uni-helper/vite-plugin-uni-components | latest |
| State | Pinia | ^2.x |
| State persistence | pinia-plugin-persistedstate | ^3.x |
| CSS | SCSS + design tokens in `src/styles/tokens.scss` |

### 2.3 Directory Layout

```
mall-frontend/
  src/
    api/
      request.ts          # base uni.request wrapper; JWT header; 401 handler
      shop.ts             # all shop endpoints (8 functions)
      product.ts          # listProducts, getProductDetail, searchProducts
    pages/
      index/
        index.vue         # Home page
      shop/
        list.vue          # Shop List
        detail.vue        # Shop Detail
      product/
        list.vue          # Product List
        detail.vue        # Product Detail
      login/
        index.vue         # Placeholder redirect target (empty shell only)
    stores/
      user.ts             # token, userId, setToken(), clear() — persisted
      cart.ts             # count, increment(), decrement()
    styles/
      tokens.scss         # design tokens (imported globally via vite config)
    types/
      api.ts              # response shape types (ShopItem, ProductDetailResp, …)
  pages.json              # uni-app router: 5 browsing pages + login placeholder
  manifest.json           # H5 config, history router mode, app title "Mall"
  vite.config.ts          # H5 proxy /api → localhost:18888, auto-import plugins
  tsconfig.json
  package.json
```

---

## 3. Scaffold & Configuration

### 3.1 Router (pages.json)

History mode (not hash). Tab bar has 3 entries: Home, Cart (placeholder icon,
no page yet), Profile (placeholder). Shop and product pages are stack pages
(no tabbar).

```json
{
  "pages": [
    { "path": "pages/index/index", "style": { "navigationBarTitleText": "首页" } },
    { "path": "pages/shop/list", "style": { "navigationBarTitleText": "店铺列表" } },
    { "path": "pages/shop/detail", "style": { "navigationBarTitleText": "店铺详情" } },
    { "path": "pages/product/list", "style": { "navigationBarTitleText": "商品列表" } },
    { "path": "pages/product/detail", "style": { "navigationBarTitleText": "商品详情" } },
    { "path": "pages/login/index", "style": { "navigationBarTitleText": "登录" } }
  ],
  "tabBar": {
    "list": [
      { "pagePath": "pages/index/index", "text": "首页" },
      { "pagePath": "pages/index/index", "text": "购物车" },
      { "pagePath": "pages/index/index", "text": "我的" }
    ]
  },
  "globalStyle": {
    "navigationBarTextStyle": "black",
    "navigationBarTitleText": "Mall",
    "backgroundColor": "#F5F5F5"
  }
}
```

### 3.2 Vite Config (vite.config.ts)

```typescript
import { defineConfig } from 'vite'
import uni from '@dcloudio/vite-plugin-uni'
import Components from '@uni-helper/vite-plugin-uni-components'
import { WotResolver } from '@uni-helper/vite-plugin-uni-components/resolvers'

export default defineConfig({
  plugins: [
    Components({ resolvers: [WotResolver()] }),
    uni(),
  ],
  server: {
    proxy: {
      '/api': { target: 'http://localhost:18888', changeOrigin: true },
    },
  },
})
```

### 3.3 Design Tokens (src/styles/tokens.scss)

```scss
// Brand
$color-primary:   #FF4B4B;
$color-primary-light: #FF7070;

// Neutrals
$color-text-primary:   #1A1A1A;
$color-text-secondary: #666666;
$color-text-hint:      #999999;
$color-bg-page:        #F5F5F5;
$color-bg-card:        #FFFFFF;
$color-border:         #EEEEEE;

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

// Radius
$radius-sm:  4px;
$radius-md:  8px;
$radius-lg:  12px;
```

Tokens are imported globally via `vite.config.ts` `css.preprocessorOptions.scss.additionalData`.

---

## 4. API Layer

### 4.1 Base Request (src/api/request.ts)

- Wraps `uni.request` in a `Promise<T>`
- Reads `useUserStore().token` and injects `Authorization: Bearer <token>` when present
- On HTTP 401: calls `useUserStore().clear()` then `uni.reLaunch({ url: '/pages/login/index' })`
- On non-2xx: rejects with `{ code, message }` from response body
- Caller never needs try/catch for auth errors; other errors surface as rejected promises

### 4.2 Shop API (src/api/shop.ts)

Typed wrappers for all 8 shop endpoints from the backend:

```typescript
getShopDetail(id: number): Promise<ShopDetailResp>
listShops(page: number, pageSize: number): Promise<ShopListResp>
getRecommendedShops(limit: number): Promise<ShopListResp>
getShopProducts(shopId: number, page: number, pageSize: number): Promise<ProductListResp>
followShop(shopId: number): Promise<OkResp>
unfollowShop(shopId: number): Promise<OkResp>
isFollowing(shopId: number): Promise<IsFollowingResp>
listFollowedShops(page: number, pageSize: number): Promise<ShopListResp>
```

### 4.3 Product API (src/api/product.ts)

```typescript
listProducts(params: { page: number; pageSize: number; shopId?: number }): Promise<ProductListResp>
getProductDetail(id: number): Promise<ProductDetailResp>
searchProducts(keyword: string, page?: number): Promise<ProductListResp>
```

### 4.4 Response Types (src/types/api.ts)

Mirror the backend `types.go` shapes exactly:

```typescript
interface ShopItem {
  id: number; name: string; logo: string; banner: string
  rating: number; productCount: number; description: string
}
interface ProductDetailResp {
  id: number; name: string; description: string
  price: number; stock: number; images: string[]
  shopId: number; status: number
}
interface ShopDetailResp { shop: ShopItem }
interface ShopListResp { shops: ShopItem[]; total: number }
interface ProductListResp { products: ProductDetailResp[]; total: number }
interface IsFollowingResp { isFollowing: boolean }
interface OkResp { ok: boolean }
```

---

## 5. Pinia Stores

### 5.1 User Store (src/stores/user.ts)

```typescript
export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userId = ref(0)
  const setToken = (t: string, id: number) => { token.value = t; userId.value = id }
  const clear = () => { token.value = ''; userId.value = 0 }
  return { token, userId, setToken, clear }
}, { persist: true })
```

### 5.2 Cart Store (src/stores/cart.ts)

```typescript
export const useCartStore = defineStore('cart', () => {
  const count = ref(0)
  const increment = () => count.value++
  const decrement = () => { if (count.value > 0) count.value-- }
  return { count, increment, decrement }
})
```

Cart count is not persisted (fetched fresh from cart-rpc on demand in a future sub-project).

---

## 6. Pages

All pages use `<script setup lang="ts">` + Wot Design Uni components.
Loading states use `wd-skeleton`; empty states use `wd-empty`.
Errors surface via `wd-toast` with "请求失败，请重试".

### 6.1 Home (`pages/index/index.vue`)

- `wd-search` bar → on confirm navigate to `/pages/product/list?keyword=<q>`
- Recommended shops: horizontal `wd-swiper` or `scroll-view`; calls `getRecommendedShops(5)` on mount
- Featured products: 2-column grid; calls `listProducts({ page: 1, pageSize: 10 })`
- Tapping a shop → `/pages/shop/detail?id=<shopId>`
- Tapping a product → `/pages/product/detail?id=<productId>`

### 6.2 Shop List (`pages/shop/list.vue`)

- Vertical list of `ShopCard` (inline component: logo, name, rating, productCount)
- Infinite scroll via `wd-load-more` (loads next page on reach-bottom)
- Calls `listShops(page, 10)` incrementally

### 6.3 Shop Detail (`pages/shop/detail.vue`)

- Receives `id` from query params
- Header: banner image, shop name, rating, follow button
  - Follow button: if `isFollowing` → "已关注" (calls `unfollowShop`), else "关注" (calls `followShop`)
  - If no token: tapping follow redirects to login placeholder
- Products tab: calls `getShopProducts(id, page, 10)` with infinite scroll

### 6.4 Product List (`pages/product/list.vue`)

- Receives `keyword` or `shopId` as query params (one or neither)
- If `keyword`: calls `searchProducts(keyword, page)`
- If `shopId`: calls `listProducts({ shopId, page, pageSize: 10 })`
- Otherwise: calls `listProducts({ page, pageSize: 10 })`
- 2-column grid with infinite scroll

### 6.5 Product Detail (`pages/product/detail.vue`)

- Receives `id` from query params
- Calls `getProductDetail(id)` on mount
- Layout: large image (`wd-img`), name, price formatted as `¥${(price/100).toFixed(2)}`, description
- Sticky bottom bar: "加入购物车" button → calls `cartStore.increment()` + `wd-toast` "已加入购物车"

---

## 7. Error Handling

| Scenario | Behavior |
|---|---|
| Network error | `wd-toast` "网络错误，请重试" |
| 401 Unauthorized | `userStore.clear()` + redirect to login placeholder |
| 404 / resource not found | `wd-empty` empty state on page |
| Loading in progress | `wd-skeleton` placeholder |
| Empty list | `wd-empty` with descriptive message |

Error handling lives in `request.ts`; pages only handle the empty/not-found case.

---

## 8. Testing & Validation

**Build gate:** `pnpm run build:h5` must exit 0 with zero TypeScript errors.

**UI smoke checklist** (to be appended to `docs/e2e-frontend-backend-checklist.md`):

- [ ] Dev server starts: `pnpm dev:h5` → opens at `http://localhost:5173`
- [ ] Home loads: recommended shops strip visible, featured products grid visible
- [ ] Shop List: scrolling loads more shops
- [ ] Shop Detail: shop info visible; follow button visible (shows redirect if not logged in)
- [ ] Product List via keyword: search "手机" returns results
- [ ] Product List via shopId: tapping a shop → shop detail → products visible
- [ ] Product Detail: price displayed as ¥xx.xx; "加入购物车" shows toast
- [ ] `pnpm run build:h5` exits 0 with no TypeScript errors

---

## 9. Out-of-Scope Deferred Items

- Login page implementation (redirected to placeholder only)
- Cart page and checkout flow
- Order history, address management pages
- WeChat Mini Program build target
- Unit / component tests
- CI pipeline integration
