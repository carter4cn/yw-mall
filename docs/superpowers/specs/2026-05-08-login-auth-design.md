# Login + Auth (My Tab) Design

## Goal

Replace the login placeholder with a working username/password form wired to the existing `/api/user/login` endpoint, and build the "我的" tab with an auth-aware profile view.

## Architecture

No backend changes. Frontend-only: 8 new/updated files.

| File | Change |
|---|---|
| `mall-frontend/src/pages/login/index.vue` | Replace placeholder with login form |
| `mall-frontend/src/pages/my/index.vue` | New — auth-aware My tab page |
| `mall-frontend/src/pages/order/list.vue` | New — placeholder (功能即将上线) |
| `mall-frontend/src/pages/address/list.vue` | New — placeholder (功能即将上线) |
| `mall-frontend/src/pages/shop/followed.vue` | New — placeholder (功能即将上线) |
| `mall-frontend/src/api/user.ts` | New — `login()`, `getUserInfo()` |
| `mall-frontend/src/types/api.ts` | Add `UserInfoResp` |
| `mall-frontend/src/pages.json` | Register 4 new pages, wire tabBar tab 3 |

## API

**Existing endpoints used:**

`POST /api/user/login` — no auth header required
```json
// request
{ "username": "alice", "password": "alice123" }
// response
{ "id": 1, "token": "<jwt>" }
```

`GET /api/user/info` — requires `Authorization: Bearer <token>`
```json
// response
{ "id": 1, "username": "alice", "phone": "13800000001", "avatar": "", "createTime": 1714500000 }
```

**New API module** (`src/api/user.ts`):
```typescript
export function login(username: string, password: string) {
  return request<{ id: number; token: string }>({
    url: '/api/user/login',
    method: 'POST',
    data: { username, password },
  })
}

export function getUserInfo() {
  return request<UserInfoResp>({ url: '/api/user/info', method: 'GET' })
}
```

**New type** (`src/types/api.ts`):
```typescript
export interface UserInfoResp {
  id: number
  username: string
  phone: string
  avatar: string
  createTime: number
}
```

## Login Page

`src/pages/login/index.vue` — replaces the existing wd-empty placeholder.

Layout (centered column, full height):
1. App icon / logo at top
2. `wd-input` — username, type text, placeholder "用户名"
3. `wd-input` — password, type password, placeholder "密码"
4. `wd-button` type="primary" block — "登录", disabled + loading while request is in-flight

**Success path:** `userStore.setToken(resp.token, resp.id)` → if `getCurrentPages().length > 1` then `uni.navigateBack()`, else `uni.reLaunch({ url: '/pages/index/index' })`.

**Error path:** caught by `request.ts` `showError()` → `uni.showToast({ icon: 'error', title: message })`. No inline error text.

No registration link (no backend endpoint exists).

## My Tab Page

`src/pages/my/index.vue` — registered as tabBar tab 3 (`pages/my/index`).

**Not logged in state** (checked in `onShow` via `userStore.token`):
- `wd-empty` image-type="person", description="登录后查看个人信息"
- `wd-button` → `uni.navigateTo({ url: '/pages/login/index' })`

**Logged in state:**
- Grey header band: avatar circle (fallback initials from username[0]) + username
- User info fetched via `getUserInfo()` on `onShow` (always fresh, not persisted)
- `wd-cell-group` with 3 rows:
  - 我的订单 → `uni.navigateTo({ url: '/pages/order/list' })`
  - 我的地址 → `uni.navigateTo({ url: '/pages/address/list' })`
  - 关注的店铺 → `uni.navigateTo({ url: '/pages/shop/followed' })`
- `wd-button` type="ghost" block "退出登录" → `userStore.clear()` + `uni.reLaunch({ url: '/pages/index/index' })`

`onShow` re-evaluates token on every tab focus — covers logout + login from another page.

## Placeholder Pages

`src/pages/order/list.vue`, `src/pages/address/list.vue`, and `src/pages/shop/followed.vue` — same pattern as login placeholder:
- `wd-empty` description="功能即将上线"
- Centered full-height view using `$color-bg-page` background token

## pages.json Changes

Register 4 new pages:
```json
{ "path": "pages/my/index",       "style": { "navigationBarTitleText": "我的" } },
{ "path": "pages/order/list",     "style": { "navigationBarTitleText": "我的订单" } },
{ "path": "pages/address/list",   "style": { "navigationBarTitleText": "我的地址" } },
{ "path": "pages/shop/followed",  "style": { "navigationBarTitleText": "关注的店铺" } }
```

Update tabBar tab 3:
```json
{ "pagePath": "pages/my/index", "text": "我的" }
```

## Error Handling

| Scenario | Handled by |
|---|---|
| Wrong credentials | `request.ts` `showError()` → toast |
| Network error | Same |
| Token expired mid-session | `request.ts` 401 handler → `userStore.clear()` + reLaunch login |
| My tab while logged out | `onShow` token check → shows login prompt |
| Logout | `userStore.clear()` + `reLaunch` home; JWT is stateless, no server invalidation |

## Smoke Tests

Append to `docs/e2e-frontend-backend-checklist.md` §7:

- Login with alice / alice123 → "我的" tab shows username
- Logout → "我的" tab shows login prompt + "去登录" button
- Tap "去登录" → login page loads; login → returns to previous page (or home)
- Tap "我的订单" → placeholder page loads
- Tap "我的地址" → placeholder page loads
- Tap "关注的店铺" → shop list page loads
- `pnpm run build:h5` exits 0 with no TypeScript errors
