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
