<template>
  <view class="page">
    <view v-if="loading && shops.length === 0">
      <wd-skeleton :row="5" />
    </view>
    <view v-else-if="shops.length === 0">
      <wd-status-tip status="empty" title="暂无店铺" />
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
