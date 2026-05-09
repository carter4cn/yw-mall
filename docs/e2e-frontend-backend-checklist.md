# C-Side Frontend-Backend E2E Checklist

Manual smoke-test checklist for verifying the C-side (consumer-facing) frontend-backend integration.
Run after `start.sh start` and `scripts/check-seed.sh`.

Base URL: `http://localhost:18888`

## 0. Preconditions

- [ ] `start.sh start` completed without critical errors
- [ ] `scripts/check-seed.sh` exits 0 (all checks pass)
- [ ] Have a test JWT: login as user `alice` / `alice123` and save token

```bash
TOKEN=$(curl -s -X POST http://localhost:18888/api/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"alice123"}' | jq -r .token)
echo $TOKEN
```

---

## 1. Shop Routes

### 1.1 Shop Detail
```bash
curl -s http://localhost:18888/api/shop/detail/1 | jq .
```
- [ ] Returns `shop` object with id, name, logo, banner, rating, productCount

### 1.2 Shop List
```bash
curl -s 'http://localhost:18888/api/shop/list?page=1&pageSize=10' | jq .
```
- [ ] Returns `shops` array with ≥6 items, `total` ≥6

### 1.3 Recommended Shops
```bash
curl -s 'http://localhost:18888/api/shop/recommended?limit=5' | jq .
```
- [ ] Returns `shops` array (ordered by rating desc), length ≤5

### 1.4 Shop Products
```bash
curl -s 'http://localhost:18888/api/shop/1/products?page=1&pageSize=10' | jq .
```
- [ ] Returns `products` array for shop 1, `total` > 0

### 1.5 Follow Shop (requires auth)
```bash
curl -s -X POST http://localhost:18888/api/shop/1/follow \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `{"ok":true}`

### 1.6 Is Following
```bash
curl -s http://localhost:18888/api/shop/1/following \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `{"following":true}`

### 1.7 List Followed Shops
```bash
curl -s 'http://localhost:18888/api/shop/followed?page=1&pageSize=10' \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns shop 1 in `shops` array

### 1.8 Unfollow Shop
```bash
curl -s -X DELETE http://localhost:18888/api/shop/1/follow \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `{"ok":true}`; re-check IsFollowing returns `false`

---

## 2. Address Routes

### 2.1 Add Address
```bash
ADDR_ID=$(curl -s -X POST http://localhost:18888/api/address/add \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"receiverName":"测试用户","phone":"13812341234","province":"北京市","city":"北京市","district":"朝阳区","detail":"朝阳区某街道1号","isDefault":true}' | jq -r .id)
echo $ADDR_ID
```
- [ ] Returns numeric `id` > 0

### 2.2 List Addresses
```bash
curl -s http://localhost:18888/api/address/list \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `addresses` array with the newly added address, `isDefault: true`

### 2.3 Get Address
```bash
curl -s http://localhost:18888/api/address/$ADDR_ID \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns address object with correct fields

### 2.4 Get Default Address
```bash
curl -s http://localhost:18888/api/address/default \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns the address marked `isDefault: true`

### 2.5 Update Address
```bash
curl -s -X PUT http://localhost:18888/api/address/update/$ADDR_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"receiverName":"更新用户","phone":"13900001111","city":"上海市","district":"浦东新区","detail":"浦东某路2号"}' | jq .
```
- [ ] Returns `{"ok":true}`

### 2.6 Set Default Address
```bash
curl -s -X POST http://localhost:18888/api/address/$ADDR_ID/default \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `{"ok":true}`

### 2.7 Delete Address
```bash
curl -s -X DELETE http://localhost:18888/api/address/delete/$ADDR_ID \
  -H "Authorization: Bearer $TOKEN" | jq .
```
- [ ] Returns `{"ok":true}`; re-check list shows address removed

---

## 3. Order with Address Snapshot

### 3.1 Create Order with Address
```bash
# Use the default address (from seed)
DEFAULT_ADDR=$(curl -s http://localhost:18888/api/address/default \
  -H "Authorization: Bearer $TOKEN" | jq -r .id)

ORDER=$(curl -s -X POST http://localhost:18888/api/order/create \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"items\":[{\"productId\":1,\"productName\":\"测试商品\",\"price\":9900,\"quantity\":2}],\"addressId\":$DEFAULT_ADDR}" | jq .)
echo $ORDER
ORDER_ID=$(echo $ORDER | jq -r .id)
```
- [ ] Returns order with `id`, `orderNo`, `totalAmount`

### 3.2 Order Detail includes receiver snapshot
```bash
curl -s http://localhost:18888/api/order/detail/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN" | jq '{receiverName,receiverPhone,receiverCity,receiverDetail}'
```
- [ ] Shows non-empty receiver fields snapshotted from address at order time

---

## 4. Product with Shop Linkage

### 4.1 Product List shows shop_id
```bash
curl -s 'http://localhost:18888/api/product/list?page=1&pageSize=5' | jq '.products[0]'
```
- [ ] Product object exists (shop_id available in list response from seed)

### 4.2 Product Search
```bash
curl -s 'http://localhost:18888/api/product/search?keyword=手机' | jq '.total'
```
- [ ] Returns total > 0

---

## 5. Error Cases

- [ ] GET `/api/shop/detail/99999` → non-200 (shop not found)
- [ ] GET `/api/address/99999` (with auth) → non-200 (not owner or not found)
- [ ] POST `/api/shop/1/follow` without token → 401
- [ ] POST `/api/address/add` without token → 401

---

## 6. Seed Verification

```bash
bash scripts/check-seed.sh
```
- [ ] All checks pass (exit 0)
- [ ] shops count ≥ 6
- [ ] products count ≥ 40
- [ ] user_addresses count ≥ 5
- [ ] orders count ≥ 5
- [ ] orders with address_id > 0 ≥ 5

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

---

## 8. Login & Auth (My Tab)

### 8.1 Unauthenticated State
- [ ] My tab shows empty state with "登录后查看个人信息"
- [ ] "去登录" button navigates to login page
- [ ] Login page shows username + password fields and disabled submit button when either field is empty

### 8.2 Login Flow
- [ ] Valid credentials → token stored, navigates back or to home
  - [ ] Login from My tab → navigates back (back stack exists)
  - [ ] Login from cold start → relaunches to home tab
- [ ] Invalid credentials → error toast shown
- [ ] Submit button disabled while request is in-flight

### 8.3 Authenticated State (My Tab)
- [ ] After login, My tab shows header with avatar initial (first letter of username)
- [ ] Username displayed in header
- [ ] "我的订单" row navigates to /pages/order/list
- [ ] "我的地址" row navigates to /pages/address/list
- [ ] "关注的店铺" row navigates to /pages/shop/followed
- [ ] After login from My tab, switching back shows authenticated state immediately (no refresh needed — validates onShow)

### 8.4 Logout
- [ ] "退出登录" clears auth state and relaunches to home
- [ ] My tab reverts to unauthenticated state after logout

---
