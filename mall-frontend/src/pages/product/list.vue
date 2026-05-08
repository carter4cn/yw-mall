<template>
  <view class="page">
    <view v-if="loading && products.length === 0">
      <wd-skeleton :row="6" />
    </view>
    <view v-else-if="products.length === 0">
      <wd-status-tip status="empty" title="暂无商品" />
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

    <wd-loadmore :state="loadMoreStatus" />
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onReachBottom } from '@dcloudio/uni-app'
import { listProducts, searchProducts } from '@/api/product'
import { showError } from '@/api/request'
import type { ProductDetailResp } from '@/types/api'

const products = ref<ProductDetailResp[]>([])
const page = ref(1)
const total = ref(0)
const loading = ref(false)
const loadMoreStatus = ref<'loading' | 'finished'>('loading')

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
    loadMoreStatus.value = products.value.length >= total.value ? 'finished' : 'loading'
  } catch (err) {
    showError(err)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (loadMoreStatus.value === 'finished' || loading.value) return
  page.value++
  await fetchPage(page.value)
}

onReachBottom(() => loadMore())

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
