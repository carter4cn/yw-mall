# Login + Auth (My Tab) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a working login form and auth-aware "我的" tab, wired to the existing `/api/user/login` and `/api/user/info` backend endpoints.

**Architecture:** Frontend-only — no backend changes. `src/api/user.ts` wraps the two user endpoints. The login page stores the JWT in Pinia's `userStore` and navigates back or home. The My tab calls `onShow()` on every focus to check `userStore.token` and fetch fresh user info. Three placeholder pages (orders, addresses, followed shops) make the nav rows functional.

**Tech Stack:** uni-app Vue 3 + TypeScript, Wot Design Uni (wd-input, wd-button, wd-cell-group, wd-cell, wd-empty), Pinia (`useUserStore` — already wired with `persist: true`).

---

## File Map

| File | Action |
|---|---|
| `mall-frontend/src/types/api.ts` | Add `UserInfoResp` interface |
| `mall-frontend/src/api/user.ts` | Create: `login()`, `getUserInfo()` |
| `mall-frontend/src/pages.json` | Add 4 page entries; update tabBar tab 3 |
| `mall-frontend/src/pages/my/index.vue` | Create: auth-aware My tab page |
| `mall-frontend/src/pages/order/list.vue` | Create: placeholder |
| `mall-frontend/src/pages/address/list.vue` | Create: placeholder |
| `mall-frontend/src/pages/shop/followed.vue` | Create: placeholder |
| `mall-frontend/src/pages/login/index.vue` | Replace placeholder with login form |
| `docs/e2e-frontend-backend-checklist.md` | Append login/auth smoke tests to §7 |

---

### Task 1: API types + user.ts

**Files:**
- Modify: `mall-frontend/src/types/api.ts`
- Create: `mall-frontend/src/api/user.ts`

- [ ] **Step 1: Add `UserInfoResp` to types/api.ts**

Append to the end of `mall-frontend/src/types/api.ts` (after the existing `ApiError` interface):

```typescript
export interface UserInfoResp {
  id: number
  username: string
  phone: string
  avatar: string
  createTime: number
}
```

After the edit the file ends with:
```typescript
export interface ApiError {
  code: number
  message: string
}

export interface UserInfoResp {
  id: number
  username: string
  phone: string
  avatar: string
  createTime: number
}
```

- [ ] **Step 2: Create `mall-frontend/src/api/user.ts`**

```typescript
import { request } from './request'
import type { UserInfoResp } from '@/types/api'

export function login(username: string, password: string) {
  return request<{ id: number; token: string }>({
    url: '/api/user/login',
    method: 'POST',
    data: { username, password },
  })
}

export function getUserInfo() {
  return request<UserInfoResp>({ url: '/api/user/info', auth: true })
}
```

- [ ] **Step 3: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -8
```

Expected: exit 0, no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/types/api.ts mall-frontend/src/api/user.ts
git commit -m "feat(frontend): add UserInfoResp type and user API module"
```

---

### Task 2: pages.json + placeholder stubs

**Files:**
- Modify: `mall-frontend/src/pages.json`
- Create: `mall-frontend/src/pages/my/index.vue` (stub, replaced in Task 4)
- Create: `mall-frontend/src/pages/order/list.vue`
- Create: `mall-frontend/src/pages/address/list.vue`
- Create: `mall-frontend/src/pages/shop/followed.vue`

All four new pages must be created before updating pages.json, otherwise the build will fail due to missing files.

- [ ] **Step 1: Create `mall-frontend/src/pages/my/index.vue` (stub)**

```vue
<template>
  <view class="page">
    <wd-empty description="个人中心即将上线" image-type="person" />
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

- [ ] **Step 2: Create `mall-frontend/src/pages/order/list.vue`**

```vue
<template>
  <view class="page">
    <wd-empty description="功能即将上线" image-type="search" />
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

- [ ] **Step 3: Create `mall-frontend/src/pages/address/list.vue`**

```vue
<template>
  <view class="page">
    <wd-empty description="功能即将上线" image-type="search" />
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

- [ ] **Step 4: Create `mall-frontend/src/pages/shop/followed.vue`**

```vue
<template>
  <view class="page">
    <wd-empty description="功能即将上线" image-type="search" />
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

