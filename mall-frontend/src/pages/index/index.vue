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
        <wd-status-tip status="empty" title="暂无商品" />
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