- [ ] **Step 5: Write `mall-frontend/src/pages.json`**

Replace the entire file with:

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
    },
    {
      "path": "pages/my/index",
      "style": { "navigationBarTitleText": "我的" }
    },
    {
      "path": "pages/order/list",
      "style": { "navigationBarTitleText": "我的订单" }
    },
    {
      "path": "pages/address/list",
      "style": { "navigationBarTitleText": "我的地址" }
    },
    {
      "path": "pages/shop/followed",
      "style": { "navigationBarTitleText": "关注的店铺" }
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
        "pagePath": "pages/my/index",
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

- [ ] **Step 6: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -8
```

Expected: exit 0.

- [ ] **Step 7: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages.json \
  mall-frontend/src/pages/my/index.vue \
  mall-frontend/src/pages/order/list.vue \
  mall-frontend/src/pages/address/list.vue \
  mall-frontend/src/pages/shop/followed.vue
git commit -m "feat(frontend): register new pages and wire My tab in tabBar"
```

---

### Task 3: Login form page

**Files:**
- Modify: `mall-frontend/src/pages/login/index.vue`

The current file contains a `wd-empty` placeholder. Replace the entire file.

- [ ] **Step 1: Write `mall-frontend/src/pages/login/index.vue`**

```vue
<template>
  <view class="page">
    <view class="logo-area">
      <wd-icon name="shop" size="64px" :color="'#FF4B4B'" />
      <text class="app-name">Mall</text>
    </view>

    <view class="form">
      <wd-input
        v-model="username"
        placeholder="用户名"
        clearable
        class="field"
      />
      <wd-input
        v-model="password"
        placeholder="密码"
        type="password"
        clearable
        class="field"
      />
      <wd-button
        type="primary"
        block
        :loading="loading"
        :disabled="loading || !username || !password"
        class="submit-btn"
        @click="handleLogin"
      >
        登录
      </wd-button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { login } from '@/api/user'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const username = ref('')
const password = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const res = await login(username.value, password.value)
    userStore.setToken(res.token, res.id)
    const pages = getCurrentPages()
    if (pages.length > 1) {
      uni.navigateBack()
    } else {
      uni.reLaunch({ url: '/pages/index/index' })
    }
  } catch {
    // error toast shown by request.ts showError()
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  background: $color-bg-page;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0 $spacing-lg;
}

.logo-area {
  margin-top: 100px;
  margin-bottom: 48px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
}

.app-name {
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  color: $color-text-primary;
}

.form {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.submit-btn {
  margin-top: $spacing-sm;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -8
```

Expected: exit 0, no TypeScript errors.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/login/index.vue
git commit -m "feat(frontend): implement login form with redirect-back on success"
```

---

### Task 4: My tab page

**Files:**
- Modify: `mall-frontend/src/pages/my/index.vue` (replace the stub from Task 2)

- [ ] **Step 1: Write `mall-frontend/src/pages/my/index.vue`**

```vue
<template>
  <view class="page">
    <!-- Not logged in -->
    <view v-if="!userStore.token" class="not-logged-in">
      <wd-empty image-type="person" description="登录后查看个人信息" />
      <wd-button type="primary" class="login-btn" @click="goLogin">去登录</wd-button>
    </view>

    <!-- Logged in -->
    <view v-else class="logged-in">
      <view class="header">
        <view class="avatar">{{ avatarLetter }}</view>
        <text class="username">{{ userInfo?.username ?? '' }}</text>
      </view>

      <wd-cell-group class="nav-group">
        <wd-cell title="我的订单" is-link @click="nav('/pages/order/list')" />
        <wd-cell title="我的地址" is-link @click="nav('/pages/address/list')" />
        <wd-cell title="关注的店铺" is-link @click="nav('/pages/shop/followed')" />
      </wd-cell-group>

      <view class="logout-area">
        <wd-button type="warning" block @click="handleLogout">退出登录</wd-button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { onShow } from '@dcloudio/uni-app'
import { getUserInfo } from '@/api/user'
import { useUserStore } from '@/stores/user'
import type { UserInfoResp } from '@/types/api'

const userStore = useUserStore()
const userInfo = ref<UserInfoResp | null>(null)

const avatarLetter = computed(() => {
  const name = userInfo.value?.username
  return name ? name[0].toUpperCase() : '?'
})

onShow(async () => {
  if (!userStore.token) {
    userInfo.value = null
    return
  }
  try {
    userInfo.value = await getUserInfo()
  } catch {
    // 401 → request.ts clears store and reLaunches login
  }
})

function nav(url: string) {
  uni.navigateTo({ url })
}

function goLogin() {
  uni.navigateTo({ url: '/pages/login/index' })
}

function handleLogout() {
  userStore.clear()
  uni.reLaunch({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.page {
  min-height: 100vh;
  background: $color-bg-page;
}

.not-logged-in {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: $spacing-xl;
}

.login-btn {
  margin-top: $spacing-lg;
  width: 240rpx;
}

.header {
  background: $color-primary;
  padding: $spacing-xl $spacing-lg;
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.avatar {
  width: 120rpx;
  height: 120rpx;
  border-radius: $border-radius-full;
  background: rgba(255, 255, 255, 0.3);
  color: #fff;
  font-size: $font-size-xl;
  font-weight: $font-weight-bold;
  display: flex;
  align-items: center;
  justify-content: center;
}

.username {
  color: #fff;
  font-size: $font-size-lg;
  font-weight: $font-weight-medium;
}

.nav-group {
  margin-top: $spacing-md;
}

.logout-area {
  padding: $spacing-xl $spacing-lg;
}
</style>
```

- [ ] **Step 2: Verify build passes**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -8
```

Expected: exit 0, no TypeScript errors.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add mall-frontend/src/pages/my/index.vue
git commit -m "feat(frontend): implement My tab with auth-aware profile view and logout"
```

---

### Task 5: Smoke checklist + final build

**Files:**
- Modify: `docs/e2e-frontend-backend-checklist.md`

- [ ] **Step 1: Append login/auth smoke tests to §7**

The file `docs/e2e-frontend-backend-checklist.md` currently ends with a §7 section that has a bullet list. Append these items to that bullet list (after the "Build gate" item):

```markdown
- [ ] **Login page:** Navigate to `http://localhost:5173/pages/login/index` → username + password fields visible; 登录 button disabled when fields are empty
- [ ] **Login with alice/alice123:** Fill in credentials → tap 登录 → navigates to home (or back if opened from another page)
- [ ] **My tab (logged in):** Tap "我的" tab → header shows username; rows for 我的订单, 我的地址, 关注的店铺 visible; 退出登录 button visible
- [ ] **My tab nav rows:** Tap each row → respective placeholder page loads with "功能即将上线"
- [ ] **Logout:** Tap 退出登录 → My tab switches to not-logged-in state with 去登录 button
- [ ] **My tab (not logged in):** Tap 去登录 → login page loads
- [ ] **Follow redirect:** On Shop Detail page tap 关注 without login → login page opens; login → navigates back to Shop Detail
```

- [ ] **Step 2: Final build verification**

```bash
cd /home/carter/workspace/go/yw-mall/mall-frontend
pnpm run build:h5 2>&1 | tail -8
```

Expected: exit 0, no TypeScript errors, `dist/build/h5/` produced.

- [ ] **Step 3: Commit**

```bash
cd /home/carter/workspace/go/yw-mall
git add docs/e2e-frontend-backend-checklist.md
git commit -m "docs: add login/auth smoke tests to e2e checklist"
```

---

## Self-Review Notes

**Spec coverage:**
- Login form + username/password inputs → Task 3 ✓
- `POST /api/user/login` → `userStore.setToken` → navigate back or home → Task 3 ✓
- `GET /api/user/info` fresh on `onShow` → Task 4 ✓
- My tab not-logged-in state (wd-empty + 去登录) → Task 4 ✓
- My tab logged-in state (header + 3 nav rows + logout) → Task 4 ✓
- 3 placeholder pages (order/list, address/list, shop/followed) → Task 2 ✓
- tabBar tab 3 wired to pages/my/index → Task 2 ✓
- 401 handling: already in request.ts (no new code needed) ✓
- Smoke tests appended to §7 → Task 5 ✓
